package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
	"math/rand"
	"github.com/DistributedClocks/GoVector/govec"
)

type Message struct {
	ID   int
	Data string
}

var start chan bool
var stopHeartbeat bool
var N int
var id int
var M int
var K int
var rootPort int
var received []bool


var Logger *govec.GoLog
var opts govec.GoLogOptions

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}


func receive(addr *net.UDPAddr) {
	// Poslušamo
	conn, err := net.ListenUDP("udp", addr)
	checkError(err)
	conn.SetDeadline(time.Now().Add(time.Second * time.Duration(M) ))
	defer conn.Close()

	fmt.Println("Proces", id, "posluša na", addr)


	for {
		buffer := make([]byte, 1024)
		var msg []byte
		var msgId int

		_, err = conn.Read(buffer)
		checkError(err)
		Logger.UnpackReceive("Prejeto sporocilo ", buffer, &msg, opts)
		msgId, _ = strconv.Atoi(string(msg))

		if !received[msgId] {
			fmt.Printf("Process %d je prejel: %s\n", id, string(msg))
			received[msgId] = true
			sendMessages(string(msg))
		}
	}

}


func sendMessages(msg string) {
	targets := selectRandomTargets()

	for _, target := range targets {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", rootPort+target))
		checkError(err)
		conn, err := net.DialUDP("udp", nil, addr)
		checkError(err)

		Logger.LogLocalEvent("Pripravljam sporocilo", opts)
		msgVC := Logger.PrepareSend("Posiljam sporocilo", msg, opts)
		_, err = conn.Write(msgVC)
		checkError(err)
		conn.Close()
	}
}

func selectRandomTargets() []int {
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


func main() {
	// Preberi argumente
	portPtr := flag.Int("p", 9000, "# start port")
	idPtr := flag.Int("id", 0, "# process id")
	NPtr := flag.Int("n", 2, "total number of processes")
	KPtr := flag.Int("k", 2, "number of processes to forward message to")
	MPtr := flag.Int("m", 5, "number of messages to send")
	flag.Parse()

	id = *idPtr	
	rootPort = *portPtr
	K = *KPtr
	M = *MPtr
	N = *NPtr

	Logger = govec.InitGoVector("Proces-"+strconv.Itoa(id), "Log-Proces-"+strconv.Itoa(id), govec.GetDefaultConfig())
	opts = govec.GetDefaultLogOptions()


	received = make([]bool, M)
	localAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", rootPort+id))

	if id == 0 {
		for i := range received {
			received[i] = true
		}
		go receive(localAddr)
		time.Sleep(time.Second * 2)
		for i := range M {
			time.Sleep(time.Second)
			sendMessages(strconv.Itoa(i))
		}
	} else {

		fmt.Println("Pproces", id, "posilja na:", localAddr)
		go receive(localAddr)
	}

	for stop := false; !stop; time.Sleep(time.Millisecond * 100) {
		stop = true
		for _, r := range received {
			if !r {
				stop = false
				break
			}
		}
	}


	fmt.Println("Koncan proces: ", id)


}