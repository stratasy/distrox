package main

import (
	"bufio"
	"github.com/drp6/distrox/proxy"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	// setup logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	args := os.Args
	host := args[1]
	port, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}
	is_leader, err := strconv.ParseBool(args[3])
	if err != nil {
		log.Fatal(err)
	}

	p := proxy.CreateProxyNode(host, port, is_leader, "config.json")
	go p.HandleRequests()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, " ")
		if tokens[0] == "connect" {
			message := p.Info.Url
			msg := proxy.CreateMessage([]byte(message), p.Info.Url, proxy.JOIN_REQUEST_MESSAGE)
			bytes := proxy.MessageToBytes(msg)
			p.Unicast(bytes, tokens[1])
		} else {
			msg := proxy.CreateMessage([]byte(line), p.Info.Url, proxy.UNICAST_MESSAGE)
			bytes := proxy.MessageToBytes(msg)
			p.Multicast(bytes)
		}
	}
}
