package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"sync"

	"github.com/google/generative-ai-go/genai"
)

type GeminiChatHistory struct {
	FunctionNames []string
	History       []*genai.Content
}

func geminiChat(model *genai.GenerativeModel, prompt string, messages []*genai.Content) genai.Text {
	cs := model.StartChat()
	if messages != nil {
		cs.History = messages
	}
	resp, err := cs.SendMessage(context.Background(), genai.Text(prompt))

	if err != nil {
		fmt.Println(222)
		return ""
	}

	c, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		panic("fail to parse resp")
	}
	return c
}

func generateGeminiTest(model *genai.GenerativeModel, sourceCodeList []string, generatedTestFile string, historyFile string, workers int) {
	total := len(sourceCodeList)
	file, err := os.Create(generatedTestFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	chunks := total / workers

	var allHistories []GeminiChatHistory
	for i := 0; i < workers; i++ {
		start := i * chunks
		end := start + chunks
		if i == workers-1 {
			end = total
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			for k := start; k < end; k++ {
				prompt := sourceCodeList[k]
				completion := geminiChat(model, prompt, nil)

				history := []*genai.Content{
					{
						Role: "user",
						Parts: []genai.Part{
							genai.Text(prompt),
						},
					},
					{
						Role: "model",
						Parts: []genai.Part{
							genai.Text(completion),
						},
					},
				}

				testFunctionStr := extractedGeneratedCode(string(completion))

				functionNameList := extractFuntionName(testFunctionStr)
				chatH := GeminiChatHistory{
					FunctionNames: functionNameList,
					History:       history,
				}

				mutex.Lock()
				allHistories = append(allHistories, chatH)
				_, err = file.WriteString(fmt.Sprintf("%s\n\n", testFunctionStr))
				mutex.Unlock()

				if err != nil {
					panic(err)
				}
			}
		}(start, end)
	}

	wg.Wait() // 等待所有goroutine完成
	fmt.Println("All goroutines completed.")
	// 保存所有历史记录到文件
	err = saveSliceToFile(allHistories, historyFile)
	if err != nil {
		panic(err)
	}

}

func geminiRepairCompilation(model *genai.GenerativeModel, historyFile string, errorFile string, saveHistroyFile string, saveCompletionFile string, baseprompt string, workers int, testFilePath string) {
	histories, err := loadSliceFromFileGemini(historyFile)
	if err != nil {
		panic(err)
	}

	errorJson := loadErrorJson(errorFile)

	for k := range errorJson {
		errorJson[k] = baseprompt + errorJson[k]
	}

	emptyFunctions := collect_empty(testFilePath)
	fmt.Println("empty function", len(emptyFunctions))

	for _, emptyFunction := range emptyFunctions {
		errorJson[emptyFunction] = fmt.Sprintf("The function %s is empty, re-generate it again ", emptyFunction)
	}

	errors := geminiIntegration(errorJson, histories)

	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, workers)

	file, err := os.Create(saveCompletionFile)

	if err != nil {
		panic(err)
	}

	testfile, err := os.OpenFile(testFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer testfile.Close()

	for targetName, errorMsg := range errors {
		wg.Add(1)
		sem <- struct{}{}
		fmt.Println(targetName)
		go func(targetName, errorMsg string) {
			defer wg.Done()
			defer func() { <-sem }()

			for k, chatHistory := range histories {
				functionNames := chatHistory.FunctionNames
				flag := false
				for _, functionName := range functionNames {
					if functionName == targetName {
						flag = true
					}
				}
				if flag {
					completion := geminiChat(model, errorMsg, histories[k].History)
					mutex.Lock()
					histories[k].History = append(histories[k].History, &genai.Content{
						Role: "user",
						Parts: []genai.Part{
							genai.Text(errorMsg),
						},
					},
						&genai.Content{
							Role: "model",
							Parts: []genai.Part{
								genai.Text(completion),
							},
						})
					extractedFunctions := extractedGeneratedCode(string(completion))
					if _, err := testfile.WriteString(extractedFunctions + "\n"); err != nil {
						panic(err)
					}
					mutex.Unlock()

				}
			}
		}(targetName, errorMsg)
	}

	wg.Wait()
	fmt.Println("All goroutines completed.")

	for k, history := range histories {
		chatMsgs := history.History
		lastMsg := chatMsgs[len(chatMsgs)-1]
		c, ok := lastMsg.Parts[0].(genai.Text)
		if !ok {
			panic("fail to parse resp")
		}
		extractedFunctions := extractedGeneratedCode(string(c))
		funcNames := extractFuntionName(extractedFunctions)
		histories[k].FunctionNames = funcNames
		_, err := file.WriteString(extractedFunctions + "\n")
		if err != nil {
			panic(err)
		}
	}

	err = saveSliceToFile(histories, saveHistroyFile)
	if err != nil {
		panic(err)
	}
}

func loadSliceFromFileGemini(filename string) ([]GeminiChatHistory, error) {
	gob.Register(genai.Text(""))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var slice []GeminiChatHistory
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&slice); err != nil {
		return nil, err
	}

	return slice, nil
}

