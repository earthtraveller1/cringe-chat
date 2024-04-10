package main;

import (
    "net/http"
    "log"
    "os"
    "fmt"
    "strings"
    "context"
    "text/template"
    "encoding/json"

    "github.com/gorilla/websocket"
)

type ChatTemplateParameters struct {
    Username string
}

type ChatMessage struct {
    Username string `json:"username"`
    Message string `json:"message"`
}

func indexHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    http.ServeFile(pWriter, pRequest, "pages/index.html")
}

func chatHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    chatTemplate, err := template.ParseFiles("pages/chat.html") 
    if err != nil {
        fmt.Fprintf(pWriter, "Failed to parse the template.")
        pWriter.WriteHeader(500)
    }

    parameters := ChatTemplateParameters {
        Username: fmt.Sprintf("\"%s\"", pRequest.FormValue("username")),
    }

    err = chatTemplate.Execute(pWriter, parameters)
    if err != nil {
        fmt.Fprintf(pWriter, "Failed to execute the template.")
        pWriter.WriteHeader(500)
    }
}

func chatSocketHandler(pMessages chan Message, pWriter http.ResponseWriter, pRequest *http.Request) {
    upgrader := websocket.Upgrader {}
    connection, err := upgrader.Upgrade(pWriter, pRequest, nil)
    if err != nil {
        fmt.Fprintf(pWriter, "Failed to upgrade the websocket request. Error: %e\n", err)
        pWriter.WriteHeader(500)
        return
    }

    defer connection.Close()

    messageType, usernameMessage, err := connection.ReadMessage()
    if err != nil {
        log.Printf("Error while trying to read a message from the websocket. Error: %e\n", err)
        return
    }

    go func() {
        for {
            message := <- pMessages

            err := connection.WriteMessage(websocket.TextMessage, json.Marshal(message))
            if err != nil {
                log.Printf("Error while trying to send a message back to the client. Error: %e\n", err)
                return
            }
        }
    } ()

    for {
        messageType, message, err := connection.ReadMessage()
        if err != nil {
            log.Printf("Error while trying to read a message from the websocket. Error: %en\n", err)
            return
        }

        pMessages <- Message {
            Username: string(usernameMessage),
            Message: string(message)
        }
    }
}

func staticFilesHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    realPath, _ := strings.CutPrefix(pRequest.URL.Path, "/")
    http.ServeFile(pWriter, pRequest, realPath)
}

func main() {
    serverMux := http.NewServeMux()

    serverMux.HandleFunc("/", indexHandler)
    serverMux.HandleFunc("/chat", chatHandler)
    serverMux.HandleFunc("/build/", staticFilesHandler)
    serverMux.HandleFunc("/vendor/", staticFilesHandler)

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
