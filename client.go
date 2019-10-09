package main

import (
    "net"
    "fmt"
    "bufio"
)

func main() {
    // connect to localhost:8080
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        return;
    }
    // send a message
    fmt.Fprintf(conn, "The client is sending a message!\n")

    // and print the response
    status, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        return
    }
    fmt.Printf(status)
}
