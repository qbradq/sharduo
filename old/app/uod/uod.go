package uod

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/qbradq/sharduo/internal/accounting"
	"github.com/qbradq/sharduo/internal/world"
)

// Main is the uod main function
func Main() {
	// Global setup
	rand.Seed(time.Now().Unix())

	// Start services
	go accounting.Start()
	defer accounting.Stop()
	go world.Start()
	defer world.Stop()
	go netStart()
	defer netStop()

	fmt.Println("ShardUO Root Console")
	fmt.Println("--------------------------------------------------------------------------------")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		switch command {
		case "abort":
			// This may not be printed before process termination
			fmt.Println("Server abort requested from the root conosle")
			os.Exit(1)
		case "help":
			fmt.Println("abort       Immediately aborts the server process")
			fmt.Println("help        Prints this help text")
			fmt.Println("quit        Stops the server gracefully")
		case "quit":
			fmt.Println("Server shutdown requested from the root console")
			return
		case "default":
			fmt.Println("Unknown command", command)
		}
	}
}
