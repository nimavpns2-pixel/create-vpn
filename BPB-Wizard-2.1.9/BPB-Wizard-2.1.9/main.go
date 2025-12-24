package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	srcPath    string
	workerPath string
	cachePath  string
	isAndroid  = false
	workerURL  = "https://github.com/bia-pain-bache/BPB-Worker-Panel/releases/latest/download/worker.js"
	VERSION    = "dev"
)

func init() {
	showVersion := flag.Bool("version", false, "Show version")
	testLocal := flag.Bool("test-local-worker", false, "Test local worker.js path resolution and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	// If running in test mode, just test worker fetching logic and exit
	if *testLocal {
		initPaths()
		if err := downloadWorker(); err != nil {
			fmt.Printf("Error testing worker fetching: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Local worker.js test completed successfully.")
		os.Exit(0)
	}

	initPaths()
	setDNS()
	checkAndroid()
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		runWizard()
	}()

	server := &http.Server{Addr: ":8976"}
	http.HandleFunc("/oauth/callback", callback)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			failMessage("Error serving localhost.")
			log.Fatalln(err)
		}
	}()

	wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
