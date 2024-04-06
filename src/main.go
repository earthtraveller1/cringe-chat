package main;

import (
    "net/http"
    "log"
    "os"
    "fmt"
    "strings"
    "context"
)

func indexHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    http.ServeFile(pWriter, pRequest, "pages/index.html")
}

func buildFilesHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    realPath, _ := strings.CutPrefix(pRequest.URL.Path, "/")
    log.Printf("Serving %s...\n", realPath)
    http.ServeFile(pWriter, pRequest, realPath)
}

func main() {
    serverMux := http.NewServeMux()

    serverMux.HandleFunc("/", indexHandler)
    serverMux.HandleFunc("/build/", buildFilesHandler)

    serverAddr := "0.0.0.0:6969"
    if len(os.Args) > 2 {
        serverAddr = os.Args[1]
    }

    server := http.Server {
        Addr: serverAddr,
        Handler: serverMux,
    }

    log.Printf("Listening at %s...\n", serverAddr)
    go server.ListenAndServe()

    for {
        var command string
        fmt.Scanln(&command)

        if strings.HasPrefix(command, "q") {
            server.Shutdown(context.TODO())
            break
        }
    }
}
