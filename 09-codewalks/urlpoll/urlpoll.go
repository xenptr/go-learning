// Package main polls a list of URLs and tracks their status using goroutines
// and channels — demonstrating Go's "share memory by communicating" philosophy.
//
// Source: https://go.dev/doc/codewalk/sharemem/
//
// ── What this codewalk teaches ──────────────────────────────────────────────
//
// The traditional approach to shared state in concurrent programs:
//   - Keep a shared map of URL → status
//   - Protect it with a mutex (Lock/Unlock around every read and write)
//
// Go's alternative: DON'T share memory. Instead, communicate ownership
// of data through channels. Only one goroutine touches a value at a time.
//
//   "Do not communicate by sharing memory;
//    instead, share memory by communicating." — Go proverb
//
// Architecture of this program:
//
//   pending channel ──► Poller goroutines ──► complete channel
//        ▲                    │                      │
//        │                    ▼                      │
//        │              status channel               │
//        │                    │                      │
//        │             StateMonitor goroutine        │
//        │                                           │
//        └─────────── Resource.Sleep ◄───────────────┘
//
// A Resource travels around a pipeline:
//   1. Sent to pending by main
//   2. Picked up by a Poller, which does the HTTP request
//   3. Poller sends status update to StateMonitor via status channel
//   4. Poller sends completed Resource to complete channel
//   5. Resource.Sleep waits, then sends back to pending — restarting the cycle
//
// The urlStatus map lives ONLY inside the StateMonitor goroutine.
// No mutex needed — only one goroutine ever reads or writes it.
//
// Run:
//
//	go run urlpoll/urlpoll.go
//
// (Requires internet access. Ctrl+C to stop.)
package main

import (
	"log"
	"net/http"
	"time"
)

const (
	numPollers     = 2                // number of concurrent Poller goroutines
	pollInterval   = 60 * time.Second // how often to re-poll each URL
	statusInterval = 10 * time.Second // how often to print the status map
	errTimeout     = 10 * time.Second // extra back-off delay per consecutive error
)

var urls = []string{
	"http://www.google.com/",
	"http://golang.org/",
	"http://blog.golang.org/",
}

// State carries a URL and its most recent HTTP status string.
// It is passed through the status channel from Pollers to StateMonitor.
// Because it is sent by value through a channel, no locking is needed —
// the channel transfer is the synchronisation point.
type State struct {
	url    string
	status string
}

// StateMonitor owns the urlStatus map exclusively.
// It runs in its own goroutine and is the ONLY code that reads or writes
// urlStatus. This is the "share memory by communicating" pattern:
// instead of exposing the map and locking it, we expose a channel and
// let callers send updates through it.
//
// Returns a send-only channel (chan<- State) so callers can only send,
// not receive — enforced by the type system at compile time.
func StateMonitor(updateInterval time.Duration) chan<- State {
	updates := make(chan State)
	urlStatus := make(map[string]string) // owned exclusively by this goroutine
	ticker := time.NewTicker(updateInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				// Timer fired — print the current status snapshot.
				logState(urlStatus)
			case s := <-updates:
				// A Poller sent a new status — update the map.
				// Safe without a mutex because only THIS goroutine touches urlStatus.
				urlStatus[s.url] = s.status
			}
		}
	}()

	return updates // caller gets the write end of the channel
}

// logState prints the full URL→status map.
func logState(s map[string]string) {
	log.Println("Current state:")
	for k, v := range s {
		log.Printf("  %s %s", k, v)
	}
}

// Resource represents a URL being polled, along with its consecutive error count.
// The error count drives exponential back-off: each error adds errTimeout to
// the sleep before the next poll.
type Resource struct {
	url      string
	errCount int
}

// Poll performs one HTTP HEAD request and returns the status string.
// HEAD is used instead of GET because it fetches only headers, not the body —
// sufficient for a liveness check and much cheaper.
func (r *Resource) Poll() string {
	resp, err := http.Head(r.url) //nolint:noctx // acceptable for this demo
	if err != nil {
		log.Println("Error", r.url, err)
		r.errCount++
		return err.Error()
	}
	r.errCount = 0
	return resp.Status
}

// Sleep waits for the appropriate interval and then sends r back to done.
// The wait is pollInterval + (errCount × errTimeout), providing back-off
// when a URL is repeatedly failing.
//
// done chan<- *Resource — send-only channel parameter. The caller (main)
// created both ends; Sleep only needs to send.
func (r *Resource) Sleep(done chan<- *Resource) {
	time.Sleep(pollInterval + errTimeout*time.Duration(r.errCount))
	done <- r // signal: ready to be polled again
}

// Poller runs in a goroutine. It reads Resources from in, polls each one,
// sends a status update, then forwards the Resource to out.
//
// in  <-chan *Resource  — receive-only (Poller only reads from in)
// out  chan<- *Resource  — send-only   (Poller only writes to out)
// status chan<- State    — send-only   (Poller only writes status)
//
// The directional channel types make data flow explicit and are enforced
// at compile time — a Poller cannot accidentally send to its input channel.
func Poller(in <-chan *Resource, out chan<- *Resource, status chan<- State) {
	for r := range in { // blocks until a Resource arrives; exits when in is closed
		s := r.Poll()
		status <- State{r.url, s} // send status update to StateMonitor
		out <- r                  // send Resource to complete channel
	}
}

func main() {
	// pending: Resources waiting to be polled
	// complete: Resources that have just been polled
	pending, complete := make(chan *Resource), make(chan *Resource)

	// StateMonitor runs in its own goroutine and returns the channel
	// through which Pollers send status updates.
	status := StateMonitor(statusInterval)

	// Launch numPollers Poller goroutines. They all read from pending,
	// write to complete, and write status updates to status.
	// Multiple goroutines reading from the same channel is safe — the
	// channel ensures each Resource is delivered to exactly one Poller.
	for i := 0; i < numPollers; i++ {
		go Poller(pending, complete, status)
	}

	// Seed the pending channel with the initial Resources.
	// Done in a goroutine because the sends would block until a Poller
	// is ready to receive — without a goroutine this would deadlock.
	go func() {
		for _, url := range urls {
			pending <- &Resource{url: url}
		}
	}()

	// Main loop: as each Resource completes, schedule it to sleep then
	// re-enter the pending queue. This keeps the polling cycle running forever.
	for r := range complete {
		go r.Sleep(pending) // goroutine per Resource so sleeps are independent
	}
}
