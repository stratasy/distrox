package main

import (
    "net/http"
    "io/ioutil"
    "fmt"
    "strings"
)

func Redir(w http.ResponseWriter, r *http.Request) {
    // strip: /proxy/www.google.com -> www.google.com
    request_path := strings.TrimPrefix(r.URL.Path, "/proxy/")
    // append http://
    request_path = fmt.Sprintf("http://%s", request_path)

    // create new HTTP request with the target URL (everything else is the same)
    new_request, err := http.NewRequest(r.Method, request_path, r.Body)

    // send request to server
    fmt.Printf("Sending HTTP request to %s\n", request_path)

    client := &http.Client{}
    res, err := client.Do(new_request)
    if err != nil {
        fmt.Println(err)
        return
    }

    // forward response to client
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Fprintf(w, string(body))
}

func main() {
    http.HandleFunc("/proxy/", Redir)
    http.ListenAndServe(":8080", nil)
}
