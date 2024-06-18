package floats

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

const structType = "// argsort is a helper that implements sort.Interface, as used by// Argsort and ArgsortStable.\ntype argsort struct {\ns []float64 \n inds []int\n}"

func GetBaseFunctionDoc() map[string]string {
	baseFunctionMap := make(map[string]string)
	filepath := "/Users/maike/Desktop/gonum/internal/asm/f64/stubs_amd64.go"

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {

		fn, ok := n.(*ast.FuncDecl)
		if ok {
			docString := fn.Doc.Text()
			name := fn.Name.Name
			baseFunctionMap[name] = docString
		}
		return true
	})
	return baseFunctionMap
}

func ExtractFunctionLevel_3(filename string, baseFunctionMap map[string]string) []string {
	functionList := make([]string, 0)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		fmt.Println(err)
	}

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		funcStart := fset.Position(funcDecl.Pos()).Offset
		funcEnd := fset.Position(funcDecl.End()).Offset
		docString := funcDecl.Doc.Text()
		var content []byte
		content, err = os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		funcCode := strings.TrimSpace(string(content[funcStart:funcEnd]))

		funcStr := docString + funcCode

		functionName := funcDecl.Name.Name
		if functionName == "Len" || functionName =="Less" || functionName == "Swap" {
			funcStr += structType
		}

		baseFunctionDocList := make([]string, 0)

		if funcDecl.Body != nil {
			for _, stmt := range funcDecl.Body.List {
				baseFunctionDoc := inspectStmt(stmt, baseFunctionMap)
				if baseFunctionDoc != "" {
					baseFunctionDocList = append(baseFunctionDocList, baseFunctionDoc)
				}
			}
		}

		funcStr += strings.Join(baseFunctionDocList, "\n")
		functionList = append(functionList, funcStr)
	}
	// fmt.Println(functionList[1])
	return functionList
}

func inspectStmt(stmt ast.Stmt, baseFunctionMap map[string]string) string {
	var docstring string = ""
	ast.Inspect(stmt, func(n ast.Node) bool {

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			methodName := selExpr.Sel.Name
			if docs, ok := baseFunctionMap[methodName]; ok {
				docstring = docs
			}
		}

		return true
	})
	return docstring
}
