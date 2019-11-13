package proxy

import (
    "bufio"
    "encoding/json"
    "fmt"
    "hash/fnv"
    "io"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "os"
    "time"
    "bytes"
    "net/url"
)

const (
    HTTP_REQUEST_MESSAGE = 0
    HTTP_RESPONSE_MESSAGE = 0
    MULTICAST_MESSAGE = 1
)

type ProxyConfig struct {
    LeaderId         int         `json:"leader_id"`
    BlockedSitesPath string      `json:"blocked_sites_path"`
    Nodes            []*NodeInfo `json:"nodes"`
}

type NodeInfo struct {
    Host     string `json:"host"`
    Port     int    `"json:"port"`
    Url      string
    Id       int `"json:"id"`
    IsLeader bool
}

type ProxyNode struct {
    BlockedSites   map[string]string
    Info           *NodeInfo
    PeerInfo       []*NodeInfo
    SendingPeerIdx int
    Cache          *LocalCache

    Listener net.Listener
    RecentMessageHashes map[uint32]time.Time
}

type Message struct {
    Timestamp time.Time
    Method string
    Url url.URL
    SenderId int
    MessageType int
    Hash uint32
}

type ResponseMessage struct {
    Timestamp time.Time
    Response []byte
}

func CreateMessage(request http.Request, sender_id int, message_type int) Message {
    rv := Message{
	Timestamp: time.Now(),
	SenderId:  sender_id,
	Url: *request.URL,
	Method: request.Method,
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

func RecentMessagesContains(hash uint32, node *ProxyNode) bool {
    _, ok := node.RecentMessageHashes[hash]
    return ok
}

func Prune(node *ProxyNode) {
    now := time.Now()
    for key := range node.RecentMessageHashes{
	if now.After(node.RecentMessageHashes[key].Add(time.Duration(1.0 * time.Second))) {
	    delete(node.RecentMessageHashes, key)
	}
    }
}

func CreateProxyNode(nodes []*NodeInfo, id int) *ProxyNode {
    rv := new(ProxyNode)
    rv.BlockedSites = make(map[string]string)

    for _, node_info := range nodes {
	if id == node_info.Id {
	    rv.Info = node_info
	} else {
	    rv.PeerInfo = append(rv.PeerInfo, node_info)
	}
    }

    rv.SendingPeerIdx = 0
    rv.Cache = CreateLocalCache()
    return rv
}

func ReadConfig(path string) ProxyConfig {
    var config ProxyConfig
    file, err := ioutil.ReadFile(path)
    if err != nil {
	log.Fatal(err)
    }

    err = json.Unmarshal([]byte(file), &config)
    if err != nil {
	log.Fatal(err)
    }

    for _, node := range config.Nodes {
	node.Url = fmt.Sprintf("%s:%d", node.Host, node.Port)
	if node.Id == config.LeaderId {
	    node.IsLeader = true
	} else {
	    node.IsLeader = false
	}
    }
    return config
}

func (p *ProxyNode) ReadBlockedSites(path string) {
    file, err := os.Open(path)
    if err != nil {
	log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
	site := scanner.Text()
	p.BlockedSites[site] = site
    }
}

func HashByteSlice(data string) uint32 {
    h := fnv.New32a()
    h.Write([]byte(data))
    return h.Sum32()
}

func (p *ProxyNode) HandleMessage(buf []byte) {
    message := ByteSliceToMessage(buf)
    hash_val := HashByteSlice(message.Url.String())
    Prune(p)

    found := RecentMessagesContains(hash_val, p)

    // new message that we haven't received yet!
    if !found && message.SenderId != p.Info.Id {
	if message.MessageType == HTTP_REQUEST_MESSAGE {
	    p.RecentMessageHashes[hash_val] = message.Timestamp
	    println(message.Url.String())
	    res, err := http.Get(message.Url.String())
	    if err != nil {
		log.Panic(err)
	    }
	    body_bytes, err := ioutil.ReadAll(res.Body)
	    if err != nil {
		log.Panic(err)
	    }
	    p.Unicast(body_bytes, 0)
	}
	else if message.MessageType == HTTP_RESPONSE_MESSAGE {

	}
    }
}

func (p *ProxyNode) Unicast(message []byte, whichOne int) {
    if (whichOne == p.Info.Id) {
        return;
    }
    for _, info := range p.PeerInfo{
        if info.Id == whichOne {
            url := fmt.Sprintf("%s:%d", info.Host, info.Port)
	    println(url)
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


func (p *ProxyNode) StartServer() {

    // if this node is the leader, listen for incoming HTTP requests
    if p.Info.IsLeader {
	go func () {
	    http.HandleFunc("/", p.HandleRequest)
	    http.HandleFunc("/proxy/", p.HandleProxyRequest)
	    port := fmt.Sprintf(":%d", 8080) //TODO: Fix hardcoding
	    log.Fatal(http.ListenAndServe(port, nil))
	}()
    }

    println(p.Info.Url)
    rv, err := net.Listen("tcp", p.Info.Url)
    if err != nil {
	log.Fatal(err)
    }
    p.Listener = rv
    p.RecentMessageHashes = make(map[uint32]time.Time)

    for {
	conn, err := p.Listener.Accept()
	defer conn.Close()
	if err != nil {
	    log.Fatal(err.Error())
	}
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	byte_data := buf.Bytes()
	go p.HandleMessage(byte_data)
    }
}

func (p *ProxyNode) ForwardRequest(w http.ResponseWriter, r *http.Request, peer_id int) bool {
    // default to ourselves
    //url := p.Info.Url
    url := "localhost:8080"

    if peer_id != -1 {
	// figure out which node to send it to (round robin for now)
	url = p.PeerInfo[peer_id].Url
    }
    println(url)
    message := CreateMessage(*r, p.Info.Id, HTTP_REQUEST_MESSAGE)
    p.Unicast(MessageToByteSlice(message), p.PeerInfo[0].Id) //Hardcoded!

    /*
    // make a copy of the request and send it to the corresponding child
    request_path := fmt.Sprintf("http://%s/proxy%s", url, r.URL.Path)
    new_request, err := http.NewRequest(r.Method, request_path, r.Body)
    new_request.Header = r.Header
    new_request.Host = r.Host
    client := &http.Client{}

    // copy back the response from the child
    res, err := client.Do(new_request)
    if err != nil {
	// couldn't connect to the child :(
	return false
    }
    for key, slice := range res.Header {
	for _, val := range slice {
	    w.Header().Add(key, val)
	}
    }
    io.Copy(w, res.Body)
    //p.Cache.CacheSet(*r.URL, res, time.Duration(5.0))
    return true
    */

    return true
}

func (p *ProxyNode) HandleRequest(w http.ResponseWriter, r *http.Request) {

    // if we are the leader, forward the response to a child node!
    if p.Info.IsLeader {

	/*
	res := p.Cache.CacheGet(*r.URL)
	if res != nil {
	    log.Printf("Cached request: %s\n", r.URL.Host)
	    for key, slice := range res.Header {
		for _, val := range slice {
		    w.Header().Add(key, val)
		}
	    }
	    io.Copy(w, res.Body)
	    return
	}
	*/

	// try forwarding the response a child node
	for i := 0; i < len(p.PeerInfo); i++ {
	    if p.ForwardRequest(w, r, p.SendingPeerIdx) {
		p.SendingPeerIdx = (p.SendingPeerIdx + 1) % len(p.PeerInfo)
		return
	    }
	    // update the round-robin counter
	    p.SendingPeerIdx = (p.SendingPeerIdx + 1) % len(p.PeerInfo)
	}

	// if we reach here, we were unable to connect to any child nodes, so we'll just send the request ourselves
	p.ForwardRequest(w, r, -1)
    }
}

func (p *ProxyNode) HandleProxyRequest(w http.ResponseWriter, r *http.Request) {

    // check if this site is blocked
    _, blocked := p.BlockedSites[r.Host]
    if blocked {
	log.Println("Blocked site!")
	fmt.Fprintf(w, "Site is blocked!\n")
	return
    }

    // format new request
    request_path := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path[len("/proxy"):])
    // create new HTTP request with the target URL (everything else is the same)
    new_request, err := http.NewRequest(r.Method, request_path, r.Body)

    // send request to server
    log.Printf("Sending %s request to %s\n", r.Method, request_path)
    client := &http.Client{}
    res, err := client.Do(new_request)
    if err != nil {
	log.Panic(err)
    }
    defer res.Body.Close()

    // copy the headers over to the ResponseWriter.
    //res.Header is a map of string -> slice (string)
    for key, slice := range res.Header {
	for _, val := range slice {
	    w.Header().Add(key, val)
	}
    }

    // forward response to client
    _, err = io.Copy(w, res.Body)
    if err != nil {
	log.Panic(err)
    }
}
