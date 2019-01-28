package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"time"
)

func main() {
	port := flag.Int("p", 10001, "proxy port")
	serverPort := flag.Int("s", 1234, "server port")
	d := flag.Int("delay", 10, "delay in ms (one way)")
	flag.Parse()

	delay := time.Duration(*d) * time.Millisecond
	if err := run(*port, *serverPort, delay); err != nil {
		panic(err)
	}
}

func run(port, serverPort int, delay time.Duration) error {
	receiverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", serverPort))
	if err != nil {
		return err
	}

	saddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	sconn, err := net.ListenUDP("udp", saddr)
	if err != nil {
		return err
	}
	fmt.Printf("Proxying connections %s <-> %s (%s delay)\n", sconn.LocalAddr(), receiverAddr, delay)

	caddr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return err
	}
	cconn, err := net.ListenUDP("udp", caddr)
	if err != nil {
		return err
	}

	clientAddrChan := make(chan *net.UDPAddr, 1)
	go func() {
		if err := runUpstream(sconn, cconn, receiverAddr, clientAddrChan, delay); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := runDownstream(sconn, cconn, clientAddrChan, delay); err != nil {
			panic(err)
		}
	}()

	select {}
}

func runUpstream(sconn, cconn *net.UDPConn, receiverAddr *net.UDPAddr, clientAddrChan chan<- *net.UDPAddr, delay time.Duration) error {
	var hasClientAddr bool
	b := make([]byte, 8)
	for {
		n, addr, err := sconn.ReadFromUDP(b)
		if err != nil {
			return err
		}
		if n != 8 {
			return errors.New("small read")
		}
		if !hasClientAddr {
			clientAddrChan <- addr
			hasClientAddr = true
		}
		time.AfterFunc(delay, func() {
			if _, err := cconn.WriteTo(b, receiverAddr); err != nil {
				panic(err)
			}
		})
	}
}

func runDownstream(sconn, cconn *net.UDPConn, clientAddrChan <-chan *net.UDPAddr, delay time.Duration) error {
	senderAddr := <-clientAddrChan
	b := make([]byte, 8)
	for {
		n, _, err := cconn.ReadFromUDP(b)
		if err != nil {
			return err
		}
		if n != 8 {
			return errors.New("small read")
		}
		time.AfterFunc(delay, func() {
			if _, err := sconn.WriteTo(b, senderAddr); err != nil {
				panic(err)
			}
		})
	}
}
