package main

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func addMapVToSlice(m map[string]string) []string {
	l := []string{}
	for _, v := range m {
		l = append(l, v)
	}
	return l
}

func geminigenerateTestLevel_1(model *genai.GenerativeModel, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_1(sourceFilePath, "boltdb")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateGeminiTest(model, sourceCodeList, "gemini_generation/function/boltdb/temp0.2/level_1/first_run/boltdb_test.txt", "gemini_generation/history/boltdb/temp0.2/level_1/first_run/boltdb_history.gob", workers)
}

func geminigenerateTestLevel_2(model *genai.GenerativeModel, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_2(sourceFilePath, "package_Info/boltdb/typeMap.json", "boltdb")
	sourceCodeList := addMapVToSlice(sourceCodeMap)

	generateGeminiTest(model, sourceCodeList, "gemini_generation/function/boltdb/temp0.2/level_2/first_run/boltdb_test.txt", "gemini_generation/history/boltdb/temp0.2/level_2/first_run/boltdb_history.gob", workers)
}

func geminigenerateTestLevel_3(model *genai.GenerativeModel, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_3(sourceFilePath, "package_Info/boltdb/typeMap.json", "boltdb")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateGeminiTest(model, sourceCodeList, "gemini_generation/function/boltdb/temp0.2/level_3/first_run/boltdb_test.txt", "gemini_generation/history/boltdb/temp0.2/level_3/first_run/boltdb_history.gob", workers)
}

var sourceFilePath string
var testFilePath string
var basePrompt string
var errorFilePath string
var repo string

func init() {
	repo = "boltdb"
	sourceFilePath = boltdbConfig.sourceFilePath
	testFilePath = boltdbConfig.testFilePath
	basePrompt = boltdbConfig.testGenerationBasePrompt
	errorFilePath = boltdbConfig.errorFilePath
}

func check() {
	slice, _ := loadSliceFromFile("test_generation/history/boltdb/temp0.2/level_3/second_run/compilation_fixed.gob")
	for _, v := range slice {
		for _, n := range v.FunctionNames {
			if n == "TestGetInt64" {
				fmt.Println(v.History)
				fmt.Println("----------")
			}
		}
	}
}

func checkgemini() {
	slice, _ := loadSliceFromFileGemini("gemini_generation/history/boltdb/temp0.2/level_2/first_run/boltdb_history.gob")
	for _, v := range slice {
		for _, n := range v.FunctionNames {
			if n == "TestObject_reset" {
				for _, h := range v.History {
					s := h.Parts[0]
					fmt.Println(s)
					fmt.Println("--------------------")
				}
			}
		}
	}
}

func main() {
	// getFunctionSignType(sourceFilePath,"structs/boltdb/typeMap.json")

	// client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	// repairCompilation(client, "test_generation/history/boltdb/temp0.2/level_3/first_run/boltdb_history.gob", errorFilePath, "test_generation/history/boltdb/temp0.2/level_3/second_run/compilation_fixed.gob", "test_generation/function/boltdb/temp0.2/level_3/second_run/compilation_fixed.txt", compilationBasePrompt,5,testFilePath)
	// repairFailing(client, "test_generation/history/boltdb/temp0.2/level_3/second_run/compilation_fixed.gob", errorFilePath, "test_generation/history/boltdb/temp0.2/level_3/second_run/failed_fixed.gob", "test_generation/function/boltdb/temp0.2/level_3/second_run/failed_fixed.txt", failedTestBasePrompt, 5)
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyB94op2w6N6YYpck6cK8xcRCXebtEl9nlw"))
	if err != nil {
		panic(err)
	}

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	}
	model.SetTemperature(1.0)

	// geminigenerateTestLevel_3(model, sourceFilePath, basePrompt, 5)

	// removeFunction(testFilePath, loadErrorJson(errorFilePath))
	// geminiRepairCompilation(model, "gemini_generation/history/boltdb/temp0.2/level_3/first_run/boltdb_history.gob", errorFilePath, "gemini_generation/history/boltdb/temp0.2/level_3/second_run/compilation_fixed.gob", "gemini_generation/function/boltdb/temp0.2/level_3/second_run/compilation_fixed.txt", compilationBasePrompt,5,testFilePath)

	// geminiRepairFailing(model, "gemini_generation/history/boltdb/temp0.2/level_3/second_run/compilation_fixed.gob", errorFilePath, "gemini_generation/history/boltdb/temp0.2/level_3/second_run/failed_fixed.gob", "gemini_generation/function/boltdb/temp0.2/level_3/second_run/failed_fixed.txt", failedTestBasePrompt, 5)

	// checkgemini()

	// -------------------
	// generatePromptFile("/Users/maike/go/src/github.com/boltdb/bolt/reports.json", "package_Info/boltdb/typeMap.json")
	generateCompletionFile_Gemini(model, 80)

}
