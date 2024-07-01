package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

type ChatHistory struct {
	FunctionNames []string
	History       []openai.ChatCompletionMessage
}

func extractCodeInFunctionByStr(codeStr string) (error, []string) {
	functionList := make([]string, 0)

	fset := token.NewFileSet()
	var node *ast.File
	var err error

	node, err = parser.ParseFile(fset, "", codeStr, parser.ParseComments)

	if err != nil {
		return err, nil
	}

	ast.Inspect(node, func(n ast.Node) bool {

		fn, ok := n.(*ast.FuncDecl)
		if ok {

			funcStart := fset.Position(fn.Pos()).Offset
			funcEnd := fset.Position(fn.End()).Offset

			content := []byte(codeStr)

			functionList = append(functionList, strings.TrimSpace(string(content[funcStart:funcEnd])))
		}
		return true
	})

	return err, functionList
}

func removeCommentsInFunction(fn *ast.FuncDecl) {
	// 重置函数的文档注释
	fn.Doc = nil
	// 遍历并修改所有节点，移除注释
	ast.Inspect(fn, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CommentGroup:
			*x = ast.CommentGroup{} // 移除注释组
		}
		return true
	})
}

func formatFunction(fn *ast.FuncDecl, fset *token.FileSet) string {
	var sb strings.Builder
	printer.Fprint(&sb, fset, fn) // 使用 printer 包来格式化 AST 节点
	return sb.String()
}


// level 1 means with only source code with no comment
func extractFunctionLevel_1(filename string) []string {
	functionList := []string{}
	fset := token.NewFileSet()
	// 解析源文件
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// 遍历 AST 中的每个节点
	ast.Inspect(file, func(n ast.Node) bool {
		// 检查节点是否为函数声明
		if fn, ok := n.(*ast.FuncDecl); ok {
			// 移除函数中的所有注释
			removeCommentsInFunction(fn)
			// 打印没有注释的函数
			functionList = append(functionList, formatFunction(fn, fset))
		}
		return true 
	})
	return functionList
}

func extractFunctionLevel_2(filename string) []string {
	functionList := make([]string, 0)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {

		fn, ok := n.(*ast.FuncDecl)
		if ok {

			funcStart := fset.Position(fn.Pos()).Offset
			funcEnd := fset.Position(fn.End()).Offset
			docString := fn.Doc.Text()
			var content []byte
			content, err = os.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			funcCode := strings.TrimSpace(string(content[funcStart:funcEnd]))

			funcStr := docString + funcCode
			functionList = append(functionList, funcStr)
		}
		return true
	})
	return functionList
}

func chat(client *openai.Client, prompt string, messages []openai.ChatCompletionMessage) string {
	var promptMessages []openai.ChatCompletionMessage
	if prompt != "" {
		promptMessages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
	} else {
		promptMessages = messages
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo0125,
			Messages: promptMessages,
		},
	)
	if err != nil {
		panic(err)
	}
	return resp.Choices[0].Message.Content
}

func extractCodeByRegex(completion string) string {
	code := ""
	re := regexp.MustCompile("(?s)```.*?\n(.*?)\n```")
	matches := re.FindAllStringSubmatch(completion, -1)
	for k, match := range matches {
		code = code + match[1]
		if k < len(matches)-1 {
			code = code + "\n\n"
		}
	}
	return code
}

func extractFuntionName(str string) []string {
	var functionNames []string

	re := regexp.MustCompile(`func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?\s+(\w+)`)
	matches := re.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		functionNames = append(functionNames, match[1])
	}
	return functionNames
}

func extractSourceFuntionName(filename string) ([]string, error) {
	fset := token.NewFileSet() // positions are relative to fset

	// 解析文件
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcNames []string
	// 使用 Walk 来遍历所有的节点
	ast.Inspect(node, func(n ast.Node) bool {
		// 找到函数声明节点
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				// 获取接收者的类型
				var recvName string
				if star, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
					// 指针接收者
					recvName = fmt.Sprintf("*%s", star.X.(*ast.Ident).Name)
				} else {
					// 非指针接收者
					recvName = fmt.Sprintf("%s", fn.Recv.List[0].Type.(*ast.Ident).Name)
				}
				// 组合接收者和函数名
				funcNames = append(funcNames, fmt.Sprintf("%s.%s", recvName, fn.Name.Name))
			} else {
				// 没有接收者，直接添加函数名
				funcNames = append(funcNames, fn.Name.Name)
			}
		}
		return true // 继续遍历
	})

	return funcNames, nil
}

