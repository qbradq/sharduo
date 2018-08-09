package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/common"
	"github.com/qbradq/sharduo/internal/world"

	"github.com/qbradq/sharduo/internal/accounting"
	"github.com/qbradq/sharduo/internal/network"
)

func main() {
	wg := new(sync.WaitGroup)
	defer wg.Wait()

	// Global setup
	rand.Seed(time.Now().Unix())

	// Load static data
	common.Config = common.NewConfigurationFromFile("data/config.txt")

	// Spin up all required services. This also starts all fixed instances.
	go accounting.Service(wg)
	defer accounting.Stop()
	go world.Service(wg)
	defer world.Stop()

	// Start network services
	server := network.NewPacketServer(
		common.Config.GetString("network.internalIP", "127.0.0.1"),
		common.Config.GetInt("network.port", 2593),
	)
	go server.Run(wg)
	defer server.Stop()

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
