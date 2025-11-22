package main

// main.go is the entry point for the mangahub CLI application.
import (
	cmd "mangahub/cmd/cli/command"
)

func main() {
	cmd.Execute()
}
