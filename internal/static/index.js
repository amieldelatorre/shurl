function onSubmit(event) {
    event.preventDefault();
}

document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("index-create-url-form").addEventListener("submit", onSubmit);
})