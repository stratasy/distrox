package main;

import (
    "net/http"
    "bufio"
    "io/ioutil"
    "fmt"
    "os"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
		res, err := http.Get("http://localhost:8080/" + line)
		fmt.Println(string("http://localhost:8080/" + line))
        if err != nil {
            fmt.Println(err)
            continue
        }
	    body, err := ioutil.ReadAll(res.Body)
	    if err != nil {
	        fmt.Println(err)
	        return;
	    }
	    fmt.Println(string(body))
    }
}