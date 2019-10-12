package main

import (
    "net/http"
    "io/ioutil"
    "fmt"
)

func main() {
    res, err := http.Get("http://localhost:8080/test")
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
