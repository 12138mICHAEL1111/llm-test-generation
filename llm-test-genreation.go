package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"regexp"

	openai "github.com/sashabaranov/go-openai"
)

type mode = int8

const (
	withOnlySourceCode mode = 0
)

func extractFunctionsFromCode(codeStr string, filename string, mode mode) []string {
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
		panic(err)
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

	return functionList
}

func chat(client *openai.Client, prompt string) string {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo0125,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
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

func generateTest(client *openai.Client, sourceCodeList []string, basePrompt string) {
	total := len(sourceCodeList)
	fmt.Println(total)
	testFunctionFilename := "test_function.txt"
	if _, err := os.Stat(testFunctionFilename); err == nil {
		err := os.Remove(testFunctionFilename)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(testFunctionFilename)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	for k, sourceCode := range sourceCodeList {
		fmt.Println(k)
		prompt := basePrompt + "\n" + sourceCode
		completion := chat(client, prompt)

		generatedCode := completion
		if strings.Contains(completion, "```") {
			generatedCode = extractCodeFromCompletion(completion)
		}

		var testFunctionList []string
		if strings.Contains(generatedCode, "package") {
			testFunctionList = extractFunctionsFromCode(generatedCode, "", withOnlySourceCode)
		} else {
			testFunctionList = append(testFunctionList, generatedCode)
		}

		testFunctionStr := strings.Join(testFunctionList, "\n\n")

		_, err = file.WriteString(fmt.Sprintf("%s\n\n", testFunctionStr))
		if err != nil {
			panic(err)
		}
	}
}
