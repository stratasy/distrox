package main

import (
    "fmt"
    "net"
    "log"
    "strconv"
    "os"
    "hash/fnv"
    "bufio"
    "time"
    "encoding/json"
)

type NodeInfo struct {
    Host string
    Port int
}

type RMulticastConnection struct {
    host string
    id int
    nodes []NodeInfo
    listener net.Listener
    peer_urls []string

    last_message_hash uint32
}

type Message struct {
    Timestamp time.Time
    Data []byte
    SenderId int
}

func CreateMessage(message []byte, sender_id int) Message {
    rv := Message{
        Timestamp:  time.Now(),
        SenderId: sender_id,
        Data: message,
    }
    return rv
}

func MessageToByteSlice(message Message) []byte {
    b, err := json.Marshal(message)
    if err != nil {
        log.Fatal(err)
    }
    return b
}

func ByteSliceToMessage(input []byte) Message {
    rv := Message{}
    json.Unmarshal(input, &rv)
    return rv
}

func InitConnection(nodes []NodeInfo, host string, port int) *RMulticastConnection {
    rv := &RMulticastConnection{}
    rv.id = port
    rv.host = host
    rv.nodes = nodes
    rv.last_message_hash = 0

    url := fmt.Sprintf("%s:%d", host, port)
    log.Println(url)
    l, err := net.Listen("tcp", url)
    if err != nil {
	log.Fatal(err.Error())
    }
    rv.listener = l

    for _, node_info := range nodes {
	if (rv.id == node_info.Port) {
	    continue
	}

	url := fmt.Sprintf("%s:%d", node_info.Host, node_info.Port)
        rv.peer_urls = append(rv.peer_urls, url)
    }

    return rv
}

func HandleRequest(buf []byte, rm_conn *RMulticastConnection) {

    message := ByteSliceToMessage(buf)
    hash_val := HashByteSlice(buf)

    // new message that we haven't received yet!
    if hash_val != rm_conn.last_message_hash && message.SenderId != rm_conn.id {
        rm_conn.last_message_hash = hash_val
        println(string(message.Data))
        rm_conn.Multicast(buf)
    }
}

func HandleRequests(rm_conn *RMulticastConnection) {
    l := rm_conn.listener
    for {
        conn, err := l.Accept()
        if err != nil {
            log.Fatal(err.Error())
        }
        buf := make([]byte, 4096)
        n, err := conn.Read(buf)
        if err != nil {
	    log.Fatal(err.Error())
        }
        buf = buf[:n]
	go HandleRequest(buf, rm_conn)
        conn.Close()
    }
}

func HashByteSlice(message []byte) uint32 {
    h := fnv.New32a()
    h.Write(message)
    return h.Sum32()
}

func (rm_conn *RMulticastConnection) Multicast(message []byte) {
    for _, url := range rm_conn.peer_urls{
	conn, err := net.Dial("tcp", url)
        if err != nil {
            log.Fatal(err.Error())
        }
        _, err = conn.Write(message)
        if err != nil {
            log.Fatal(err.Error())
        }
    }
}

func (conn *RMulticastConnection) CloseConnection() {
    conn.listener.Close()
}

// testing
func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    n1 := NodeInfo {
	Host: "localhost",
	Port: 8080,
    }
    n2 := NodeInfo {
	Host: "localhost",
	Port: 8081,
    }
    /*
    n3 := NodeInfo {
	Host: "localhost",
	Port: 8082,
    }
    n4 := NodeInfo {
	Host: "localhost",
	Port: 8083,
    }
    */

    port, _ := strconv.Atoi(os.Args[1])

    nodes := []NodeInfo{n1, n2}
    rm_conn := InitConnection(nodes, "localhost", port)

    defer rm_conn.CloseConnection()
    go HandleRequests(rm_conn)

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
        msg := CreateMessage([]byte(line), rm_conn.id)
        bytes := MessageToByteSlice(msg)
        rm_conn.Multicast(bytes)
    }
}
