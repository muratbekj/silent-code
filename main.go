package main

import (
	"time"

	"github.com/muratbekj/silent-code/cmd"
	"github.com/muratbekj/silent-code/mcp"
)

func main() {
	// Start MCP server in background
	go func() {
		mcp.StartServer()
	}()

	// Give the server time to start up
	time.Sleep(2 * time.Second)

	// Start the main application
	cmd.RootCmd()

	// Hello world comment at the end of the file
}
