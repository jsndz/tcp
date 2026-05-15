package main

import (
	"fmt"
	"syscall"
)

func main() {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, 200)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)
	addr := syscall.SockaddrInet4{
		Addr: [4]byte{127, 0, 0, 1},
	}
	data := []byte("hello myself")
	err = syscall.Sendto(fd, data, 0, &addr)
	if err != nil {
		panic(err)
	}

	fmt.Print("Sent")
}
