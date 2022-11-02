package config

import "os"

var EnvironmentVariables EnvironmentVars

func ReadEnvironmentVars() {
	// Read environment variables
	HelloWorld := os.Getenv("HELLO_WORLD")
	// Set environment variables
	EnvironmentVariables.HelloWorld = HelloWorld
}
