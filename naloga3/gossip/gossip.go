package gossip

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
)

// Message structure for gossiping
type Message struct {
	ID   int
	Data string
}

// Start initializes the gossip protocol
func Start(id, N, M, K, basePort int) {
	// Initialize logger
	logger := govec.InitGoVector(fmt.Sprintf("Process-%d", id), fmt.Sprintf("Log-%d", id), govec.GetDefaultConfig())
	opts := govec.GetDefaultLogOptions()

	// Set up UDP listening
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", basePort+id))
	checkError(err)
	conn, err := net.ListenUDP("udp", localAddr)
	checkError(err)
	defer conn.Close()


	// Special behavior for the master process
	if id == 0 {
		time.Sleep(1 * time.Second) // Ensure all processes are ready before dissemination
		go listenMaster(conn, id, logger, opts, M)
		for i := 1; i <= M; i++ {
			msg := Message{ID: i, Data: fmt.Sprintf("Message-%d", i)}
			sendMessage(id, N, K, msg, basePort, logger, opts)
			time.Sleep(100 * time.Millisecond)
		}
	} else {
		go listen(conn, id, N, K, logger, opts)
		select {
		}

	}
}

func listenMaster(conn *net.UDPConn,  id int, logger *govec.GoLog, opts govec.GoLogOptions, M int) {
	buffer := make([]byte, 1024)
	received := map[int]bool{} // Master already "knows" the messages it sends

	// Initialize the master's received map with all messages it will send
	for i := 1; i <= M; i++ {
		received[i] = true
	}

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, _, err := conn.ReadFromUDP(buffer)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			// Terminate if no message is received before the timeout
			fmt.Printf("Process %d: No messages received. Terminating.\n", id)
			return
		} else if err != nil {
			fmt.Printf("Process %d: Error receiving message: %v\n", id, err)
			continue
		}

		// Unpack message and vector clock
		var msg Message
		logger.UnpackReceive("Received Message", buffer, &msg, opts)
	}
}

// listen handles incoming messages
func listen(conn *net.UDPConn, id, N, K int, logger *govec.GoLog, opts govec.GoLogOptions) {
	buffer := make([]byte, 1024)
	received := map[int]bool{} // Track received messages

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, _, err := conn.ReadFromUDP(buffer)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			// Terminate if no message is received before the timeout
			fmt.Printf("Process %d: No messages received. Terminating.\n", id)
			return
		} else if err != nil {
			fmt.Printf("Process %d: Error receiving message: %v\n", id, err)
			continue
		}

		// Unpack message and vector clock
		var msg Message
		logger.UnpackReceive("Received Message", buffer, &msg, opts)

		// Print only new messages
		if !received[msg.ID] {
			fmt.Printf("Process %d received: %s\n", id, msg.Data)
			received[msg.ID] = true

			// Forward message to K random processes
			sendMessage(id, N, K, msg, conn.LocalAddr().(*net.UDPAddr).Port, logger, opts)
		}
	}
}

// sendMessage forwards a message to K random processes
func sendMessage(id, N, K int, msg Message, basePort int, logger *govec.GoLog, opts govec.GoLogOptions) {
	// Select K random processes to send the message
	targets := selectRandomTargets(id, N, K)

	for _, target := range targets {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", basePort+target))
		checkError(err)
		conn, err := net.DialUDP("udp", nil, addr)
		checkError(err)

		// Prepare and send message
		logger.LogLocalEvent("Preparing Message", opts)
		msgVC := logger.PrepareSend("Sending Message", msg, opts)
		_, err = conn.Write(msgVC)
		checkError(err)
		conn.Close()
	}
}

// selectRandomTargets picks K unique random targets
func selectRandomTargets(id, N, K int) []int {
	rand.Seed(time.Now().UnixNano())
	targets := map[int]bool{}
	for len(targets) < K {
		r := rand.Intn(N)
		if r != id {
			targets[r] = true
		}
	}
	keys := []int{}
	for k := range targets {
		keys = append(keys, k)
	}
	return keys
}

// checkError handles errors
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
