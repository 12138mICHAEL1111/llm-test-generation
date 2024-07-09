package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func addMapVToSlice(m map[string]string)[]string{
	l := []string{}
	for _, v := range (m){
		l = append(l, v)
	}
	return l
}

func generateTestLevel_1(client *openai.Client, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_1(sourceFilePath,"boltdb")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/boltdb/temp0.2/level_1/first_run/boltdb_test.txt", "test_generation/history/boltdb/temp0.2/level_1/first_run/boltdb_history.gob", workers)
}

func generateTestLevel_2(client *openai.Client, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_2(sourceFilePath, "package_Info/boltdb/temp0.2/typeMap.json","boltdb")
	sourceCodeList := addMapVToSlice(sourceCodeMap)

	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/boltdb/temp0.2/level_2/first_run/boltdb_test.txt", "test_generation/history/boltdb/temp0.2/level_2/first_run/boltdb_history.gob", workers)
}

func generateTestLevel_3(client *openai.Client, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_3(sourceFilePath, "package_Info/boltdb/typeMap.json","boltdb")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateTest(client, sourceCodeList, basePrompt, "test_generation/function/boltdb/level_3/first_run/boltdb_test.txt", "test_generation/history/boltdb/temp0.2/level_3/first_run/boltdb_history.gob", workers)
}

var sourceFilePath string
var testFilePath string
var basePrompt string
var errorFilePath string

func init() {
	sourceFilePath = boltConfig.sourceFilePath
	testFilePath = boltConfig.testFilePath
	basePrompt = boltConfig.testGenerationBasePrompt
	errorFilePath = boltConfig.errorFilePath
}


func check() {
	slice, _ := loadSliceFromFile("test_generation/history/boltdb/temp0.2/level_1/first_run/boltdb_history.gob")
	for _, v := range slice {
		for _, n := range v.FunctionNames {
			if n == "TestStats" {
				fmt.Println(v.History)
				fmt.Println("----------")
			}
		}
	}
}


func main() {
	// getFunctionSignType(sourceFilePath,"structs/boltdb/typeMap.json")

	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	// generateTestLevel_1(client, sourceFilePath, basePrompt, 5)

	// repairCompilation(client, "test_generation/history/boltdb/temp0.2/level_1/first_run/boltdb_history.gob", errorFilePath, "test_generation/history/boltdb/temp0.2/level_1/second_run/compilation_fixed.gob", "test_generation/function/boltdb/temp0.2/level_1/second_run/compilation_fixed.txt", compilationBasePrompt,5,testFilePath)
	repairFailing(client, "test_generation/history/boltdb/temp0.2/level_1/second_run/compilation_fixed.gob", errorFilePath, "test_generation/history/boltdb/temp0.2/level_1/second_run/failed_fixed.gob", "test_generation/function/boltdb/temp0.2/level_1/second_run/failed_fixed.txt", failedTestBasePrompt,5)
	// check()

}
