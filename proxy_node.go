package main

import (
    "net/http"
    "io"
    "fmt"
    "os"
    "bufio"
    "log"
    "strconv"
    "time"
)

type NodeInfo struct {
    Host string
    Port int
    Url string
    Id int
    IsLeader bool
}

type ProxyNode struct {
    // use a map for constant time lookup
    BlockedSites map[string]string
    Info NodeInfo
    PeerInfo []NodeInfo
}

type Message struct {
    Timestamp int64
    Data []byte
    SenderId int
}

func CreateProxyNode(nodes []NodeInfo, id int) *ProxyNode {
    rv := new(ProxyNode)
    rv.BlockedSites = make(map[string]string)

    for _, node_info := range nodes {
        if id == node_info.Id {
            rv.Info = node_info
        } else {
            rv.PeerInfo = append(rv.PeerInfo, node_info)
        }
    }

    return rv
}

func (p *ProxyNode) ReadConfig (path string) {
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

func (p *ProxyNode) StartServer() {
    http.HandleFunc("/", p.HandleRequest)
    http.HandleFunc("/proxy/", p.HandleProxyRequest)
    port := fmt.Sprintf(":%d", p.Info.Port)
    log.Fatal(http.ListenAndServe(port, nil))
}

func (p *ProxyNode) HandleRequest(w http.ResponseWriter, r *http.Request) {

    // forward request to the "child" node on port 8081
    if (p.Info.IsLeader) {
        request_path := fmt.Sprintf("http://localhost:8081/proxy/%s", r.URL.Path)
        new_request, err := http.NewRequest(r.Method, request_path, r.Body)
        new_request.Header = r.Header
        new_request.Host = r.Host
        client := &http.Client{}
        res, err := client.Do(new_request)
        for err != nil {
            log.Fatal(err)
            time.Sleep(1 * time.Second)
        }
        for key, slice := range res.Header {
            for _, val := range slice {
                w.Header().Add(key, val)
            }
        }
        _, err = io.Copy(w, res.Body)
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


func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

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

    n1 := NodeInfo {
        Host: "localhost",
        Port: 8080,
        Url: "localhost:8080",
        Id: 8080,
        IsLeader: true,
    }
    n2 := NodeInfo {
        Host: "localhost",
        Port: 8081,
        Url: "localhost:8081",
        Id: 8081,
        IsLeader: false,
    }

    nodes := []NodeInfo{n1, n2}

    p := CreateProxyNode(nodes, port)
    p.ReadConfig("blocked_sites.txt")
    p.StartServer()
}
