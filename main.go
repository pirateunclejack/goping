package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	timeout int64
	size int
	count int
	typ uint8 = 8
	code uint8 = 0
)

type ICMP struct {
	Type uint8
	Code uint8
	Checksum uint16
	ID uint16
	SequenceNum uint16
}

func main() {
	GetCommandArgs()
	desIP := os.Args[len(os.Args)-1]

	conn, err := net.DialTimeout("ip:icmp", desIP, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	remoteAddr := conn.RemoteAddr()
	fmt.Printf("Ping %s [%s] with data length: %d\n", desIP, remoteAddr, size)
	for i := 1; i <= count; i++ {
		icmp := &ICMP{
			Type: typ,
			Code: code,
			Checksum: 0,
			ID: uint16(i),
			SequenceNum: uint16(i),
		}
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.BigEndian, icmp)
		data := make([]byte, size)
		binary.Write(&buffer, binary.BigEndian, data)
		data = buffer.Bytes()

		data[2] = byte(checkSum(data)>>8)
		data[3] = byte(uint8(checkSum(data)))

		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))

		startTime := time.Now()
		_, err := conn.Write(data)
		if err != nil {
			log.Panicln(err)
			return
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%d bytes from %d.%d.%d.%d: icmp_seq=%d ttl=%d time=%3.2f ms\n", n-28, buf[12], buf[13], buf[14], buf[15], uint32(buf[26])<<16 + uint32(buf[27]) ,buf[8], float32(time.Since(startTime).Microseconds())/1000)
	}
}

func GetCommandArgs() {
	flag.Int64Var(&timeout, "w", 1000, "Request timeout time")
	flag.IntVar(&size, "l", 32, "Send byte count")
	flag.IntVar(&count, "n", 4, "Request time")
	flag.Parse()
}

func checkSum(data []byte) uint16 {
	length := len(data)
	index := 0

	var sum uint32
	for length > 1 {
		sum += uint32(data[index]) << 8 + uint32(data[index+1])
		length -= 2
		index += 2
	}
	if length == 1 {
		sum += uint32(data[index])
	}

	hi := sum >> 16
	for hi != 0 {
		sum = hi + uint32(uint16(sum))
		hi = sum >> 16
	}

	return uint16(^sum)

}
