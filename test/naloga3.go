package main

import (
	"flag"
	"telefon/gossip"
)

func main() {
	// Parse command-line arguments
	id := flag.Int("id", 0, "Unique process ID")
	N := flag.Int("n", 2, "Total number of processes")
	M := flag.Int("m", 1, "Total number of messages to disseminate")
	K := flag.Int("k", 1, "Number of processes to forward message to")
	port := flag.Int("port", 8000, "Base port number for communication")
	flag.Parse()

	// Start gossip protocol
	gossip.Start(*id, *N, *M, *K, *port)
}
