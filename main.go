package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	sourceCodeList := extractFunctionsFromCode("", "/Users/maike/Desktop/gonum/floats/floats.go", withOnlySourceCode)
	basePrompt := "generate test function for the following golang function, the test function should be in a new test file but in the same package"

	generateTest(client, sourceCodeList, basePrompt)
	fmt.Println(extractFunctionsFromCode("package aaa\n\n //commddent \nfunc ss(){\n}", "", withOnlySourceCode))
}
