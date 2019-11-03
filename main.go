package main

import (
	"fmt"
	"github.com/drp6/distrox/proxy"
	"log"
	"os"
	"strconv"
)

func main() {
	// setup logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	args := os.Args
	if len(args) != 3 {
		fmt.Println("Arguments: [config] [id - has to match the \"id\" in the config file]")
		return
	}

	config := proxy.ReadConfig(args[1])

	id, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}

	p := proxy.CreateProxyNode(config.Nodes, id)
	p.ReadBlockedSites(config.BlockedSitesPath)
	p.StartServer()
}
