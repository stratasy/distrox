package proxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type cacheMetaData struct {
	urlBody   []byte
	savedTime time.Time
}

type localCache struct {
	// Note: Will need a mutex if localCache is accessed by multiple servers/threads.
	mem map[url.URL]cacheMetaData
}

// Create and initialize a new localCache, and return a pointer to it.
func newMem() *localCache {
	return &localCache{
		mem: make(map[url.URL]cacheMetaData),
	}
}

// Returns a []byte representing the cache'd body, if it exists.
// If the body has expired, or does not exist, then the entry is deleted
// from the cache and nil is returned.
func (lc *localCache) CacheGet(pageURL url.URL) []byte {
	cacheData := lc.mem[pageURL]
	if cacheData.savedTime.IsZero() || time.Now().After(cacheData.savedTime) {
		delete(lc.mem, pageURL)
		return nil
	}
	return cacheData.urlBody
}

// If you need to convert a string into a URL struct, use: func Parse(rawurl string) (*URL, error)
// https://golang.org/pkg/net/url/#Parse
//
// Attempt to cache the body of a response in a localCache.
// Takes a url, the body to store for that url, and a duration to store it for.
// Returns 1 on success, 0 on failure.
func (lc *localCache) CacheSet(pageURL url.URL, bodyReader io.Reader, storeDuration time.Duration) int {
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	lc.mem[pageURL] = cacheMetaData{
		urlBody:   body,
		savedTime: time.Now().Add(storeDuration),
	}
	return 1
}

func main() {
	cache := newMem()
	u, _ := url.Parse("http://www.google.com")

	s := "http://www.gmail.com"
	res, _ := http.Get(s)

	cache.CacheSet(*u, res.Body, time.Duration(0.0*10e9))
	println(string(cache.CacheGet(*u)))
}
