package proxy

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type ProxyConfig struct {
	PublicUrl    string `json:"public_url"`
	CacheTimeout int    `json:"cache_timeout"`

	BlockedSitesList []string `json:"blocked_sites"`
	BlockedSites     map[string]string
}

func LoadProxyConfig(path string) *ProxyConfig {
	config := &ProxyConfig{}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal([]byte(file), &config)
	if err != nil {
		log.Fatal(err)
	}

	config.BlockedSites = make(map[string]string)
	for _, val := range config.BlockedSitesList {
		config.BlockedSites[val] = val
	}

	return config
}

func (p *ProxyConfig) SiteIsBlocked(url string) bool {
	_, blocked := p.BlockedSites[url]
	return blocked
}
