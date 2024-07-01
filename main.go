package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func generateTestLevel_1(client *openai.Client, sourceFilePath string, basePrompt string, workers int, funcNameFilePath string) {
	sourceCodeList := extractFunctionLevel_1(sourceFilePath)
	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/boltdb/level_1/first_run/boltdb_test.txt", "test_generation/history/boltdb/level_1/first_run/boltdb_history.gob", workers, funcNameFilePath)
}

func generateTestLevel_2(client *openai.Client, sourceFilePath string, basePrompt string, workers int, funcNameFilePath string) {
	sourceCodeList := extractFunctionLevel_2(sourceFilePath)
	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/level_2/first_run/floats_test.txt", "test_generation/history/floats/level_2/first_run/floats_history.gob", workers, funcNameFilePath)
}

func generateTestLevel_3(client *openai.Client, sourceFilePath string, basePrompt string, workers int, funcNameFilePath string) {
	baseFunctionDoc := GetBaseFunctionDoc()
	sourceCodeList := ExtractFunctionLevel_3(sourceFilePath, baseFunctionDoc)

	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/level_3/first_run/floats_test.txt", "test_generation/history/floats/level_3/first_run/floats_history.gob", workers, funcNameFilePath)
}

func check() {
	slice, _ := loadSliceFromFile("test_generation/history/boltdb/level_1/third_run/compilation_fixed.gob")
	for _, v := range slice {
		for _, n := range v.FunctionNames {
			if n == "TestGrow" {
				fmt.Print(v.History)
			}
		}
	}
}

var sourceFilePath string
var testFilePath string
var basePrompt string
var errorFilePath string
var funcNameFilePath string = "func_names.json"

func init() {
	sourceFilePath = boltConfig.sourceFilePath
	testFilePath = boltConfig.testFilePath
	basePrompt = boltConfig.testGenerationBasePrompt
	errorFilePath = boltConfig.errorFilePath
}

func main() {
	// fmt.Print(extractSourceFuntionName(sourceFilePath))
	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	// generateTestLevel_1(client, sourceFilePath, basePrompt, 5, funcNameFilePath)

	repair(client, "test_generation/history/boltdb/level_1/second_run/compilation_fixed.gob", errorFilePath, "test_generation/history/boltdb/level_1/third_run/compilation_fixed.gob", "test_generation/function/boltdb/level_1/third_run/compilation_fixed.txt", thirdCompilationBasePrompt,5,funcNameFilePath)
	// check()

}
