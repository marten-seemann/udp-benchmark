package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	filename := flag.String("o", "rtt.txt", "output filename")
	flag.Parse()
	file, err := os.Create(*filename)
	if err != nil {
		panic(err)
	}
	if err := run(file); err != nil {
		panic(err)
	}
}

func run(output io.Writer) error {
	const runTime = 3 * time.Second
	const interval = time.Millisecond
	const numPackets = uint64(runTime / interval)

	var mutex sync.Mutex

	sendTimes := make(map[uint64]time.Time)

	saddr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return err
	}
	sconn, err := net.ListenUDP("udp", saddr)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	b := make([]byte, 8)
	go func() {
		for i := uint64(0); i < numPackets; i++ {
			n, _, err := sconn.ReadFrom(b)
			if n != 8 {
				panic("short read")
			}
			if err != nil {
				panic(err)
			}
			now := time.Now()
			pn := binary.BigEndian.Uint64(b)
			mutex.Lock()
			sendTime := sendTimes[pn]
			mutex.Unlock()
			fmt.Fprintf(output, "%d %d\n", pn, uint64(now.Sub(sendTime)/time.Nanosecond))
		}

		close(done)
	}()

	caddr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return err
	}
	cconn, err := net.DialUDP("udp", caddr, sconn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		return err
	}

	p := make([]byte, 8)
	for i := uint64(0); i < numPackets; i++ {
		binary.BigEndian.PutUint64(p, i)
		if _, err := cconn.Write(p); err != nil {
			return err
		}
		now := time.Now()
		mutex.Lock()
		sendTimes[i] = now
		mutex.Unlock()

		time.Sleep(interval)
	}

	<-done
	return nil
}
