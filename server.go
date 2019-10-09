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
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        return;
    }
    for {
	conn, err := ln.Accept()
	if err != nil {
            return;
	}
	go handleConnection(conn)
    }
}
