package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func generateTestLevel_1(client *openai.Client, sourceFilePath string,basePrompt string,workers int) {
	_, sourceCodeList := extractFunctionLevel_1("", sourceFilePath)
	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/boltdb/level_1/first_run/boltdb_test.txt", "test_generation/history/boltdb/level_1/first_run/boltdb_history.gob",workers)
}

func generateTestLevel_2(client *openai.Client, sourceFilePath string, basePrompt string,workers int) {
	sourceCodeList := extractFunctionLevel_2(sourceFilePath)
	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/level_2/first_run/floats_test.txt", "test_generation/history/floats/level_2/first_run/floats_history.gob",workers)
}

func generateTestLevel_3(client *openai.Client, sourceFilePath string,basePrompt string, workers int) {
	baseFunctionDoc := GetBaseFunctionDoc()
	sourceCodeList := ExtractFunctionLevel_3(sourceFilePath, baseFunctionDoc)

	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/floats/level_3/first_run/floats_test.txt", "test_generation/history/floats/level_3/first_run/floats_history.gob", workers)
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

var sourceFilePath string
var testFilePath string
var basePrompt string

func init(){
	sourceFilePath = boltConfig.sourceFilePath
	testFilePath = boltConfig.testFilePath
	basePrompt = boltConfig.testGenerationBasePrompt
}

func main() {
	// fmt.Print(extractSourceFuntionName(sourceFilePath))
	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")
	
	generateTestLevel_1(client, sourceFilePath,basePrompt,5)

	// repair(client, "test_generation/history/floats/level_3/second_run/failed_fixed.gob", "test_generation/function/floats/level_3/second_run/failed.json", "test_generation/history/floats/level_3/third_run/failed_fixed.gob", "test_generation/function/floats/level_3/third_run/failed.txt", thirdCompilationBasePrompt)
	// check()
	// removeFailedTest(testFilePath,"error.json")
}
