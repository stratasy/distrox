package main

import (
    "net/http"
    "io/ioutil"
    "fmt"
    "strings"
)

func Redir(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // filter: /proxy/www.google.com -> www.google.com
        request_path := strings.TrimPrefix(r.URL.Path, "/proxy/")
        // append http://
        request_path = fmt.Sprintf("http://%s", request_path)

        // send request to server
        fmt.Printf("Sending HTTP request to %s\n", request_path)
        res, err := http.Get(request_path)
        if err != nil {
            fmt.Println(err)
            return;
        }

        // forward response to client
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
            fmt.Println(err)
            return;
        }
        fmt.Fprintf(w, string(body))
    }
}

func main() {
    http.HandleFunc("/proxy/", Redir)
    http.ListenAndServe(":8080", nil)
}
