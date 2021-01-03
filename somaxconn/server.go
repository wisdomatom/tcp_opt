package main

import (
	"log"
	"net"
	"time"
)

func connStatus(l net.Listener) {
	time.Sleep(time.Second * 8)
	conn, err := l.Accept()
	if err != nil {
		log.Printf("accept error: (%v)", err)
		return
	}
	_, err = conn.Write([]byte("accept"))
	if err != nil {
		log.Printf("tcp write error (%v)", err)
		return
	}
	conn.Close()
}

func main() {
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Printf("net listen error (%v)", err)
		return
	}
	_ = l

	// conn, err := l.Accept()
	// if err != nil {
	// 	log.Printf("net accept error (%v)", err)
	// 	return
	// }
	// _ = conn
	go connStatus(l)
	forever := make(chan struct{})
	<-forever
}
