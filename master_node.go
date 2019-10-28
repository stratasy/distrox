package main

import (
    "net/http"
    "io"
    "fmt"
    "os"
    "bufio"
    "log"
	"strconv"
	"encoding/json"
	"io/ioutil"
)

// For holding subNode information from JSON
type SubNodes struct {
	Name string 
	Port string 
} 

type Proxy struct {
    // use a map for constant time lookup
    BlockedSites map[string]string
	port int

	// leader additions
	isMaster bool
	// subNodePorts [10]string // deafult 10 max
	// subNodeNames [10]string 
	subNodesArr [10]SubNodes
	subNodeCount int
	nodeIndex int
	
}

func CreateLeaderProxy(port int) *Proxy {
    rv := new(Proxy)
    rv.BlockedSites = make(map[string]string)
	rv.port = port
	rv.isMaster = true
	rv.nodeIndex = 0

    return rv
}

// Parses Config json file for node info
func (p *Proxy) SubNodeConfig(path string, numNodes int) {
	p.subNodeCount = numNodes

	if (!p.isMaster) {
		return
	} 

	b, err := ioutil.ReadFile(path)
    if err != nil {
        fmt.Print(err)
    }
	
	json.Unmarshal(b, &p.subNodesArr)
}

func (p *Proxy) MasterHandleRequest(w http.ResponseWriter, r *http.Request) {
	var currentNodePort = p.subNodesArr[p.nodeIndex].Port
	p.nodeIndex = (p.nodeIndex + 1) % p.subNodeCount
	log.Printf("Forwarding %s request to node at port %s\n", r.Method, currentNodePort)
	proxy_path := fmt.Sprintf("http://localhost:%s/proxy/%s", currentNodePort, r.URL.Host)

	_, err := http.Get(proxy_path)
        if err != nil {
            fmt.Println(err)
		}
}

func (p *Proxy) StartServer() {
    http.HandleFunc("/", p.MasterHandleRequest)
    port := fmt.Sprintf(":%d", p.port)
    http.ListenAndServe(port, nil)
}

////////////////////////////////////////////////////////////////
// Below are identical to proxy_node functions

func (p *Proxy) ReadConfig (path string) {
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

func (p *Proxy) HandleRequest(w http.ResponseWriter, r *http.Request) {

    // check if this site is blocked
    _, blocked := p.BlockedSites[r.URL.Host]
    if blocked {
        log.Println("Blocked site!")
        return
    }

    // format new request
    // Hardcoded to google, since input seems to be incorrectly parsed currently.
    request_path := "https://google.com"//fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.URL.Host, r.URL.Path)
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


func main() {

    args := os.Args
    if len(args) != 2 {
        fmt.Println("Arguments: [port]")
        return
    }

    port, err := strconv.Atoi(args[1])
    if err != nil {
        fmt.Println("[port] must be an integer!")
        return
    }

	p := CreateLeaderProxy(port)
	p.SubNodeConfig("subConfig.json", 3)

    p.ReadConfig("blocked_sites.txt")
    p.StartServer()
}