package main

const(
	testGenerationBasePrompt string = "generate test function for the following golang function, the test function should be in a new test file but in the same package"
	secondCompilationBasePrompt string = "The code you generated has compilation faults, fix them."
	thirdCompilationBasePrompt string = "The code you tried to fix still has compilation faults, fix them again"
	secondFailedTestBasePrompt string = "The code you generated failed to pass test, fix them to pass the test."
	thirdFailedTestBasePrompt string = "The code you generated still failed to pass test, fix them again to pass the test."
	sourceFilePath string = "/Users/maike/Desktop/gonum/floats/floats.go"
)