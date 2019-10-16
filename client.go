package main

import (
    "net/http"
    "io/ioutil"
    "fmt"
)

func main() {
    s := "www.gmail.com"
    res, err := http.Get(fmt.Sprintf("http://localhost:8080/proxy/%s", s))
    if err != nil {
        fmt.Println(err)
        return;
    }
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
        return;
    }
    fmt.Printf(string(body))
}
