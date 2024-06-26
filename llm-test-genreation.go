package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type ChatHistory struct {
	FunctionNames []string
	History       []openai.ChatCompletionMessage
}

// level 1 means with only source code
func extractFunctionLevel_1(codeStr string, filename string) (error, []string) {
	functionList := make([]string, 0)

	fset := token.NewFileSet()
	var node *ast.File
	var err error
	if codeStr != "" {
		node, err = parser.ParseFile(fset, "", codeStr, parser.ParseComments)
	} else {
		node, err = parser.ParseFile(fset, filename, nil, parser.ParseComments)
	}

	if err != nil {
		return err, nil
	}

	ast.Inspect(node, func(n ast.Node) bool {

		fn, ok := n.(*ast.FuncDecl)
		if ok {

			funcStart := fset.Position(fn.Pos()).Offset
			funcEnd := fset.Position(fn.End()).Offset

			var content []byte
			if codeStr != "" {
				content = []byte(codeStr)
			} else {
				content, err = os.ReadFile(filename)
				if err != nil {
					panic(err)
				}
			}

			functionList = append(functionList, strings.TrimSpace(string(content[funcStart:funcEnd])))
		}
		return true
	})

	return nil, functionList
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
	re := regexp.MustCompile("(?s)```go\n(.*?)\n```")
	matches := re.FindStringSubmatch(completion)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ""
	}
}

func extractFuntionName(str string) []string {
	var functionNames []string

	re := regexp.MustCompile(`func (\w+)`)
	matches := re.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		functionNames = append(functionNames, match[1])
	}
	return functionNames
}

func generateTest(client *openai.Client, sourceCodeList []string, basePrompt string, generatedTestFile string, historyFile string) {
	total := len(sourceCodeList)
	fmt.Println(total)
	if _, err := os.Stat(generatedTestFile); err == nil {
		err := os.Remove(generatedTestFile)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(generatedTestFile)

	if err != nil {
		panic(err)
	}

	histories := []ChatHistory{}

	defer file.Close()

	for k, sourceCode := range sourceCodeList {
		fmt.Println(k)
		prompt := basePrompt + "\n" + sourceCode
		completion := chat(client, prompt, nil)
		history := []openai.ChatCompletionMessage{}
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		}, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: completion,
		})
		functionList := extractFuntionName(completion)
		fmt.Println(functionList)
		chatH := ChatHistory{
			FunctionNames: functionList,
			History:       history,
		}
		histories = append(histories, chatH)

		testFunctionStr := extractTestFunction(completion)

		_, err = file.WriteString(fmt.Sprintf("%s\n\n", testFunctionStr))
		if err != nil {
			panic(err)
		}
	}

	err = saveSliceToFile(histories, historyFile)
	if err != nil {
		panic(err)
	}
}

func extractTestFunction(completion string) string {
	generatedCode := completion
	if strings.Contains(completion, "```") {
		generatedCode = extractCodeByRegex(completion)
	}

	var testFunctionList []string
	if strings.Contains(generatedCode, "package") {
		err, extractedFunctionList := extractFunctionLevel_1(generatedCode, "")
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

func repair(client *openai.Client, historyFile string, errorFile string, saveHistroyFile string, saveCompletionFile string, baseprompt string) {
	histories, err := loadSliceFromFile(historyFile)
	if err != nil {
		panic(err)
	}

	errors := parseCompliationJsonFile(errorFile)

	fmt.Println(len(errors))
	time := 0

	file, err := os.Create(saveCompletionFile)

	if err != nil {
		panic(err)
	}
	for targetName, cError := range errors {
		fmt.Println(time)
		time++
		prompt := baseprompt + cError
		for k, chatHistory := range histories {
			functionNames := chatHistory.FunctionNames
			flag := false
			for _, functionName := range functionNames {
				if functionName == targetName {
					flag = true
				}
			}
			if flag {
				histories[k].History = append(histories[k].History, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				})
				completion := chat(client, "", histories[k].History)
				histories[k].History = append(histories[k].History, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: completion,
				})

				extractedFunctions := extractTestFunction(completion)
				_, err := file.WriteString(extractedFunctions + "\n")

				if err != nil {
					panic(err)
				}

			}
		}

	}
	err = saveSliceToFile(histories, saveHistroyFile)
	if err != nil {
		panic(err)
	}
}

func parseCompliationJsonFile(filename string) map[string]string {
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
