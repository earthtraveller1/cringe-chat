function chatMain() {
    let messageList = document.getElementById("messageList")
    let messageBox = document.getElementById("messageBox")
    let sendButton = document.getElementById("sendButton")

    let chatSocket = new WebSocket(`ws://${location.host}/chat/socket`)

    chatSocket.onopen = () => {
        chatSocket.send(globalUsername)
    }

    chatSocket.onmessage = (socketMessage) => {
        let messageObject = JSON.parse(socketMessage.data)
        let message = document.createElement("p")
        message.innerText = messageObject.message
        messageList.appendChild(message)
    }
    
    sendButton.onclick = () => {
        chatSocket.send(messageBox.value)
    }
}

chatMain()
