package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProxyConfig struct {
	LeaderId         int
	BlockedSitesPath string
	Nodes            []*NodeInfo
}

type NodeInfo struct {
	Host     string
	Port     int
	Url      string
	IsLeader bool
}

type ProxyNode struct {
	BlockedSites   map[string]string
	Info           *NodeInfo
	PeerInfo       []*NodeInfo
	SendingPeerIdx int
	//Cache          *LocalCache
	Messenger *TCPMessenger
	Responses map[string]HTTPResponse
	Lock      *sync.Mutex
	CV        *sync.Cond
	CurrentForwardingIdx  int
}

func CreateProxyNode(host string, port int, leader bool) *ProxyNode {
	rv := &ProxyNode{}
	rv.BlockedSites = make(map[string]string)

	rv.SendingPeerIdx = 0
	rv.Info = CreateNodeInfo(host, port, leader)
	rv.Messenger = InitTCPMessenger(rv.Info.Url)
	rv.Responses = make(map[string]HTTPResponse)

	rv.Lock = &sync.Mutex{}
	rv.CV = sync.NewCond(rv.Lock)
	return rv
}

func CreateNodeInfo(host string, port int, leader bool) *NodeInfo {
	rv := &NodeInfo{}
	rv.Host = host
	rv.Port = port
	rv.Url = fmt.Sprintf("%s:%d", host, port)
	rv.IsLeader = leader
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
		/*
			if node.Id == config.LeaderId {
			    node.IsLeader = true
			} else {
			    node.IsLeader = false
			}
		*/
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
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *ProxyNode) HandleHttpRequest(w http.ResponseWriter, r *http.Request) {
	// check if this site is blocked
	_, blocked := p.BlockedSites[r.Host]
	if blocked {
		log.Println("Blocked site!")
		fmt.Fprintf(w, "Site is blocked!\n")
		return
	}

	req := HTTPRequest{
		Method:        r.Method,
		RequestUrl:    fmt.Sprintf("%s%s", r.Host, r.URL.Path),
		Header:        r.Header,
		ContentLength: r.ContentLength,
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Panic(err)
	}
	req.Body = b

	req_bytes := HttpRequestToBytes(req)

	msg := CreateMessage(req_bytes, p.Info.Url, HTTP_REQUEST_MESSAGE)

	succeeded := false
	println(len(p.PeerInfo))
	for i:=0; i<len(p.PeerInfo); i++ {
	    p.CurrentForwardingIdx = (p.CurrentForwardingIdx + 1) % len(p.PeerInfo)
	    if p.Unicast(MessageToBytes(msg), p.PeerInfo[p.CurrentForwardingIdx].Url) {
		succeeded = true
		break
	    }
	}
	if !succeeded {
	    println("failed!")
	}

	p.Lock.Lock()
	for !p.ContainsResponse(req.RequestUrl) {
		p.CV.Wait()
	}
	res := p.Responses[req.RequestUrl]
	delete(p.Responses, req.RequestUrl)
	p.Lock.Unlock()

	for key, slice := range res.Header {
		for _, val := range slice {
			w.Header().Add(key, val)
		}
	}
	_, err = io.Copy(w, bytes.NewReader(res.Body))
	if err != nil {
		log.Panic(err)
	}
}

func (p *ProxyNode) HandleRequests() {
	if p.Info.IsLeader {
		go func() {
			http.HandleFunc("/", p.HandleHttpRequest)
			// TODO: remove hardcoding
			log.Fatal(http.ListenAndServe("localhost:8080", nil))
		}()
	}

	l := p.Messenger.Listener
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, conn)
		b := buf.Bytes()
		go p.HandleRequest(b)
		conn.Close()
	}
}

