package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/jsndz/tcp/internal/connection"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Source Port: ")

	srcStr, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	srcStr = strings.TrimSpace(srcStr)

	srcPort, err := strconv.Atoi(srcStr)
	if err != nil {
		panic(err)
	}

	fmt.Print("Destination Port: ")

	dstStr, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	dstStr = strings.TrimSpace(dstStr)

	dstPort, err := strconv.Atoi(dstStr)
	if err != nil {
		panic(err)
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, 200)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)

	conn := connection.NewConnection(fd, "127.0.0.1", uint16(srcPort), uint16(dstPort))
	fmt.Print("Mode (server/client): ")

	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)
	if mode == "server" {

		err = conn.Accept()
		if err != nil {
			panic(err)
		}

	} else {

		err = conn.Connect()
		if err != nil {
			panic(err)
		}
	}
	go conn.Write()
	go conn.Read()
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		text = strings.TrimSpace(text)
		conn.SendChan <- []byte(text)
	}
}