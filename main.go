package main;

import (
    "net/http"
    "log"
    "os"
)

func indexHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    http.ServeFile(pWriter, pRequest, "index.html")
}

func main() {
    serverMux := http.NewServeMux()

    serverMux.HandleFunc("/", indexHandler)

    serverAddr := "0.0.0.0:6969"
    if len(os.Args) > 2 {
        serverAddr = os.Args[1]
    }

    server := http.Server {
        Addr: serverAddr,
        Handler: serverMux,
    }

    log.Printf("Listening at %s...\n", serverAddr)
    server.ListenAndServe()
}
