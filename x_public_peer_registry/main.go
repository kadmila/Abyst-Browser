package main

import (
	"abyss_open_reg/aurl"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type HostData struct {
	connection_info   []byte
	join_req_consumer SteppingConsumer
	last_update       time.Time
}

var (
	live_host_data = make(map[string]*HostData)
	mu             sync.RWMutex
)

func main() {
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".aml", "text/aml")
	mime.AddExtensionType(".obj", "model/obj")

	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/wait", eventWaiter)
	http.HandleFunc("/api/random", randomHandler)
	http.HandleFunc("/api/request", joinRequestHandler)

	static_fs := http.FileServer(http.Dir("./static/"))
	main_fs := http.FileServer(http.Dir("./main/"))

	http.Handle("/static/", http.StripPrefix("/static/", static_fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./main/index.html")
			return
		}
		main_fs.ServeHTTP(w, r)
	})

	go func() {
		for {
			time.Sleep(20 * time.Second)
			cleanup()
		}
	}()

	if (len(os.Args) > 1) && os.Args[1] == "--local" {
		log.Println("Starting server on 127.0.0.1:80")
		log.Fatal(http.ListenAndServe("127.0.0.1:80", nil))
	} else {
		log.Println("Starting server on https://irublue.com")
		log.Fatal(http.ListenAndServeTLS(":443", "../cert_man/irublue.com/fullchain.pem", "../cert_man/irublue.com/privkey.pem", nil))
	}
}

// removes host data that had no activity for 1 min.
func cleanup() {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	for k, v := range live_host_data {
		if now.Sub(v.last_update) > 1*time.Minute {
			fmt.Println("outdated: " + k)
			delete(live_host_data, k)
		}
	}
}

// registerHandler handles POST requests to /register
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the entire request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request: Failed to read body", http.StatusBadRequest)
		return
	}

	// Split into three parts
	parts := bytes.SplitN(bodyBytes, []byte{','}, 3)
	if len(parts) != 3 {
		http.Error(w, "Bad Request: Expected 3 newline-separated values, received "+strconv.Itoa(len(parts)), http.StatusBadRequest)
		return
	}

	//parse the first string (AURL)
	abyss_url, err := aurl.TryParse(string(parts[0]))
	if err != nil {
		http.Error(w, "Bad Request: Failed to parse AURL: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Store in the map (with lock)
	mu.Lock()
	//delete(live_host_data, abyss_url.Hash)
	live_host_data[abyss_url.Hash] = &HostData{
		connection_info:   bodyBytes,
		join_req_consumer: MakeSteppingConsumer(),
		last_update:       time.Now(),
	}
	mu.Unlock()
	fmt.Println("registered: " + abyss_url.Hash)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("server: registration success"))
}

// pending eventWaiter
func eventWaiter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Bad Request: URL parameter 'id' missing", http.StatusBadRequest)
		return
	}

	mu.Lock()
	host_data, ok := live_host_data[id]
	if ok {
		host_data.last_update = time.Now()
	}
	mu.Unlock()

	if !ok {
		http.Error(w, "Not registered", http.StatusConflict)
		return
	}

	ctx, ctx_cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer ctx_cancel()

	fmt.Println("waiting: " + id)
	argument, confirm_ch, ok, ok_mono := host_data.join_req_consumer.TryConsume(ctx)

	if !ok_mono {
		http.Error(w, "Conflict: Already waiting", http.StatusConflict)
		fmt.Println("waiting-conflict: " + id)
		return
	}

	if !ok {
		http.Error(w, "", http.StatusRequestTimeout) //retry required.
		fmt.Println("waiting-timeout: " + id)
		return
	}

	confirm_ch <- true
	w.Write(argument.([]byte))
	fmt.Println("waiting-received: " + id)
}

// randomHandler handles GET requests to /random
func randomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	excl := r.URL.Query().Get("excl")
	if excl == "" {
		http.Error(w, "Bad Request: URL parameter 'excl' missing", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Pick a random entry
	var keys []string
	for k := range live_host_data {
		if k == excl {
			continue
		}
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		http.Error(w, "No peers available", http.StatusNotFound)
		return
	}
	randomKey := keys[rand.Intn(len(keys))]

	w.Write([]byte(randomKey))
}

func joinRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Bad Request: URL parameter 'id' missing", http.StatusBadRequest)
		return
	}

	targ := r.URL.Query().Get("targ")
	if targ == "" {
		http.Error(w, "Bad Request: URL parameter 'targ' missing", http.StatusBadRequest)
		return
	}

	mu.Lock()
	host_data, id_ok := live_host_data[id]
	targ_data, targ_ok := live_host_data[targ]
	mu.Unlock()

	if !id_ok {
		http.Error(w, "host not registered", http.StatusNotFound)
		return
	}
	if !targ_ok {
		http.Error(w, "target not registered", http.StatusNotFound)
		return
	}

	fmt.Println("requesting: " + id)
	consume_ch, ok := targ_data.join_req_consumer.TryPut(host_data.connection_info)
	if !ok {
		http.Error(w, "", http.StatusTooManyRequests)
		return
	}

	select {
	case <-consume_ch:
		w.Write(targ_data.connection_info)
	case <-time.After(time.Second * 5):
		http.Error(w, "FATAL: server corrupted or overloaded", http.StatusInternalServerError)
	}
}
