package main;

import (
    "net/http"
    "log"
    "os"
    "fmt"
    "strings"
    "context"
    "encoding/json"
    "sync"

    "github.com/gorilla/websocket"
)

type ChatTemplateParameters struct {
    Username string
}

type ChatMessage struct {
    Username string `json:"username"`
    Message string `json:"message"`
    Closing bool `json:"closing"`
}

func indexHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    http.ServeFile(pWriter, pRequest, "pages/index.html")
}

func chatHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    http.ServeFile(pWriter, pRequest, "pages/chat.html")
}

func broadcastMessage(pMessageListeners [](chan ChatMessage), pMessage ChatMessage) {
    for _, listener := range pMessageListeners {
        listener <- pMessage
    }
}

func chatSocketHandler(pMessageMutex sync.Mutex, pMessageListeners *[](chan ChatMessage), pWriter http.ResponseWriter, pRequest *http.Request) {
    upgrader := websocket.Upgrader {}
    connection, err := upgrader.Upgrade(pWriter, pRequest, nil)
    if err != nil {
        log.Printf("Failed to upgrade the websocket request. Error: %e\n", err)
        pWriter.WriteHeader(500)
        return
    }

    defer connection.Close()

    _, usernameMessage, err := connection.ReadMessage()
    if err != nil {
        log.Printf("Error while trying to read a message from the websocket. Error: %e\n", err)
        return
    }

    messageListener := make(chan ChatMessage)
    messageListenerIndex := len(*pMessageListeners)

    pMessageMutex.Lock()
    *pMessageListeners = append(*pMessageListeners, messageListener)

    pMessageMutex.Unlock()

    go func() {
        for message := range messageListener {
            if message.Closing {
                return
            }

            jsonMessage, err := json.Marshal(message)
            if err != nil {
                log.Printf("Error while trying to marshal a message into a JSON")
                return
            }

            err = connection.WriteMessage(websocket.TextMessage, jsonMessage)
            if err != nil {
                log.Printf("Error while trying to send a message back to the client. Error: %e\n", err)
                return
            }
        }
    } ()

    pMessageMutex.Lock()

    broadcastMessage(*pMessageListeners, ChatMessage {
        Username: "SYSTEM",
        Message: fmt.Sprintf("%s joined the chat.", usernameMessage),
        Closing: false,
    })

    pMessageMutex.Unlock()

    for {
        message := []byte{}
        _, message, err := connection.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, 0) {
                log.Printf("A client has closed the connection to the server.")
                // Remove the channel from the list of channels.
                pMessageMutex.Lock()
                (*pMessageListeners)[messageListenerIndex] = (*pMessageListeners)[len(*pMessageListeners) - 1]
                *pMessageListeners = (*pMessageListeners)[:len(*pMessageListeners) - 1]

                broadcastMessage(*pMessageListeners, ChatMessage {
                    Username: "SYSTEM",
                    Message: fmt.Sprintf("%s left the chat.", usernameMessage),
                    Closing: false,
                })

                pMessageMutex.Unlock()

                messageListener <- ChatMessage {
                    Username: "",
                    Message: "",
                    Closing: true,
                }

                return
            }

            log.Printf("Error while trying to read a message from the websocket. Error: %en\n", err)
            return
        }

        log.Printf("[%s]: %s\n", string(usernameMessage), string(message))

        pMessageMutex.Lock()

        broadcastMessage(*pMessageListeners, ChatMessage {
            Username: string(usernameMessage),
            Message: string(message),
            Closing: false,
        })

        pMessageMutex.Unlock()
    }
}

func staticFilesHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
    realPath, _ := strings.CutPrefix(pRequest.URL.Path, "/")
    http.ServeFile(pWriter, pRequest, realPath)
}

func main() {
    serverMux := http.NewServeMux()


    messageListeners := [](chan ChatMessage){}
    messageMutex := sync.Mutex {}

    serverMux.HandleFunc("/", indexHandler)
    serverMux.HandleFunc("/chat", chatHandler)
    serverMux.HandleFunc("/chat/socket", func (pWriter http.ResponseWriter, pRequest *http.Request) {
        chatSocketHandler(messageMutex, &messageListeners, pWriter, pRequest)
    })
    serverMux.HandleFunc("/build/", staticFilesHandler)
    serverMux.HandleFunc("/vendor-files/", staticFilesHandler)
    serverMux.HandleFunc("/scripts/", staticFilesHandler)

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
