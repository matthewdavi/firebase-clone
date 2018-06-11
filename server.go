package main

import (
    "fmt"
    "net/http"
    "log"
    "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world")
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(os.Getenv("PORT"), nil))
}
