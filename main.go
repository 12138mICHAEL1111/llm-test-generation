package main

import (
	"fmt"

	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
)

func addMapVToSlice(m map[string]string) []string {
	l := []string{}
	for _, v := range m {
		l = append(l, v)
	}
	return l
}

func geminigenerateTestLevel_1(model *genai.GenerativeModel, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_1(sourceFilePath, "gonum")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateGeminiTest(model, sourceCodeList, "gemini_generation/function/gonum/temp0.2/level_1/first_run/gonum_test.txt", "gemini_generation/history/gonum/temp0.2/level_1/first_run/gonum_history.gob", workers)
}

func geminigenerateTestLevel_2(model *genai.GenerativeModel, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_2(sourceFilePath, "package_Info/gonum/typeMap.json", "gonum")
	sourceCodeList := addMapVToSlice(sourceCodeMap)

	generateGeminiTest(model, sourceCodeList, "gemini_generation/function/gonum/temp0.2/level_2/first_run/gonum_test.txt", "gemini_generation/history/gonum/temp0.2/level_2/first_run/gonum_history.gob", workers)
}

func geminigenerateTestLevel_3(model *genai.GenerativeModel, sourceFilePath string, basePrompt string, workers int) {
	sourceCodeMap := extractFunctionLevel_3(sourceFilePath, "package_Info/gonum/typeMap.json", "gonum")
	sourceCodeList := addMapVToSlice(sourceCodeMap)
	generateGeminiTest(model, sourceCodeList, "gemini_generation/function/gonum/temp0.2/level_3/first_run/gonum_test.txt", "gemini_generation/history/gonum/temp0.2/level_3/first_run/gonum_history.gob", workers)
}

var sourceFilePath string
var testFilePath string
var basePrompt string
var errorFilePath string

func init() {
	sourceFilePath = gonumConfig.sourceFilePath
	testFilePath = gonumConfig.testFilePath
	basePrompt = gonumConfig.testGenerationBasePrompt
	errorFilePath = gonumConfig.errorFilePath
}

func check() {
	slice, _ := loadSliceFromFile("test_generation/history/gonum/temp0.2/level_3/second_run/compilation_fixed.gob")
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
	slice, _ := loadSliceFromFileGemini("gemini_generation/history/gonum/temp0.2/level_2/first_run/gonum_history.gob")
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
	// getFunctionSignType(sourceFilePath,"structs/gonum/typeMap.json")

	client := openai.NewClient("sk-proj-QSsxtUz5aqUMrvGDyzeDT3BlbkFJotWEWJh6tFd209iQd8VZ")

	// repairCompilation(client, "test_generation/history/gonum/temp0.2/level_3/first_run/gonum_history.gob", errorFilePath, "test_generation/history/gonum/temp0.2/level_3/second_run/compilation_fixed.gob", "test_generation/function/gonum/temp0.2/level_3/second_run/compilation_fixed.txt", compilationBasePrompt,5,testFilePath)
	// repairFailing(client, "test_generation/history/gonum/temp0.2/level_3/second_run/compilation_fixed.gob", errorFilePath, "test_generation/history/gonum/temp0.2/level_3/second_run/failed_fixed.gob", "test_generation/function/gonum/temp0.2/level_3/second_run/failed_fixed.txt", failedTestBasePrompt, 5)
	// ctx := context.Background()

	// client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyB94op2w6N6YYpck6cK8xcRCXebtEl9nlw"))
	// if err != nil {
	// 	panic(err)
	// }

	// model := client.GenerativeModel("gemini-1.5-pro")
	// model.SafetySettings = []*genai.SafetySetting{
	// 	{
	// 		Category:  genai.HarmCategoryHarassment,
	// 		Threshold: genai.HarmBlockNone,
	// 	},
	// 	{
	// 		Category:  genai.HarmCategoryDangerousContent,
	// 		Threshold: genai.HarmBlockNone,
	// 	},
	// }
	// model.SetTemperature(0.2)

	// geminigenerateTestLevel_3(model, sourceFilePath, basePrompt, 5)

	// removeFunction(testFilePath, loadErrorJson(errorFilePath))
	// geminiRepairCompilation(model, "gemini_generation/history/gonum/temp0.2/level_3/first_run/gonum_history.gob", errorFilePath, "gemini_generation/history/gonum/temp0.2/level_3/second_run/compilation_fixed.gob", "gemini_generation/function/gonum/temp0.2/level_3/second_run/compilation_fixed.txt", compilationBasePrompt,5,testFilePath)

	// geminiRepairFailing(model, "gemini_generation/history/gonum/temp0.2/level_3/second_run/compilation_fixed.gob", errorFilePath, "gemini_generation/history/gonum/temp0.2/level_3/second_run/failed_fixed.gob", "gemini_generation/function/gonum/temp0.2/level_3/second_run/failed_fixed.txt", failedTestBasePrompt, 5)

	// checkgemini()

	// -------------------
	// generatePromptFile("/Users/maike/Desktop/gonum/reports.json", "package_Info/gonum/typeMap.json", "gonum")
	generateCompletionFile(client, 30)
}