func generateTest(client *openai.Client, sourceCodeList []string, basePrompt string, generatedTestFile string, historyFile string, workers int, funcNamesFile string) {
	total := len(sourceCodeList)
	file, err := os.Create(generatedTestFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	chunks := total / workers

	var allHistories []ChatHistory
	var funcNames [][]string
	for i := 0; i < workers; i++ {
		start := i * chunks
		end := start + chunks
		if i == workers-1 {
			end = total
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			histories := []ChatHistory{}

			for k := start; k < end; k++ {
				prompt := basePrompt + "\n" + sourceCodeList[k]
				completion := chat(client, prompt, nil)

				history := []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: completion,
					},
				}

				testFunctionStr := extractedGeneratedCode(completion)

				functionNameList := extractFuntionName(testFunctionStr)
				chatH := ChatHistory{
					FunctionNames: functionNameList,
					History:       history,
				}
				histories = append(histories, chatH)

				mutex.Lock()
				funcNames = append(funcNames, functionNameList)
				_, err = file.WriteString(fmt.Sprintf("%s\n\n", testFunctionStr))
				mutex.Unlock()
				if err != nil {
					panic(err)
				}
			}
			mutex.Lock()
			allHistories = append(allHistories, histories...)
			mutex.Unlock()
		}(start, end)
	}

	wg.Wait() // 等待所有goroutine完成
	fmt.Println("All goroutines completed.")
	// 保存所有历史记录到文件
	err = saveSliceToFile(allHistories, historyFile)
	if err != nil {
		panic(err)
	}

	saveFuncNamesToFile(funcNamesFile, funcNames)
}

func extractedGeneratedCode(completion string) string {
	generatedCode := completion
	if strings.Contains(completion, "```") {
		generatedCode = extractCodeByRegex(completion)
	}

	var testFunctionList []string
	if strings.Contains(generatedCode, "package") {
		err, extractedFunctionList := extractCodeInFunctionByStr(generatedCode)
		if err != nil {
			generatedCode = "///warning///\n" + completion
			fmt.Println("failed to parse generated code")
			testFunctionList = append(testFunctionList, generatedCode)
		} else {
			testFunctionList = extractedFunctionList
		}
	} else {
		testFunctionList = append(testFunctionList, generatedCode)
	}

	testFunctionStr := strings.Join(testFunctionList, "\n\n")
	return testFunctionStr
}

func repairCompilation(client *openai.Client, historyFile string, errorFile string, saveHistroyFile string, saveCompletionFile string, baseprompt string, workers int, funcNamesFile string) {
	histories, err := loadSliceFromFile(historyFile)
	if err != nil {
		panic(err)
	}
	errors := integration(errorFile, funcNamesFile)

	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, workers)

	file, err := os.Create(saveCompletionFile)

	if err != nil {
		panic(err)
	}

	for targetName, errorMsg := range errors {
		wg.Add(1)
		sem <- struct{}{}
		fmt.Println(targetName)
		go func(targetName, errorMsg string) {
			defer wg.Done()
			defer func() { <-sem }()

			prompt := baseprompt + errorMsg

			for k, chatHistory := range histories {
				functionNames := chatHistory.FunctionNames
				flag := false
				for _, functionName := range functionNames {
					if functionName == targetName {
						flag = true
					}
				}
				if flag {
					mutex.Lock()
					histories[k].History = append(histories[k].History, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					})
					mutex.Unlock()
					completion := chat(client, "", histories[k].History)
					mutex.Lock()
					histories[k].History = append(histories[k].History, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: completion,
					})
					mutex.Unlock()

				}
			}
		}(targetName, errorMsg)
	}

	wg.Wait()
	fmt.Println("All goroutines completed.")

	err = saveSliceToFile(histories, saveHistroyFile)
	if err != nil {
		panic(err)
	}
	funcNamesList := [][]string{}

	for _, history := range histories {
		chatMsgs := history.History
		lastMsg := chatMsgs[len(chatMsgs)-1]
		extractedFunctions := extractedGeneratedCode(lastMsg.Content)
		funcNames := extractFuntionName(extractedFunctions)
		history.FunctionNames = funcNames
		funcNamesList = append(funcNamesList, funcNames)
		_, err := file.WriteString(extractedFunctions + "\n")
		if err != nil {
			panic(err)
		}
	}
	saveFuncNamesToFile(funcNamesFile, funcNamesList)
}

func repair_failing(){

}

func parseJsonFile(filename string) map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var result map[string]string

	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		panic(err)
	}

	return result
}

const structType = "// argsort is a helper that implements sort.Interface, as used by// Argsort and ArgsortStable.\ntype argsort struct {\ns []float64 \n inds []int\n}"

