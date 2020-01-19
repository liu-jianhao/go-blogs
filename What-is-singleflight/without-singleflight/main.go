package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/baidu", func(w http.ResponseWriter, r *http.Request) {
		status, err := baiduStatus()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Baidu Status: %v", status)
	})

	http.ListenAndServe("127.0.0.1:8080", nil)
}

func baiduStatus() (int, error) {
	log.Println("Making request to Baidu API")
	defer log.Println("Request to Baidu API Complete")

	time.Sleep(1 * time.Second)

    resp, err := http.Get("https://www.baidu.com")
	if err != nil {
        log.Println("get baidu.com error")
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, err
}
