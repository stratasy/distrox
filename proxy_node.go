package main

import (
    "net/http"
    "io"
    //"io/ioutil"
    "fmt"
    "os"
    "bufio"
    "log"
    "strconv"
)

type Proxy struct {
    // use a map for constant time lookup
    BlockedSites map[string]string
    port int
}

func CreateProxy(port int) *Proxy {
    rv := new(Proxy)
    rv.BlockedSites = make(map[string]string)
    rv.port = port
    return rv
}

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

func (p *Proxy) StartServer() {
    http.HandleFunc("/", p.HandleRequest)
    port := fmt.Sprintf(":%d", p.port)
    http.ListenAndServe(port, nil)
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

    p := CreateProxy(port)
    p.ReadConfig("blocked_sites.txt")
    p.StartServer()
}
