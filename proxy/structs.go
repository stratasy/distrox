package proxy

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
}
