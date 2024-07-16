package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"

	boltpackageInfo "llm-test-generation/package_Info/boltdb"
	fastjsonPackageInfo "llm-test-generation/package_Info/fastjson"

	openai "github.com/sashabaranov/go-openai"
)

type ChatHistory struct {
	Duplicated    bool
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
func extractFunctionLevel_1(filename string, repo string) map[string]string {
	functionList := map[string]string{}
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
			if repo == "boltdb" {
				if fn.Name.Name == "Error" {
					return true
				}
				if !unicode.IsUpper(rune(fn.Name.Name[0])) {
					return true
				}
			}
			// 移除函数中的所有注释
			removeCommentsInFunction(fn)
			// 打印没有注释的函数
			funcStr := formatFunction(fn, fset)
			p := basePrompt
			p = strings.Replace(p, "{functionName}", fn.Name.Name, 1)
			funcStr = p + funcStr

			functionList[fn.Name.Name] = funcStr
		}
		return true
	})
	return functionList
}

func extractFunctionLevel_2(filename string, typeFile string, repo string) map[string]string {
	functionList_1 := extractFunctionLevel_1(filename, repo)
	file, err := os.Open(typeFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	functionList := map[string]string{}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	var typeMap map[string][]string

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&typeMap); err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			if _, ok := functionList_1[fn.Name.Name]; !ok {
				return true
			}

			paramsTypeList := map[string]bool{}
			funcStr := functionList_1[fn.Name.Name]
			if paramsTypes, ok := typeMap[fn.Name.Name]; ok {
				for _, paramsType := range paramsTypes {
					if stuctInfo, ok := fastjsonPackageInfo.StructMap_2[paramsType]; ok {
						funcStr += stuctInfo
						paramsTypeList[paramsType] = true
					}
				}
			}
			funcStr = funcStr + addOptionForBoltdb(repo, paramsTypeList, 2)

			funcStr = funcStr + addFunSig(fn.Name.Name, "package_Info/boltdb/funSig_2.json")

			if repo == "fastjson" {
				funcStr += fastjsonPackageInfo.Conststr_2
			}

			functionList[fn.Name.Name] = funcStr
		}
		return true
	})
	return functionList
}

func extractFunctionLevel_3(filename string, typeFile string, repo string) map[string]string {
	file, err := os.Open(typeFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	functionList := map[string]string{}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	var typeMap map[string][]string

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&typeMap); err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			if repo == "boltdb" {
				if fn.Name.Name == "Error" {
					return true
				}
				if !unicode.IsUpper(rune(fn.Name.Name[0])) {
					return true
				}
			}
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
			paramsTypeList := map[string]bool{}
			if paramsTypes, ok := typeMap[fn.Name.Name]; ok {
				for _, paramsType := range paramsTypes {
					if stuctInfo, ok := fastjsonPackageInfo.StructMap_2[paramsType]; ok {
						funcStr += stuctInfo
						paramsTypeList[paramsType] = true
					}
				}
			}
			funcStr = funcStr + addOptionForBoltdb(repo, paramsTypeList, 3)

			funcStr = funcStr + addFunSig(fn.Name.Name, "package_Info/boltdb/funSig_3.json")

			if repo == "fastjson" {
				funcStr += fastjsonPackageInfo.Conststr_3
			}

			p := basePrompt
			p = strings.Replace(p, "{functionName}", fn.Name.Name, 1)
			funcStr = p + funcStr
			functionList[fn.Name.Name] = funcStr
		}
		return true
	})
	return functionList
}

func addFunSig(funcName string, sigJson string) string {
	data, err := os.ReadFile(sigJson)
	if err != nil {
		panic(err)
	}

	// 解析 JSON 到 map
	funcs := make(map[string]string)
	err = json.Unmarshal(data, &funcs)
	if err != nil {
		panic(err)
	}

	// 遍历 map 的 values 并拼接成一个字符串
	var combinedValues string
	for k, value := range funcs {
		if k == funcName {
			continue
		}
		combinedValues += value + ", " // 添加空格作为分隔符
	}
	return "Here are other function signatures defined in the same source file you may needed, DO NOT generate test functions for them." + combinedValues
}

