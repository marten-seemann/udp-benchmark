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

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("Proxying connections %s <-> %s (%s delay)\n", conn.LocalAddr(), receiverAddr, delay)

	for {
		b := make([]byte, 8)
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			return err
		}
		if n != 8 {
			return errors.New("small read")
		}

		time.AfterFunc(delay, func() {
			if _, err := conn.WriteTo(b, receiverAddr); err != nil {
				panic(err)
			}
		})
	}
}
