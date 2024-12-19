package gossip

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
)

type Message struct {
	ID   int
	Data string
}

func Start(id, N, M, K, basePort int) {
	logger := govec.InitGoVector(fmt.Sprintf("Process-%d", id), fmt.Sprintf("Log-%d", id), govec.GetDefaultConfig())
	opts := govec.GetDefaultLogOptions()

	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", basePort+id))
	checkError(err)
	conn, err := net.ListenUDP("udp", localAddr)
	checkError(err)
	defer conn.Close()


	if id == 0 {
		time.Sleep(1 * time.Second)
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

func listenMaster(conn *net.UDPConn,  id int, logger *govc.GoLog, opts govec.GoLogOptions, M int) {
	buffer := make([]byte, 1024)
	received := map[int]bool{}

	for i := 1; i <= M; i++ {
		received[i] = true
	}

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, _, err := conn.ReadFromUDP(buffer)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Printf("Process %d: No messages received. Terminating.\n", id)
			return
		} else if err != nil {
			fmt.Printf("Process %d: Error receiving message: %v\n", id, err)
			continue
		}

		var msg Message
		logger.UnpackReceive("Received Message", buffer, &msg, opts)
	}
}

func listen(conn *net.UDPConn, id, N, K int, logger *govec.GoLog, opts govec.GoLogOptions) {
	buffer := make([]byte, 1024)
	received := map[int]bool{}

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, _, err := conn.ReadFromUDP(buffer)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Printf("Process %d: No messages received. Terminating.\n", id)
			return
		} else if err != nil {
			fmt.Printf("Process %d: Error receiving message: %v\n", id, err)
			continue
		}

		var msg Message
		logger.UnpackReceive("Received Message", buffer, &msg, opts)

		if !received[msg.ID] {
			fmt.Printf("Process %d received: %s\n", id, msg.Data)
			received[msg.ID] = true

			sendMessage(id, N, K, msg, conn.LocalAddr().(*net.UDPAddr).Port, logger, opts)
		}
	}
}

func sendMessage(id, N, K int, msg Message, basePort int, logger *govec.GoLog, opts govec.GoLogOptions) {
	targets := selectRandomTargets(id, N, K)

	for _, target := range targets {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", basePort+target))
		checkError(err)
		conn, err := net.DialUDP("udp", nil, addr)
		checkError(err)

		logger.LogLocalEvent("Preparing Message", opts)
		msgVC := logger.PrepareSend("Sending Message", msg, opts)
		_, err = conn.Write(msgVC)
		checkError(err)
		conn.Close()
	}
}

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

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
