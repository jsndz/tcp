package main

import (
	"fmt"
	"syscall"
)

func main() {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, 6)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)

	buf := make([]byte, 65535)

	for {

		n, from, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v", from)
		fmt.Printf("%x", buf[:n])
	}
}
