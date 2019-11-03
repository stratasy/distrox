package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

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

func (p *ProxyNode) StartServer() {
	http.HandleFunc("/", p.HandleRequest)
	http.HandleFunc("/proxy/", p.HandleProxyRequest)
	port := fmt.Sprintf(":%d", p.Info.Port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (p *ProxyNode) ForwardRequest(w http.ResponseWriter, r *http.Request, peer_id int) bool {
	// default to ourselves
	url := p.Info.Url

	if peer_id != -1 {
		// figure out which node to send it to (round robin for now)
		url = p.PeerInfo[peer_id].Url

	}

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
	return true
}

func (p *ProxyNode) HandleRequest(w http.ResponseWriter, r *http.Request) {

	// if we are the leader, forward the response to a child node!
	if p.Info.IsLeader {

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

/*
func main() {
    // setup logger
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    args := os.Args
    if len(args) != 3 {
        fmt.Println("Arguments: [config] [id - has to match the \"id\" in the config file]")
        return
    }

    config := ReadConfig(args[1])

    id, err := strconv.Atoi(args[2])
    if err != nil {
        log.Fatal(err)
    }

    p := CreateProxyNode(config.Nodes, id)
    p.ReadBlockedSites(config.BlockedSitesPath)
    p.StartServer()
}
*/
