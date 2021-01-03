package main

import (
	"errors"
	"log"
	"time"
)

func testDefer() {
	defer func() {
		log.Printf("first defer.")
	}()

	err := errors.New("some error")
	if err != nil {
		return
	}
	defer func() {
		log.Printf("second defer.")
	}()
}

func testPanic() {
	defer func() {
		e := recover()
		if e != nil {
			log.Printf("recover : (%v)", e)
		}
	}()

	// panic("error happen")

	// func() {
	// 	panic("general func call error")
	// }()

	// go func() {
	// 	panic("goroutine error")
	// }()

	time.Sleep(time.Second * 2)

}

func testFor() {
	for i := 0; i < 10; i++ {
		go func(num *int) {
			log.Printf("i: %v", *num)
		}(&i)
	}
	time.Sleep(time.Second * 5)
}

func testChan() {
	ch := make(chan int, 100)
	producer := func(ch chan int) {
		for i := 0; i < 10; i++ {
			ch <- 10
		}
		close(ch)
	}

	consumer := func(ch chan int) {
		for true {
			select {
			case msg := <-ch:
				log.Printf("receive: %v", msg)
				time.Sleep(time.Millisecond * 100)
			}
		}
		// for msg := range ch {
		// 	log.Printf("receive: %v", msg)
		// 	time.Sleep(time.Millisecond * 100)
		// }
	}

	go producer(ch)
	consumer(ch)

	// forever := make(chan struct{}, 1)
	// <-forever
}

func main() {

	// testDefer()

	// testPanic()

	// testFor()

	testChan()

}
