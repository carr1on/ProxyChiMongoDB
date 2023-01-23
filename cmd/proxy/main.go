package main

import (
	"io"
	"log"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("unable to bind port")
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print("unable to accept connection")
		}
		go echo(conn)
	}
}

func echo(src net.Conn) {
	var (
		adr       []string
		stringadr string
	)
	adr = append(adr, "app1:8081", "app2:8081")

	a := time.Now().UnixNano() / int64(time.Millisecond)
	b := a % 2
	if b == 0 {
		stringadr = adr[1]
	} else {
		stringadr = adr[0]
	}

	dst, err := net.Dial("tcp", stringadr)
	if err != nil {
		log.Print("unable connect to api")
	}
	defer dst.Close()
	go func() {
		if _, err = io.Copy(dst, src); err != nil {
			log.Print("err")
		}
	}()
	if _, err = io.Copy(src, dst); err != nil {
		log.Print(err)
	}
}
