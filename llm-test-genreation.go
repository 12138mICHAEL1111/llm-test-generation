package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"llm-test-generation/util"
	"os"
	"strings"

	"regexp"

	openai "github.com/sashabaranov/go-openai"
)

// func getTestedFunction(testFilename string) []string{
// 	functionList := make([]string, 0)

// 	fset := token.NewFileSet()
// 	node, err := parser.ParseFile(fset, testFilename, nil, parser.ParseComments)

// 	if err != nil {
// 		panic(err)
// 	}

// 	ast.Inspect(node, func(n ast.Node) bool {

// 		fn, ok := n.(*ast.FuncDecl)
// 		if ok {
// 			name := fn.Name.Name
// 			if strings.HasPrefix(name,"Test"){
// 				name = name[4:]
// 				functionList = append(functionList, name)
// 			}

// 		}
// 		return true
// 	})
// 	fmt.Println(functionList)
// 	return functionList
// }

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

func extractCodeFromCompletion(completion string) string {
	re := regexp.MustCompile("(?s)```go\n(.*?)\n```")
	matches := re.FindStringSubmatch(completion)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ""
	}
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

	histories := [][]openai.ChatCompletionMessage{}

	defer file.Close()

	for k, sourceCode := range sourceCodeList {
		fmt.Println(k)
		prompt := basePrompt + "\n" + sourceCode
		completion := chat(client, prompt,nil)
		history := []openai.ChatCompletionMessage{}
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		}, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: completion,
		})
		histories = append(histories, history)

		generatedCode := completion
		if strings.Contains(completion, "```") {
			generatedCode = extractCodeFromCompletion(completion)
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

		_, err = file.WriteString(fmt.Sprintf("%s\n\n", testFunctionStr))
		if err != nil {
			panic(err)
		}
	}

	err = util.SaveSliceToFile(histories, historyFile)
	if err != nil {
		panic(err)
	}
}

func repair(client *openai.Client, filename string, functionName string, numOfRepair int, prompt string) {
	histories, err := util.LoadSliceFromFile(filename)
	if err != nil {
		panic(err)
	}

	baseprompt := "the code generated has compilation faults, fix them. For code t := []float64{7.0, 8.0, 9.0}, no new variables on left side of := and cannot use []float64{…} (value of type []float64) as *testing.T value in assignment. For code result := AddTo(dst, s, t), cannot use t (variable of type *testing.T) as []float64 value in argument to AddTo"
	for _, history := range histories {
		assistantMsg := history[2*numOfRepair-1].Content
		if strings.Contains(assistantMsg, functionName) {
			history = append(history, openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleUser,
				Content: baseprompt,
			})
			completion := chat(client,"",history)
			history = append(history, openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleAssistant,
				Content: completion,
			})
			util.LoadSliceFromFile(filename)
			fmt.Println(completion)
			break
		}
	}

}
