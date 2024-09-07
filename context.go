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
	filename := "/Users/maike/Desktop/fastjson/parser.go" 
	jsonFilename := "package_Info/fastjson/funSig_2.json"        
	wJson := "package_Info/fastjson/funSig_3.json"

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

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			funcName := fn.Name.Name
			if currentValue, exists := funcs[funcName]; exists {
				if fn.Doc != nil {
					funcs[funcName] = currentValue + fn.Doc.Text()
				}
			}
		}
		return true 
	})

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