func geminiIntegration(errorJSON map[string]string, histories []GeminiChatHistory) map[string]string {
	names := [][]string{}
	for _, history := range histories {
		names = append(names, history.FunctionNames)
	}

	funcNeedFmt := make(map[string]bool)
	keyToRemove := make([]string, 0)

	for funcName, errMsg := range errorJSON {
		for _, funcNameList := range names {
			if len(funcNameList) == 1 {
				continue
			}
			for index, name := range funcNameList {
				if name == funcName && index > 0 {
					targetFuncName := funcNameList[0]
					if _, exists := errorJSON[targetFuncName]; !exists {
						errorJSON[targetFuncName] = fmt.Sprintf("%s: %s", funcName, errMsg)
					} else {
						funcNeedFmt[targetFuncName] = true
						errorJSON[targetFuncName] += fmt.Sprintf("%s: %s", funcName, errMsg)
					}
					keyToRemove = append(keyToRemove, funcName)
					break
				}
			}
		}
	}

	for _, key := range keyToRemove {
		delete(errorJSON, key)
	}

	for f := range funcNeedFmt {
		errorJSON[f] = fmt.Sprintf("%s: %s", f, errorJSON[f])
	}

	for k := range errorJSON {
		errorJSON[k] = errorJSON[k][:len(errorJSON[k])-1]
	}
	return errorJSON
}

func geminiRepairFailing(model *genai.GenerativeModel, historyFile string, errorFile string, saveHistroyFile string, saveCompletionFile string, baseprompt string, workers int) {
	histories, err := loadSliceFromFileGemini(historyFile)
	if err != nil {
		panic(err)
	}
	errorJson := loadErrorJson(errorFile)
	removeFunction(testFilePath, errorJson)

	errors := geminiIntegration(errorJson, histories)
	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, workers)

	file, err := os.Create(saveCompletionFile)

	if err != nil {
		panic(err)
	}

	testfile, err := os.OpenFile(testFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer testfile.Close()

	for targetName, errorMsg := range errors {
		wg.Add(1)
		sem <- struct{}{}
		go func(targetName, errorMsg string) {
			defer wg.Done()
			defer func() { <-sem }()

			prompt := baseprompt + errorMsg

			for k, chatHistory := range histories {
				functionNames := chatHistory.FunctionNames
				flag := false
				for _, functionName := range functionNames {
					if functionName == targetName {
						flag = true
					}
				}
				if flag {
					completion := geminiChat(model, prompt, histories[k].History)
					extractedFunctions := extractedGeneratedCode(string(completion))
					funcNames := extractFuntionName(extractedFunctions)
					mutex.Lock()
					histories[k].FunctionNames = funcNames
					histories[k].History = append(histories[k].History, &genai.Content{
						Role: "user",
						Parts: []genai.Part{
							genai.Text(prompt),
						},
					},
						&genai.Content{
							Role: "model",
							Parts: []genai.Part{
								genai.Text(completion),
							},
						})
					_, err := file.WriteString(extractedFunctions + "\n")
					if err != nil {
						panic(err)
					}
					if _, err := testfile.WriteString(extractedFunctions + "\n"); err != nil {
						panic(err)
					}
					mutex.Unlock()

				}
			}
		}(targetName, errorMsg)
	}

	wg.Wait()

	err = saveSliceToFile(histories, saveHistroyFile)
	if err != nil {
		panic(err)
	}

}
