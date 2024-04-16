function chatMain() {
    let messageList = document.getElementById("messageList")
    let messageBox = document.getElementById("messageBox")
    let sendButton = document.getElementById("sendButton")
    let messageForm = document.getElementById("messageForm")

    let chatSocket = new WebSocket(`ws://${location.host}/chat/socket`)

    chatSocket.onopen = () => {
        chatSocket.send(globalUsername)
    }

    chatSocket.onmessage = (socketMessage) => {
        let messageObject = JSON.parse(socketMessage.data)
        let message = document.createElement("p")
        let username = document.createElement("strong")
        username.appendChild(document.createTextNode(`[${messageObject.username}]: `))
        message.appendChild(username)
        message.appendChild(document.createTextNode(messageObject.message))

        messageList.appendChild(message)
    }
    
    messageForm.addEventListener("submit", (event) => {
        event.preventDefault()
        chatSocket.send(messageBox.value)
    })
}

chatMain()
