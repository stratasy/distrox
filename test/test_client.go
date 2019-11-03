package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	s := "www.gmail.com"
	res, err := http.Post(fmt.Sprintf("http://localhost:8080/proxy/%s", s), "application/json", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(string(body))
}
