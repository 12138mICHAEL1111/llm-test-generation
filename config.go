package main

const (
	compilationBasePrompt string  = "The code you generated has compilation faults, fix them. "
	failedTestBasePrompt  string  = "The code you generated failed to pass test, fix them to pass the test. "
	chatGPTemp            float32 = 0.2
)

type filepathConfig struct {
	testGenerationBasePrompt string
	sourceFilePath           string
	testFilePath             string
	errorFilePath            string
}

var gonumConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for function {functionName}, the test function should be in a new test file but in the same package. The package name is floats. DO NOT return tested function code to me. ",
	sourceFilePath:           "/Users/maike/Desktop/gonum/floats/floats.go",
	testFilePath:             "/Users/maike/Desktop/gonum/floats/floats_test.go",
	errorFilePath:            "/Users/maike/Desktop/gonum/error.json",
}

var boltConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for function {functionName}, the test function should be in a new test file and in different package. The tested code file package name is bolt and the package path is github.com/boltdb/bolt , the test file package name shoule be bolt_test. DO NOT return tested function code to me. ",
	sourceFilePath:           "/Users/maike/go/src/github.com/boltdb/bolt/db.go",
	testFilePath:             "/Users/maike/go/src/github.com/boltdb/bolt/db_test.go",
	errorFilePath:            "/Users/maike/go/src/github.com/boltdb/bolt/error.json",
}

var fastjsonConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for function {functionName}, the test function should be in a new test file but in the same package. The package name is fastjson. DO NOT return tested function code to me. ",
	sourceFilePath:           "/Users/maike/Desktop/fastjson/parser.go",
	testFilePath:             "/Users/maike/Desktop/fastjson/parser_test.go",
	errorFilePath:            "/Users/maike/Desktop/fastjson/error.json",
}