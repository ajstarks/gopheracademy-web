package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var config struct { // [1]
	port int
	host string
}

const (
	usage = `usage: %s
Run HTTP server

Options:
`
)

func main() {
	flag.IntVar(&config.port, "port", config.port, "port to listen on")    // [2]
	flag.StringVar(&config.host, "host", config.host, "host to listen on") // [3]
	flag.Usage = func() {                                                  // [4]
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse() // [5]

	http.HandleFunc("/", handler)
	addr := fmt.Sprintf("%s:%d", config.host, config.port)
	fmt.Printf("server ready on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("error: %s", err)
	}

}

func init() { // [6]
	// Set defaults
	s := os.Getenv("HTTPD_PORT")
	p, err := strconv.Atoi(s)
	if err == nil {
		config.port = p
	} else {
		config.port = 8080
	}

	h := os.Getenv("HTTPD_HOST")
	if len(h) > 0 {
		config.host = h
	} else {
		config.host = "localhost"
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Gophers\n")
}
