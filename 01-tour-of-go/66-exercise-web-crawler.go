package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type SafeCache struct {
	mu      sync.Mutex
	visited map[string]bool
}

// Shared cache of already visited URLs.
// Protected by a mutex because multiple goroutines
// may read/write it simultaneously.
var cache = SafeCache{
	visited: make(map[string]bool),
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, done chan bool) {
	// Notify parent Crawl that this Crawl has finished.
	defer func() {
		done <- true
	}()

	if depth <= 0 {
		return
	}

	// Check and mark URL as visited atomically.
	cache.mu.Lock()
	if cache.visited[url] {
		cache.mu.Unlock()
		return
	}
	cache.visited[url] = true
	cache.mu.Unlock()

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)

	// Used to wait for all child crawls spawned
	// from the current page.
	children := make(chan bool)

	count := 0
	// Spawn child crawls concurrently.
	for _, u := range urls {
		count++
		go Crawl(u, depth-1, fetcher, children)
	}

	// Wait until every child crawl signals completion.
	for i := 0; i < count; i++ {
		<-children
	}
}

func main() {
	done := make(chan bool)

	go Crawl("https://golang.org/", 4, fetcher, done)

	<-done
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
