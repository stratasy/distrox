package main

import (
    "net/http"
    "io/ioutil"
    "fmt"
)

func Redir(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET"{
        res, err := http.Get("https://www.bing.com/")
        if err != nil {
            fmt.Println(err)
        }
	    body, err := ioutil.ReadAll(res.Body)
	    if err != nil {
	        fmt.Println(err)
	        return;
	    }
        
        fmt.Fprintf(w, string(body))
        fmt.Fprintf(w, "\nWould you like to make this your default browser?\n")
    }
}

func Default(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET"{
        fmt.Fprintf(w, "Received default request!\n")
    }
    
}

func main() {
    http.HandleFunc("/www.google.com", Redir)
    http.HandleFunc("/", Default)
    //http.HandleFunc("/", TestBase)
    http.ListenAndServe(":8080", nil)
}