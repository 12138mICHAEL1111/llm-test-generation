package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func addMapVToSlice(m map[string]string) []string {
	l := []string{}
	for _, v := range m {
		l = append(l, v)
	}
	return l
}

func generateTestLevel_1(client *openai.Client, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_1(sourceFilePath, "fastjson")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateTest(client, sourceCodeList, "test_generation/function/fastjson/temp1.0/level_1/first_run/fastjson_test.txt", "test_generation/history/fastjson/temp1.0/level_1/first_run/fastjson_history.gob", workers)
}

func generateTestLevel_2(client *openai.Client, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_2(sourceFilePath, "package_Info/fastjson/typeMap.json", "fastjson")
	sourceCodeList := addMapVToSlice(sourceCodeMap)

	generateTest(client, sourceCodeList, "test_generation/function/fastjson/temp1.0/level_2/first_run/fastjson_test.txt", "test_generation/history/fastjson/temp1.0/level_2/first_run/fastjson_history.gob", workers)
}

func generateTestLevel_3(client *openai.Client, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_3(sourceFilePath, "package_Info/fastjson/typeMap.json", "fastjson")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateTest(client, sourceCodeList, "test_generation/function/fastjson/temp1.0/level_3/first_run/fastjson_test.txt", "test_generation/history/fastjson/temp1.0/level_3/first_run/fastjson_history.gob", workers)
}

var sourceFilePath string
var testFilePath string
var basePrompt string
var errorFilePath string

func init() {
	sourceFilePath = fastjsonConfig.sourceFilePath
	testFilePath = fastjsonConfig.testFilePath
	basePrompt = fastjsonConfig.testGenerationBasePrompt
	errorFilePath = fastjsonConfig.errorFilePath
}

func check() {
	slice, _ := loadSliceFromFile("test_generation/history/fastjson/temp1.0/level_1/first_run/fastjson_history.gob")
	for _, v := range slice {
		for _, n := range v.FunctionNames {
			if n == "TestType" {
				fmt.Println(v.History)
				fmt.Println("----------")
			}
		}
	}
}

func main() {
	// getFunctionSignType(sourceFilePath,"structs/fastjson/typeMap.json")

	// client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	// generateTestLevel_1(client, sourceFilePath, basePrompt, 5)

	// repairCompilation(client, "test_generation/history/fastjson/temp1.0/level_1/first_run/fastjson_history.gob", errorFilePath, "test_generation/history/fastjson/temp1.0/level_1/second_run/compilation_fixed.gob", "test_generation/function/fastjson/temp1.0/level_1/second_run/compilation_fixed.txt", compilationBasePrompt,5,testFilePath)
	// repairFailing(client, "test_generation/history/fastjson/temp1.0/level_1/second_run/compilation_fixed.gob", errorFilePath, "test_generation/history/fastjson/temp1.0/level_1/second_run/failed_fixed.gob", "test_generation/function/fastjson/temp1.0/level_1/second_run/failed_fixed.txt", failedTestBasePrompt, 5)
	// removeFunction(testFilePath, loadErrorJson(errorFilePath))
	// check()

	contextGen()
	// getFunctionSignType(sourceFilePath,"package_Info/fastjson/typeMap.json")
}
