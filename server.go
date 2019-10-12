package main

import (
    "net"
    "bufio"
    "fmt"
)

func handleConnection(conn net.Conn) {
    message, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        return
    }
    conn.Write([]byte(fmt.Sprintf("Message got: %s\n", message)))
}

func main() {
    // startup server on localhost:8080
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        return;
    }
    // infinite loop
    for {
        // wait for clients to connect
	conn, err := ln.Accept()
	if err != nil {
            return;
	}

        // and when one connects, asynchronously go send a response
	go handleConnection(conn)
    }
}
