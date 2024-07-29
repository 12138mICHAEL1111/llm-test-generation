package main

const (
	compilationBasePrompt string  = "The code you generated has compilation faults, fix them. "
	failedTestBasePrompt  string  = "The code you generated failed to pass test, fix them to pass the test. "
	temp                  float32 = 1.0
)

type filepathConfig struct {
	testGenerationBasePrompt string
	sourceFilePath           string
	testFilePath             string
	errorFilePath            string
}

var gonumConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for function {functionName}, the test function should be in a new test file but in the same package. The test package name and source code package name are both called floats, so no need to import the source file package. Do not include any tested funtion code in your completion \n",
	sourceFilePath:           "/Users/maike/Desktop/gonum/floats/floats.go",
	testFilePath:             "/Users/maike/Desktop/gonum/floats/floats_test.go",
	errorFilePath:            "/Users/maike/Desktop/gonum/error.json",
}

var boltdbConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for function {functionName}, the test function should be in a new test file and in different package. The tested code file package name is bolt and the package path is github.com/boltdb/bolt, the test file package name shoule be bolt_test. DO NOT include any source function code in your completion. If bolt.Open method is called, it will create a temp file which path is the first parameter of the function, use defer to delete any temp file created BEFORE bolt.Open statement\n",
	sourceFilePath:           "/Users/maike/go/src/github.com/boltdb/bolt/db.go",
	testFilePath:             "/Users/maike/go/src/github.com/boltdb/bolt/db_test.go",
	errorFilePath:            "/Users/maike/go/src/github.com/boltdb/bolt/error.json",
}

var fastjsonConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for function {functionName}, the test function should be in a new test file but in the same package. The test package name and source code package name are both called fastjson, so no need to import the source file package. Do not include any tested funtion code in your completion\n",
	sourceFilePath:           "/Users/maike/Desktop/fastjson/parser.go",
	testFilePath:             "/Users/maike/Desktop/fastjson/parser_test.go",
	errorFilePath:            "/Users/maike/Desktop/fastjson/error.json",
}
