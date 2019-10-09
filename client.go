package main

import (
    "net"
    "fmt"
    "bufio"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        return;
    }
    fmt.Fprintf(conn, "The client is sending a message!\n")
    status, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        return
    }
    fmt.Printf(status)
}