func addOptionForBoltdb(repo string, paramTypeList map[string]bool, level int) string {
	if repo != "boltdb" {
		return ""
	}

	flag := true

	for k := range paramTypeList {
		if k == "Options" {
			flag = false
		}
	}
	if flag {
		if level == 2 {
			return boltpackageInfo.StructMap_2["Options"]
		} else {
			return boltpackageInfo.StructMap_3["Options"]
		}
	}
	return ""
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
			Model:       openai.GPT3Dot5Turbo0125,
			Messages:    promptMessages,
			Temperature: chatGPTemp,
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

func generateTest(client *openai.Client, sourceCodeList []string, generatedTestFile string, historyFile string, workers int) {
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
				prompt := sourceCodeList[k]
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

func removeFunction(testFile string, errors map[string]string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// 初始化一个新的声明列表
	var newDecls []ast.Decl
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if _, found := errors[funcDecl.Name.Name]; found {
				// 如果函数名在 toDelete 中，跳过此声明
				continue
			}
		}
		// 将非目标函数添加到新的声明列表中
		newDecls = append(newDecls, decl)
	}
	node.Decls = newDecls

	// 创建一个缓冲区并将 AST 输出为 Go 代码
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		panic(err)
	}

	// 将修改后的代码写回原始文件
	if err := os.WriteFile(testFile, buf.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func repairFailing(client *openai.Client, historyFile string, errorFile string, saveHistroyFile string, saveCompletionFile string, baseprompt string, workers int) {
	histories, err := loadSliceFromFile(historyFile)
	if err != nil {
		panic(err)
	}
	errorJson := loadErrorJson(errorFile)
	removeFunction(testFilePath, errorJson)

	errors := integration(errorJson, histories)

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
					extractedFunctions := extractedGeneratedCode(completion)
					funcNames := extractFuntionName(extractedFunctions)
					mutex.Lock()
					histories[k].FunctionNames = funcNames
					histories[k].History = append(histories[k].History, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: completion,
					})

					_, err := file.WriteString(extractedFunctions + "\n")
					if err != nil {
						panic(err)
					}
					mutex.Unlock()

				}
			}
		}(targetName, errorMsg)
	}

	wg.Wait()

	err = saveSliceToFile(histories, saveHistroyFile)
	if err != nil {
		panic(err)
	}

}

func isEmptyFunction(f *ast.FuncDecl) bool {
	// 检查函数体是否为空或仅包含注释
	if f.Body == nil || len(f.Body.List) == 0 {
		return true
	}
	return false
}

func collect_empty(testFilePath string) []string {
	emptyFunction := []string{}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFilePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// 遍历 AST
	ast.Inspect(file, func(n ast.Node) bool {
		// 检查是否为函数声明
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			// 检查函数是否为空
			if isEmptyFunction(funcDecl) {
				emptyFunction = append(emptyFunction, funcDecl.Name.Name)
			}
		}
		return true
	})

	return emptyFunction
}

type duJson struct {
	Origin    string `json:"origin"`
	Change    string `json:"change"`
	UniqueStr string `json:"uniqueStr"`
}

