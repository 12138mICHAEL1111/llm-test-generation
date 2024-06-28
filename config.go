package main

const (
	secondCompilationBasePrompt string = "The code you generated has compilation faults, fix them."
	thirdCompilationBasePrompt  string = "The code you tried to fix still has compilation faults, fix them again"
	secondFailedTestBasePrompt  string = "The code you generated failed to pass test, fix them to pass the test."
	thirdFailedTestBasePrompt   string = "The code you generated still failed to pass test, fix them again to pass the test."
)

type filepathConfig struct {
	testGenerationBasePrompt string
	sourceFilePath           string
	testFilePath             string
}

var gonumConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for the following golang function, the test function should be in a new test file but in the same package",
	sourceFilePath:           "/Users/maike/Desktop/gonum/floats/floats.go",
	testFilePath:             "/Users/maike/Desktop/gonum/floats/floats_test.go",
}

var boltConfig = filepathConfig{
	testGenerationBasePrompt: "generate test function for the following golang function, the test function should be in a new test file and in different package. The tested code file package name is bolt and the package path is github.com/boltdb/bolt , the test file package name shoule be bolt_test.",
	sourceFilePath:           "/Users/maike/go/src/github.com/boltdb/bolt/db.go",
	testFilePath:             "/Users/maike/go/src/github.com/boltdb/bolt/db_test.go",
}
