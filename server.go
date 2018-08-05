package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/qbradq/sharduo/common"

	"github.com/qbradq/sharduo/accounting"
	"github.com/qbradq/sharduo/network"
)

func main() {
	wg := new(sync.WaitGroup)
	defer wg.Wait()

	common.Config = common.NewConfigurationFromFile("data/config.txt")

	go accounting.Service(wg)
	defer accounting.Stop()

	login := network.NewPacketServer(
		common.Config.GetString("network.internalIP", "127.0.0.1"),
		common.Config.GetInt("network.port", 2593),
	)
	go login.Run(wg)
	defer login.Stop()

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
			break
		case "default":
			fmt.Println("Unknown command", command)
		}
	}
}
