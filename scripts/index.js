let globalUsername = null

htmx.on("#join-form", "submit", (event) => {
    let formUsername = document.getElementById("form-username")
    globalUsername = formUsername.value
})
