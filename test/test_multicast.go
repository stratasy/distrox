package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "hash/fnv"
    "log"
    "net"
    "os"
    "strconv"
    "time"
    "bytes"
    "io"
    "strings"
)

const (
    UNICAST_MESSAGE = 0
    MULTICAST_MESSAGE = 1
    JOIN_MESSAGE = 2
    JOIN_RESPONSE_MESSAGE = 3
)

type NodeInfo struct {
    Host string
    Port int
}

type Node struct {
    host      string
    id        int
    nodes     []NodeInfo
    listener  net.Listener
    peer_urls []string

    recent_message_hashes map[uint32]time.Time
}

type Message struct {
    Timestamp time.Time
    Data      []byte
    SenderId  int
    MessageType int
}

func CreateMessage(message []byte, sender_id int, message_type int) Message {
    rv := Message{
	Timestamp: time.Now(),
	SenderId:  sender_id,
	Data:      message,
	MessageType: message_type,
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

func RecentMessagesContains(hash uint32, node *Node) bool {
    _, ok := node.recent_message_hashes[hash]
    return ok
}

func Prune(node *Node) {
    now := time.Now()
    for key := range node.recent_message_hashes {
	if now.After(node.recent_message_hashes[key].Add(time.Duration(1.0 * time.Second))) {
	    delete(node.recent_message_hashes, key)
	}
    }
}

func InitConnection(nodes []NodeInfo, host string, port int) *Node {
    rv := &Node{}
    rv.id = port
    rv.host = host
    rv.nodes = nodes
    rv.recent_message_hashes = make(map[uint32]time.Time)

    url := fmt.Sprintf("%s:%d", host, port)
    log.Println(url)
    l, err := net.Listen("tcp", url)
    if err != nil {
	log.Fatal(err.Error())
    }
    rv.listener = l

    for _, node_info := range nodes {
	if rv.id == node_info.Port {
	    continue
	}

	url := fmt.Sprintf("%s:%d", node_info.Host, node_info.Port)
	rv.peer_urls = append(rv.peer_urls, url)
    }

    return rv
}

func (node *Node) ConstructUrlString() string {
    rv := fmt.Sprintf("%s:%d", node.host, node.id)
    for _, val := range node.peer_urls {
	rv += " "
	rv += val
    }
    return rv
}

func Contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func HandleRequest(buf []byte, node *Node) {

    message := ByteSliceToMessage(buf)
    hash_val := HashByteSlice(message.Data)
    Prune(node)

    found := RecentMessagesContains(hash_val, node)

    // new message that we haven't received yet!
    if !found && message.SenderId != node.id {

	node.recent_message_hashes[hash_val] = message.Timestamp
	if message.MessageType == MULTICAST_MESSAGE {
	    println(string(message.Data))
	    node.Multicast(buf)
	} else if message.MessageType == JOIN_MESSAGE {
	    remote_url := string(message.Data)
	    tokens := strings.Split(remote_url, ":")
	    port, _ := strconv.Atoi(tokens[1])
	    new_node := NodeInfo{
		Host: tokens[0],
		Port: port,
	    }
	    node.nodes = append(node.nodes, new_node)
	    //response_msg := CreateMessage([]byte(fmt.Sprintf("%s:%d", node.host, node.id)), node.id, JOIN_RESPONSE_MESSAGE)

	    node.peer_urls = append(node.peer_urls, remote_url)

	    data := node.ConstructUrlString()

	    msg := CreateMessage([]byte(data), node.id, JOIN_RESPONSE_MESSAGE)
	    node.Multicast(MessageToByteSlice(msg))

	} else if message.MessageType == UNICAST_MESSAGE {
	    println(string(message.Data))

	} else if message.MessageType == JOIN_RESPONSE_MESSAGE {
	    node.Multicast(buf)
	    peer_urls := strings.Split(string(message.Data), " ")
	    for _, url := range peer_urls {
		if !Contains(node.peer_urls, url) {
		    tokens := strings.Split(url, ":")
		    port, _ := strconv.Atoi(tokens[1])
		    new_node := NodeInfo{
			Host: tokens[0],
			Port: port,
		    }
		    node.nodes = append(node.nodes, new_node)
		    println(url)
		}
	    }
	    node.peer_urls = peer_urls
	}
    }
}

func HandleRequests(node *Node) {
    l := node.listener
    for {
	conn, err := l.Accept()
	if err != nil {
	    log.Fatal(err.Error())
	}
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	byte_data := buf.Bytes()
	byte_data = byte_data[:buf.Len()]
	go HandleRequest(byte_data, node)
	conn.Close()
    }
}

func HashByteSlice(message []byte) uint32 {
    h := fnv.New32a()
    h.Write(message)
    return h.Sum32()
}

func (node *Node) Multicast(message []byte) {
    for _, url := range node.peer_urls {
	conn, err := net.Dial("tcp", url)
	defer conn.Close()
	if err != nil {
	    log.Fatal(err.Error())
	}
	_, err = conn.Write(message)
	if err != nil {
	    log.Fatal(err.Error())
	}
    }
}

func (node *Node) Unicast(message []byte, whichOne int) {
    if (whichOne == node.id) {
	return;
    }
    for _, info := range node.nodes {
	if info.Port == whichOne {
	    url := fmt.Sprintf("%s:%d", info.Host, info.Port)
	    conn, err := net.Dial("tcp", url)
	    defer conn.Close()
	    if err != nil {
		log.Fatal(err.Error())
	    }
	    _, err = conn.Write(message)
	    if err != nil {
		log.Fatal(err.Error())
	    }
	}
    }
}

func (conn *Node) CloseConnection() {
    conn.listener.Close()
}

// testing
func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    n1 := NodeInfo{
	Host: "localhost",
	Port: 8080,
    }
    /*
    n2 := NodeInfo{
	Host: "localhost",
	Port: 8081,
    }
    n3 := NodeInfo{
	Host: "localhost",
	Port: 8082,
    }
    n4 := NodeInfo{
	Host: "localhost",
	Port: 8083,
    }
    */

    port, _ := strconv.Atoi(os.Args[1])

    nodes := []NodeInfo{n1}
    node := InitConnection(nodes, "localhost", port)

    defer node.CloseConnection()
    go HandleRequests(node)

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
	line := scanner.Text()
	tokens := strings.Split(line, " ")
	if tokens[0] == "connect" {
	    my_url := fmt.Sprintf("%s:%d", node.host, node.id)
	    msg := CreateMessage([]byte(my_url), node.id, JOIN_MESSAGE)
	    target_port, _ := strconv.Atoi(tokens[1])
	    bytes := MessageToByteSlice(msg)
	    node.Unicast(bytes, target_port)

	} else {
	    msg := CreateMessage([]byte(line), node.id, UNICAST_MESSAGE)
	    bytes := MessageToByteSlice(msg)
	    //node.Unicast(bytes, 8081)
	    node.Multicast(bytes)
	}
    }
}
