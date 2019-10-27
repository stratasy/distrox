package main;

import (
    "net/http"
    "bufio"
    "io"
    "io/ioutil"
    "fmt"
    "os"
    "strings"
)

// Takes a body reader, and a map of [keywords] to replace with value string
// Returns the modified body, as a byte array.
func ParseReplaceBody(bodyReader io.Reader, keyWords map[string]string) []byte {
  body, err := ioutil.ReadAll(bodyReader)
  if err != nil {
      fmt.Println(err)
      return []byte{}
  }
  bodyStr := string(body)
  for key, value := range keyWords {
    bodyStr = strings.ReplaceAll(bodyStr, key, value)
  }
  return []byte(bodyStr)
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()

        // localhost:8080/proxy/[TARGET URL]
        proxy_path := fmt.Sprintf("http://localhost:8080/proxy/%s", line)

        res, err := http.Get(proxy_path)
        if err != nil {
            fmt.Println(err)
            continue
        }
        // Testing parse and replace.
        m := make(map[string]string)
        m["Google"] = "TESTING VALUE"
        m["google"] = "SECONDARY TESTING VALUE"
        fmt.Println(string(ParseReplaceBody(res.Body, m)))
    }
}
