package main

import (
	"llm-test-generation/floats"

	openai "github.com/sashabaranov/go-openai"
)

func generateTestLevel_1(client *openai.Client, sourceFilePath string) {
	_, sourceCodeList := extractFunctionLevel_1("", sourceFilePath)
	basePrompt := "generate test function for the following golang function, the test function should be in a new test file but in the same package"

	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/floats_test_1.txt","test_generation/history/floats/first/floats_history_1.gob")
}

func generateTestLevel_2(client *openai.Client, sourceFilePath string) {
	sourceCodeList := extractFunctionLevel_2(sourceFilePath)
	basePrompt := "generate test function for the following golang function, the test function should be in a new test file but in the same package"

	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/floats_test_2.txt","test_generation/history/floats/first/floats_history_2.gob")
}

func generateTestLevel_3(client *openai.Client, sourceFilePath string) {
	baseFunctionDoc := floats.GetBaseFunctionDoc()
	sourceCodeList := floats.ExtractFunctionLevel_3(sourceFilePath, baseFunctionDoc)
	basePrompt := "generate test function for the following golang function, the test function should be in a new test file but in the same package"

	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/floats_test_3.txt","test_generation/history/floats/first/floats_history_3.gob")
}

func main() {
	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	sourceFilePath := "/Users/maike/Desktop/gonum/floats/floats.go"

	// generateTestLevel_2(client,sourceFilePath)
	generateTestLevel_1(client, sourceFilePath)

}
