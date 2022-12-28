package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"
)

// Cache is a struct that stores the HTML of a webpage and the timestamp
// when it was last requested
type Cache struct {
	HTML      string
	Timestamp time.Time
}

var cache = make(map[string]Cache)

// Worker is a struct that stores a URL and a retry limit
type Worker struct {
	URL       string
	RetryLimit int
}

// WorkQueue is a channel that receives Worker structs
var WorkQueue = make(chan Worker)

// WorkerPool is a slice of Worker channels
var WorkerPool []chan Worker

// MaxWorkers is the maximum number of active workers
const MaxWorkers = 10

func main() {
	// Create the worker pool
	for i := 0; i < MaxWorkers; i++ {
		WorkerPool = append(WorkerPool, make(chan Worker))
	}

	// create cache folder if it doesnot exists to store downloaded files
	err := os.Mkdir("cache", 0755)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("cache folder already exists")
		} else {
			fmt.Printf("error creating cache folder: %v\n", err)
		}
	}
	fmt.Println("created cache folder")

	// Start the worker pipeline
	go workerPipeline()

	http.HandleFunc("/download", downloadHandler)
	http.ListenAndServe(":8080", nil)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the query string to get the URL and retry limit
	query := r.URL.Query()
	url := query.Get("url")
	retryLimit, _ := strconv.Atoi(query.Get("retry_limit"))
	if retryLimit > 10 {
		retryLimit = 10
	}

	// Check if the webpage has been requested in the last 24 hours
	if c, ok := cache[url]; ok && time.Since(c.Timestamp) < 24*time.Hour {
		// Serve the webpage from the cache
		c.Timestamp = time.Now()
		w.Write([]byte(c.HTML))
		return
	}

	// Send the Worker to the work queue
	WorkQueue <- Worker{
		URL:       url,
		RetryLimit: retryLimit,
	}
}

func workerPipeline() {
	for {
		// Get the next worker from the work queue
		worker := <-WorkQueue

		// Find an available worker channel
		var workerChannel chan Worker
		for _, wc := range WorkerPool {
			select {
			case wc <- worker:
				workerChannel = wc
			default:
			}
		}

		// If no worker channel is available, start a new worker
		if workerChannel == nil {
			workerChannel = make(chan Worker)
			WorkerPool = append(WorkerPool, workerChannel)
			go workerFunc(workerChannel)
			workerChannel <- worker
		}
	}
}

func workerFunc(workerChannel chan Worker) {
	for {
		// Get the next worker from the worker channel
		worker := <-workerChannel

		// Download the webpage
		html, err := downloadWebpage(worker.URL, worker.RetryLimit)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Cache the webpage
		cache[worker.URL] = Cache{
			HTML:      html,
			Timestamp: time.Now(),
		}

		// Download the webpage to the local file system
		err = saveWebpage(html, worker.URL)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func downloadWebpage(url string, retryLimit int) (string, error) {
	retries := 0
	for {
		resp, err := http.Get(url)
		if err != nil {
			if retries < retryLimit {
				retries++
				continue
			}
			return "", fmt.Errorf("error downloading webpage: %v", err)
		}
		defer resp.Body.Close()

		html, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if retries < retryLimit {
				retries++
				continue
			}
			return "", fmt.Errorf("error reading webpage body: %v", err)
		}

		return string(html), nil
	}
}

func saveWebpage(html, url string) error {	
	replacer := strings.NewReplacer("/", "_", ":", "^")
	updatedUrl := replacer.Replace(url)
	file, err := os.Create("cache/" + fmt.Sprintf("%s.html", updatedUrl))
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, html)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

