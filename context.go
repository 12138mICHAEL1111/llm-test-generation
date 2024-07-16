package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

func contextGen() {
	filename := "/Users/maike/Desktop/fastjson/parser.go" // Go 源文件
	jsonFilename := "package_Info/fastjson/funSig_2.json"        // JSON 文件名
	wJson := "package_Info/fastjson/funSig_3.json"

	// 读取 JSON 文件到 map
	data, err := ioutil.ReadFile(jsonFilename)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	funcs := make(map[string]string)
	err = json.Unmarshal(data, &funcs)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}

	// 解析 Go 文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}

	// 遍历语法树
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			funcName := fn.Name.Name
			// 检查 map 中是否有这个函数
			if currentValue, exists := funcs[funcName]; exists {
				// 如果函数有文档注释，添加到 map 的 value
				if fn.Doc != nil {
					funcs[funcName] = currentValue + fn.Doc.Text()
				}
			}
		}
		return true // 继续遍历树
	})

	// 将更新的数据写回 JSON 文件
	updatedData, err := json.MarshalIndent(funcs, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	err = ioutil.WriteFile(wJson, updatedData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}

	fmt.Println("Updated JSON file successfully.")
}
