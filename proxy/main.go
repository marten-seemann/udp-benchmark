package main

import (
	"errors"
	"fmt"
	"net"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	receiverAddr, err := net.ResolveUDPAddr("udp", "localhost:1234")
	if err != nil {
		return err
	}

	saddr, err := net.ResolveUDPAddr("udp", "localhost:10001")
	if err != nil {
		return err
	}
	sconn, err := net.ListenUDP("udp", saddr)
	if err != nil {
		return err
	}
	fmt.Printf("Proxying connections %s <-> %s\n", sconn.LocalAddr(), receiverAddr)

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
		if err := runUpstream(sconn, cconn, receiverAddr, clientAddrChan); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := runDownstream(sconn, cconn, clientAddrChan); err != nil {
			panic(err)
		}
	}()

	select {}
}

func runUpstream(sconn, cconn *net.UDPConn, receiverAddr *net.UDPAddr, clientAddrChan chan<- *net.UDPAddr) error {
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
		if _, err := cconn.WriteTo(b, receiverAddr); err != nil {
			return err
		}
	}
}

func runDownstream(sconn, cconn *net.UDPConn, clientAddrChan <-chan *net.UDPAddr) error {
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
		if _, err := sconn.WriteTo(b, senderAddr); err != nil {
			return err
		}
	}
}