func (p *ProxyNode) HandleRequest(b []byte) {
	message := BytesToMessage(b)
	message_hash := HashBytes(b)

	p.Lock.Lock()
	p.Messenger.PruneStoredMessages()
	message_found := p.Messenger.HasMessageStored(message_hash)
	p.Lock.Unlock()

	if !message_found && message.SenderUrl != p.Info.Url {
		p.Lock.Lock()
		p.Messenger.RecentMessageHashes[message_hash] = message.Timestamp
		p.Lock.Unlock()

		if message.MessageType == MULTICAST_MESSAGE {
			println(string(message.Data))
			p.Multicast(b)
		} else if message.MessageType == JOIN_REQUEST_MESSAGE {
			m := string(message.Data)
			tokens := strings.Split(m, ":")
			port, _ := strconv.Atoi(tokens[1])
			new_node_info := CreateNodeInfo(tokens[0], port, false)
			log.Printf("New node joined with URL %s!", new_node_info.Url)

			if !p.ContainsUrl(new_node_info.Url) {
				p.PeerInfo = append(p.PeerInfo, new_node_info)
			}

			// notify all the nodes in the group of the new node joining
			msg := p.ConstructNodeJoinedMessage()
			p.Multicast(MessageToBytes(msg))

		} else if message.MessageType == JOIN_NOTIFY_MESSAGE {
			p.Multicast(b)
			peer_infos := strings.Split(string(message.Data), " ")
			for _, info := range peer_infos {
				tokens := strings.Split(info, ":")
				port, _ := strconv.Atoi(tokens[1])
				url := fmt.Sprintf("%s:%d", tokens[0], port)

				if url == p.Info.Url {
					continue
				}
				if !p.ContainsUrl(url) {
					new_node_info := CreateNodeInfo(tokens[0], port, false)
					p.PeerInfo = append(p.PeerInfo, new_node_info)
					log.Printf("New node joined with URL %s!", new_node_info.Url)
				}
			}
		} else if message.MessageType == LEAVE_NOTIFY_MESSAGE {
			url_to_remove := string(message.Data)
			log.Printf("Node has died with URL %s!", url_to_remove)
			p.RemoveNodeFromPeers(url_to_remove)
			p.Multicast(b)
		} else if message.MessageType == HTTP_REQUEST_MESSAGE {
			r := BytesToHttpRequest(message.Data)

			request_path := fmt.Sprintf("http://%s", r.RequestUrl)
			new_request, err := http.NewRequest(r.Method, request_path, bytes.NewReader(r.Body))

			log.Printf("Sending %s request to %s\n", r.Method, request_path)
			client := &http.Client{}
			res, err := client.Do(new_request)
			if err != nil {
				log.Panic(err)
			}
			defer res.Body.Close()

			body_bytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Panic(err)
			}

			res_to_send := HTTPResponse{
				Status:        res.Status,
				RequestUrl:    r.RequestUrl,
				Header:        res.Header,
				Body:          body_bytes,
				ContentLength: res.ContentLength,
			}

			bytes_to_send := HttpResponseToBytes(res_to_send)
			msg := CreateMessage(bytes_to_send, p.Info.Url, HTTP_RESPONSE_MESSAGE)
			p.Unicast(MessageToBytes(msg), "localhost:8081") // TODO: remove hardcoding
		} else if message.MessageType == HTTP_RESPONSE_MESSAGE {
			res := BytesToHttpResponse(message.Data)
			p.Lock.Lock()
			p.Responses[res.RequestUrl] = res
			p.Lock.Unlock()
			p.CV.Broadcast()
		} else if message.MessageType == UNICAST_MESSAGE {
			println(string(message.Data))
		}
	}
}

func (p *ProxyNode) Unicast(message []byte, url string) bool {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		// Unable to connect with the other node, that node must have died.
		p.RemoveNodeFromPeers(url)
		log.Printf("Node has died with URL %s!", url)

		msg := p.ConstructNodeLeftMessage(url)
		p.Multicast(MessageToBytes(msg))
		return false
	}

	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		log.Panic(err)
	}
	return true
}

func (p *ProxyNode) Multicast(message []byte) {
	succeeded := false
	for !succeeded {
		good := true
		for _, info := range p.PeerInfo {
			url := info.Url
			if !p.Unicast(message, url) {
				good = false
				break
			}
		}
		succeeded = good
	}
}

func (p *ProxyNode) ConstructNodeJoinedMessage() Message {
	rv := p.Info.Url
	for _, info := range p.PeerInfo {
		rv += " "
		rv += info.Url
	}
	msg := CreateMessage([]byte(rv), p.Info.Url, JOIN_NOTIFY_MESSAGE)
	return msg
}

func (p *ProxyNode) ConstructNodeLeftMessage(url string) Message {
	msg := CreateMessage([]byte(url), p.Info.Url, LEAVE_NOTIFY_MESSAGE)
	return msg
}

func (p *ProxyNode) ContainsUrl(url string) bool {
	for _, info := range p.PeerInfo {
		if url == info.Url {
			return true
		}
	}
	return false
}

func (p *ProxyNode) IndexFromString(url string) int {
	for i, info := range p.PeerInfo {
		if info.Url == url {
			return i
		}
	}
	return -1
}

func (p *ProxyNode) RemoveNodeFromPeers(url string) {
	idx := p.IndexFromString(url)
	if idx == -1 {
		return
	}
	p.PeerInfo[idx] = p.PeerInfo[len(p.PeerInfo)-1]
	p.PeerInfo[len(p.PeerInfo)-1] = nil
	p.PeerInfo = p.PeerInfo[:len(p.PeerInfo)-1]
}

func (p *ProxyNode) ContainsResponse(url string) bool {
	_, ok := p.Responses[url]
	return ok
}
