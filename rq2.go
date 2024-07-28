package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	fastjsonPackageInfo "llm-test-generation/package_Info/fastjson"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Mutator 代表 "mutator" 键的值
type Mutator struct {
	MutatedSourceCode string `json:"mutatedSourceCode"`
}

// MutationRecord 代表 "killed" 数组中的一个元素
type MutationRecord struct {
	Mutator  Mutator `json:"mutator"`
	Checksum string  `json:"checksum"`
}

// MutationReport 代表整个 JSON 对象
type MutationReport struct {
	Killed []MutationRecord `json:"killed"`
}

func rq2_extractFunctionLevel_3(src string, typeFile string, repo string, funName string) string {
	file, err := os.Open(typeFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	var typeMap map[string][]string

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&typeMap); err != nil {
		panic(err)
	}

	funcStr := ""
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok {

			if funName != fn.Name.Name {
				return true
			}

			funcStart := fset.Position(fn.Pos()).Offset
			funcEnd := fset.Position(fn.End()).Offset
			docString := fn.Doc.Text()
			funcCode := strings.TrimSpace(string(src[funcStart:funcEnd]))
			funcStr = docString + funcCode
			paramsTypeList := map[string]bool{}
			if paramsTypes, ok := typeMap[fn.Name.Name]; ok {
				for _, paramsType := range paramsTypes {
					if stuctInfo, ok := fastjsonPackageInfo.StructMap_3[paramsType]; ok {
						funcStr += stuctInfo
						paramsTypeList[paramsType] = true
					}
				}
			}

			funcStr = funcStr + addFunSig(fn.Name.Name, "package_Info/gonum/funSig_3.json")

			if repo == "fastjson" {
				funcStr += fastjsonPackageInfo.Conststr_3
			}

			p := basePrompt
			p = strings.Replace(p, "{functionName}", fn.Name.Name, 1)
			funcStr = p + funcStr
		}
		return true
	})
	return funcStr
}

func rq_2_generate(client *openai.Client, promptMap map[string]string, workers int) map[string]string {
	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, workers)

	completionMap := map[string]string{}
	counter := 0
	for checksum, prompt := range promptMap {
		wg.Add(1)
		sem <- struct{}{}
		go func(checksum, prompt string) {
			defer wg.Done()
			defer func() { <-sem }()

			var completion string
			if prompt != "" {
				completion = chat(client, prompt, nil)
			}

			completion = extractFirstCodeByRegex(completion)
			mutex.Lock()

			completionMap[checksum] = completion
			mutex.Unlock()
		}(checksum, prompt)

		counter++
		fmt.Println(counter)
		if counter%50 == 0 {
			fmt.Printf("Processed %d items, pausing for 1m\n", counter)
			time.Sleep(1 * time.Minute)
		}
	}

	wg.Wait()
	return completionMap
}

func extractFirstCodeByRegex(completion string) string {
	re := regexp.MustCompile("(?s)```.*?\n(.*?)\n```")
	match := re.FindStringSubmatch(completion)
	if match != nil {
		return match[1]
	}
	return completion
}

func generatePromptFile(reportFile string, typefile string, repo string) {
	functionNamedata, err := os.ReadFile("function_names.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 创建 map 用来存储 JSON 数据
	var functionNameMap map[string]string

	// 解析 JSON 数据到 map
	err = json.Unmarshal(functionNamedata, &functionNameMap)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	reportData, err := os.ReadFile(reportFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var report MutationReport
	err = json.Unmarshal(reportData, &report)
	if err != nil {
		log.Fatalf("Error parsing JSON: %s", err)
	}

	promptMap := map[string]string{}
	for checksum, functionName := range functionNameMap {
		for _, mutants := range report.Killed {
			if mutants.Checksum == checksum {
				prompt := rq2_extractFunctionLevel_3(mutants.Mutator.MutatedSourceCode, typefile, repo, functionName)
				promptMap[checksum] = prompt
				break
			}
		}
	}

	promptFile, err := os.Create("prompt.json")
	if err != nil {
		panic(err)
	}

	j, err := json.Marshal(promptMap)
	if err != nil {
		panic(err)
	}

	_, err = promptFile.Write(j)
	if err != nil {
		panic(err)
	}
}

func generateCompletionFile(client *openai.Client, workers int) {
	promptdata, err := os.ReadFile("prompt.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 创建 map 用来存储 JSON 数据
	var promptMap map[string]string

	// 解析 JSON 数据到 map
	err = json.Unmarshal(promptdata, &promptMap)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	completionMap := rq_2_generate(client, promptMap, workers)

	promptFile, err := os.Create("rq2_completion/gonum/completion_1.json")
	if err != nil {
		panic(err)
	}

	j, err := json.Marshal(completionMap)
	if err != nil {
		panic(err)
	}

	_, err = promptFile.Write(j)
	if err != nil {
		panic(err)
	}
}