func modify_duplicated(duplicatedJsonFile string, histories *[]ChatHistory) {
	file, err := os.Open(duplicatedJsonFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 读取文件内容
	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var duplicatedJson []duJson
	err = json.Unmarshal(byteValue, &duplicatedJson)
	if err != nil {
		panic(err)
	}

	for _, duError := range duplicatedJson {
		for k, history := range *histories {
			chatMsgs := history.History
			lastMsg := chatMsgs[len(chatMsgs)-1]
			if strings.Contains(lastMsg.Content, duError.UniqueStr) {
				functionNames := history.FunctionNames
				for k, item := range functionNames {
					if item == duError.Origin {
						functionNames[k] = duError.Change
						break
					}
				}
				fmt.Println(1111)
				(*histories)[k].Duplicated = true
				(*histories)[k].FunctionNames = functionNames
				break
			}
		}
	}
}

func repairCompilation(client *openai.Client, historyFile string, errorFile string, saveHistroyFile string, saveCompletionFile string, baseprompt string, workers int, testFilePath string) {
	histories, err := loadSliceFromFile(historyFile)
	if err != nil {
		panic(err)
	}

	errorJson := loadErrorJson(errorFile)

	for k := range errorJson {
		errorJson[k] = baseprompt + errorJson[k]
	}

	emptyFunctions := collect_empty(testFilePath)
	fmt.Println("empty function", len(emptyFunctions))

	for _, emptyFunction := range emptyFunctions {
		errorJson[emptyFunction] = fmt.Sprintf("The function %s is empty, re-generate it again ", emptyFunction)
	}

	modify_duplicated("duplicated.json", &histories)

	errors := integration(errorJson, histories)

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
						Content: errorMsg,
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

	for k, history := range histories {
		chatMsgs := history.History
		lastMsg := chatMsgs[len(chatMsgs)-1]
		extractedFunctions := extractedGeneratedCode(lastMsg.Content)
		if !history.Duplicated {
			funcNames := extractFuntionName(extractedFunctions)
			histories[k].FunctionNames = funcNames
		}

		_, err := file.WriteString(extractedFunctions + "\n")
		if err != nil {
			panic(err)
		}
	}

	err = saveSliceToFile(histories, saveHistroyFile)
	if err != nil {
		panic(err)
	}
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

func loadErrorJson(errorFilePath string) map[string]string {
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
	return errorJSON
}

func integration(errorJSON map[string]string, histories []ChatHistory) map[string]string {
	names := [][]string{}
	for _, history := range histories {
		names = append(names, history.FunctionNames)
	}

	funcNeedFmt := make(map[string]bool)
	keyToRemove := make([]string, 0)

	for funcName, errMsg := range errorJSON {
		for _, funcNameList := range names {
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

func getFunctionSignType(sourceFile string, saveJsonFile string) {
	typeMap := map[string][]string{}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			typeMap[funcDecl.Name.Name] = []string{}
			if funcDecl.Recv != nil {
				for _, recv := range funcDecl.Recv.List {
					if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
						if ident, ok := starExpr.X.(*ast.Ident); ok {
							typeMap[funcDecl.Name.Name] = AppendIfNotPresent(typeMap[funcDecl.Name.Name], ident.Name)
						}
					} else if ident, ok := recv.Type.(*ast.Ident); ok {
						typeMap[funcDecl.Name.Name] = AppendIfNotPresent(typeMap[funcDecl.Name.Name], ident.Name)
					}
				}
			}

			for _, param := range funcDecl.Type.Params.List {
				for range param.Names {
					typeName := getTypeName(param.Type)
					typeMap[funcDecl.Name.Name] = AppendIfNotPresent(typeMap[funcDecl.Name.Name], typeName)
				}
			}

			if funcDecl.Type.Results != nil {
				for _, result := range funcDecl.Type.Results.List {
					resultTypeName := getTypeName(result.Type)
					typeMap[funcDecl.Name.Name] = AppendIfNotPresent(typeMap[funcDecl.Name.Name], resultTypeName)
				}
			}
		}
	}
	jsonData, err := json.MarshalIndent(typeMap, "", "    ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(saveJsonFile, jsonData, 0644)
	if err != nil {
		panic(err)
	}
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	case *ast.ArrayType:
		// 对于数组或切片，递归获取元素类型的名称
		elemType := getTypeName(t.Elt)
		return elemType

	case *ast.FuncType:
		return "need manual"
	}
	return ""
}

func AppendIfNotPresent(slice []string, item string) []string {
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}
