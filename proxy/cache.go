package proxy

import (
	"net/http"
	"net/url"
	"time"
)

type CacheMetadata struct {
	Res       *http.Response
	SavedTime time.Time
}

type LocalCache struct {
	// Note: Will need a mutex if LocalCache is accessed by multiple servers/threads.
	Mem map[url.URL]CacheMetadata
}

// Create and initialize a new LocalCache, and return a pointer to it.
func CreateLocalCache() *LocalCache {
	return &LocalCache{
		Mem: make(map[url.URL]CacheMetadata),
	}
}

// Returns a []byte representing the cache'd body, if it exists.
// If the body has expired, or does not exist, then the entry is deleted
// from the cache and nil is returned.
func (cache *LocalCache) CacheGet(pageURL url.URL) *http.Response {
	cacheData := cache.Mem[pageURL]
	println(time.Now().String())
	println(cacheData.SavedTime.String())
	if cacheData.SavedTime.IsZero() || time.Now().After(cacheData.SavedTime) {
		delete(cache.Mem, pageURL)
		return nil
	}
	return cacheData.Res
}

// If you need to convert a string into a URL struct, use: func Parse(rawurl string) (*URL, error)
// https://golang.org/pkg/net/url/#Parse
//
// Attempt to cache the body of a response in a LocalCache.
// Takes a url, the body to store for that url, and a duration to store it for.
// Returns 1 on success, 0 on failure.
func (cache *LocalCache) CacheSet(pageURL url.URL, Res *http.Response, storeDuration time.Duration) int {
	cache.Mem[pageURL] = CacheMetadata{
		Res:       Res,
		SavedTime: time.Now().Add(storeDuration * time.Second),
	}
	return 1
}

/*
func main() {
	cache := CreateLocalCache()
	u, _ := url.Parse("http://www.google.com")

	s := "http://www.gmail.com"
	res, _ := http.Get(s)

	cache.CacheSet(*u, res.Body, time.Duration(0.0*10e9))
	println(string(cache.CacheGet(*u)))
}
*/
