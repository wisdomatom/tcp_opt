package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func req(ch chan string, num int) {

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", strings.NewReader(""))
	if err != nil {
		e := fmt.Sprintf("new request error (%v)", err)
		log.Printf(e)
		ch <- e
		return
	}
	req.Header.Set("num", fmt.Sprintf("%v", num))
	// resp, err := http.Get("http://127.0.0.1:8000/")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		e := fmt.Sprintf("http request error (%v)", err)
		log.Printf(e)
		ch <- e
		return
	}
	defer resp.Body.Close()
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e := fmt.Sprintf("response body read error (%v)", err)
		log.Printf(e)
		ch <- e
		return
	}
	// _ = bts
	// log.Printf("http ok: (%v)", string(bts))
	ch <- fmt.Sprintf("http success: (%v)", string(bts))
}

func main() {
	conc := 40000
	start := time.Now().Unix()
	ch := make(chan string, conc)
	for i := 0; i < conc; i++ {
		go req(ch, i)
	}
	sendOver := time.Now().Unix()
	success := 0
	fail := 0
	for true {
		select {
		case msg := <-ch:
			if strings.Contains(msg, "error") {
				fmt.Printf("error: (%v)", msg)
				fail++
			}
			if strings.Contains(msg, "success") {

				success++
			}
			if success+fail == conc {
				log.Printf("\nsuccess: (%v)", success)
				log.Printf("failed: (%v)", fail)
				log.Printf("send use: (%v)", (sendOver - start))
				log.Printf("process request use: (%v)", (time.Now().Unix() - sendOver))
				return
			}
		}

	}
}
