package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func dial(ch chan string) {
	// conn, err := net.Dial("tcp", ":8000")
	conn, err := net.DialTimeout("tcp", ":8000", time.Second*3)
	if err != nil {
		e := fmt.Sprintf("net dial error (%v)", err)
		log.Printf(e)
		ch <- e
		return
	}
	_ = conn

	ch <- fmt.Sprintf("success: %v", conn.RemoteAddr())
}

func main() {
	conc := 6
	ch := make(chan string, conc)
	for i := 0; i < conc; i++ {
		go dial(ch)
	}

	success := 0
	fail := 0
	for true {
		select {
		case msg := <-ch:
			if strings.Contains(msg, "success") {
				success++
			}
			if strings.Contains(msg, "error") {
				fail++
			}
			if success+fail == conc {
				log.Printf("success: %v", success)
				log.Printf("error: %v", fail)
				return
			}
		}
	}
}
