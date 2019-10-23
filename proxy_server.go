package main

import (
    "net/http"
    "io"
    "fmt"
)

func Redir(w http.ResponseWriter, r *http.Request) {
    // format new request
    request_path := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.URL.Host, r.URL.Path)

    // create new HTTP request with the target URL (everything else is the same)
    new_request, err := http.NewRequest(r.Method, request_path, r.Body)

    // send request to server
    fmt.Printf("Sending %s request to %s\n", r.Method, request_path)

    client := &http.Client{}
    res, err := client.Do(new_request)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer res.Body.Close()

    // copy the headers over to thre ResponseWriter. 
    //res.Header is a map of string -> slice (string)
    for key, slice := range res.Header {
        for _, val := range slice {
            w.Header().Add(key, val)
        }
    }

    // forward response to client
    _, err = io.Copy(w, res.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
}

func main() {
    http.HandleFunc("/", Redir)
    http.ListenAndServe(":8080", nil)
}
