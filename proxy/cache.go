package proxy

import (
	"net/http"
	"time"
)

type CacheMetadata struct {
	// Looks like I only need header and body.
	// Can just make two seperate items, header and body
	// in here, which should allow for easy incorporation.
	Header    http.Header
	Body      []byte
	SavedTime time.Time
}

type LocalCache struct {
	// Note: Will need a mutex if LocalCache is accessed by multiple servers/threads.
	Mem map[string]CacheMetadata
}

// Create and initialize a new LocalCache, and return a pointer to it.
func CreateLocalCache() *LocalCache {
	return &LocalCache{
		Mem: make(map[string]CacheMetadata),
	}
}

// Returns a []byte representing the cache'd body, if it exists.
// If the body has expired, or does not exist, then the entry is deleted
// from the cache and nil is returned.
func (cache *LocalCache) CacheGet(pageURL string) *CacheMetadata {
	cacheData := cache.Mem[pageURL]
	if cacheData.SavedTime.IsZero() || time.Now().After(cacheData.SavedTime) {
		delete(cache.Mem, pageURL)
		return nil
	}
	return &cacheData
}

// If you need to convert a string into a URL struct, use: func Parse(rawurl string) (*URL, error)
// https://golang.org/pkg/net/url/#Parse
//
// Attempt to cache the body of a response in a LocalCache.
// Takes a url, the body to store for that url, and a duration to store it for.
// Returns 1 on success, 0 on failure.
func (cache *LocalCache) CacheSet(pageURL string, Res HTTPResponse, secondsToStore int) int {
	storeDuration := time.Duration(secondsToStore)
	cache.Mem[pageURL] = CacheMetadata{
		Header:    Res.Header,
		Body:      Res.Body,
		SavedTime: time.Now().Add(storeDuration * time.Second),
	}
	return 1
}

/*func main() {
	cache := CreateLocalCache()

	s := "http://www.gmail.com"
	res, _ := http.Get(s)

	cache.CacheSet("http://www.google.com", res, 5)
	_, b := cache.CacheGet("http://www.google.com")
	println(string(*b))
}*/
