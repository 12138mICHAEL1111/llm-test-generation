package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func generateTestLevel_1(client *openai.Client, sourceFilePath string) {
	_, sourceCodeList := extractFunctionLevel_1("", sourceFilePath)
	generateTest(client, sourceCodeList, testGenerationBasePrompt, "test_generation/function/floats/firstrun/level_1/floats_test_1.txt", "test_generation/history/floats/firstrun/floats_history_1.gob")
}

func generateTestLevel_2(client *openai.Client, sourceFilePath string) {
	sourceCodeList := extractFunctionLevel_2(sourceFilePath)
	generateTest(client, sourceCodeList, testGenerationBasePrompt, "test_generation/function/floats/level_2/first_run/floats_test.txt", "test_generation/history/floats/level_2/first_run/floats_history.gob")
}

func generateTestLevel_3(client *openai.Client, sourceFilePath string) {
	baseFunctionDoc := GetBaseFunctionDoc()
	sourceCodeList := ExtractFunctionLevel_3(sourceFilePath, baseFunctionDoc)

	generateTest(client, sourceCodeList, testGenerationBasePrompt, "test_generation/function/floats/level_3/first_run/floats_test.txt", "test_generation/history/floats/level_3/first_run/floats_history.gob")
}

func check() {
	slice, _ := loadSliceFromFile("test_generation/history/floats/level_3/second_run/compilation_fixed.gob")
	for _, v := range slice {
		for _, n := range v.FunctionNames {
			if n == "TestCount" {
				fmt.Print(v.History)
			}
		}
	}
}

func main() {
	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")
	
	// generateTestLevel_3(client, sourceFilePath)

	repair(client, "test_generation/history/floats/level_3/second_run/failed_fixed.gob", "test_generation/function/floats/level_3/second_run/failed.json", "test_generation/history/floats/level_3/third_run/failed_fixed.gob", "test_generation/function/floats/level_3/third_run/failed.txt", thirdCompilationBasePrompt)
	// check()
}