func GetBaseFunctionDoc() map[string]string {
	baseFunctionMap := make(map[string]string)
	filepath := "/Users/maike/Desktop/gonum/internal/asm/f64/stubs_amd64.go"

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {

		fn, ok := n.(*ast.FuncDecl)
		if ok {
			docString := fn.Doc.Text()
			name := fn.Name.Name
			baseFunctionMap[name] = docString
		}
		return true
	})
	return baseFunctionMap
}

func ExtractFunctionLevel_3(filename string, baseFunctionMap map[string]string) []string {
	functionList := make([]string, 0)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		fmt.Println(err)
	}

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		funcStart := fset.Position(funcDecl.Pos()).Offset
		funcEnd := fset.Position(funcDecl.End()).Offset
		docString := funcDecl.Doc.Text()
		var content []byte
		content, err = os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		funcCode := strings.TrimSpace(string(content[funcStart:funcEnd]))

		funcStr := docString + funcCode

		functionName := funcDecl.Name.Name
		if functionName == "Len" || functionName == "Less" || functionName == "Swap" {
			funcStr += structType
		}

		baseFunctionDocList := make([]string, 0)

		if funcDecl.Body != nil {
			for _, stmt := range funcDecl.Body.List {
				baseFunctionDoc := inspectStmt(stmt, baseFunctionMap)
				if baseFunctionDoc != "" {
					baseFunctionDocList = append(baseFunctionDocList, baseFunctionDoc)
				}
			}
		}

		funcStr += strings.Join(baseFunctionDocList, "\n")
		functionList = append(functionList, funcStr)
	}
	// fmt.Println(functionList[1])
	return functionList
}

func inspectStmt(stmt ast.Stmt, baseFunctionMap map[string]string) string {
	var docstring string = ""
	ast.Inspect(stmt, func(n ast.Node) bool {

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			methodName := selExpr.Sel.Name
			if docs, ok := baseFunctionMap[methodName]; ok {
				docstring = docs
			}
		}

		return true
	})
	return docstring
}

func saveSliceToFile(slice []ChatHistory, filename string) error {
	if _, err := os.Stat(filename); err == nil {
		err := os.Remove(filename)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(slice); err != nil {
		return err
	}

	return nil
}

func loadSliceFromFile(filename string) ([]ChatHistory, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var slice []ChatHistory
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&slice); err != nil {
		return nil, err
	}

	return slice, nil
}

type NamesStruct struct {
	Names [][]string `json:"names"`
}

func saveFuncNamesToFile(filename string, funcNames [][]string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	names := NamesStruct{
		Names: funcNames,
	}
	err = encoder.Encode(names)
	if err != nil {
		panic(err)
	}
}

func integration(errorFilePath, funcNameFilePath string) map[string]string {
	efile, err := os.Open(errorFilePath)
	if err != nil {
		panic(err)
	}

	defer efile.Close()

	bytes, err := io.ReadAll(efile)
	if err != nil {
		panic(err)
	}

	errorJSON := map[string]string{}
	if err := json.Unmarshal(bytes, &errorJSON); err != nil {
		panic(err)
	}

	ffile, err := os.Open(funcNameFilePath)
	if err != nil {
		panic(err)
	}
	defer ffile.Close()

	bytes, err = io.ReadAll(ffile)
	if err != nil {
		panic(err)
	}

	names := NamesStruct{}
	if err := json.Unmarshal(bytes, &names); err != nil {
		panic(err)
	}

	funcNeedFmt := make(map[string]bool)
	keyToRemove := make([]string, 0)

	for funcName, errMsg := range errorJSON {
		for _, funcNameList := range names.Names {
			if len(funcNameList) == 1 {
				continue
			}
			for index, name := range funcNameList {
				if name == funcName && index > 0 {
					targetFuncName := funcNameList[0]
					if _, exists := errorJSON[targetFuncName]; !exists {
						errorJSON[targetFuncName] = fmt.Sprintf("%s: %s", funcName, errMsg)
					} else {
						funcNeedFmt[targetFuncName] = true
						errorJSON[targetFuncName] += fmt.Sprintf("%s: %s", funcName, errMsg)
					}
					keyToRemove = append(keyToRemove, funcName)
					break
				}
			}
		}
	}

	for _, key := range keyToRemove {
		delete(errorJSON, key)
	}

	for f := range funcNeedFmt {
		errorJSON[f] = fmt.Sprintf("%s: %s", f, errorJSON[f])
	}

	for k := range errorJSON {
		errorJSON[k] = errorJSON[k][:len(errorJSON[k])-1]
	}
	return errorJSON
}
