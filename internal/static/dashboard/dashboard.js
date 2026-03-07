import { changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, changeButtonToFailed, LOGIN_URL_ENDPOINT, DEFAULT_HEADERS, LOGIN_URL, createShortUrl, addCookieBanner, ALLOW_LOGIN, INFO_BANNER_CONTAINER, LOGOUT_URL_ENDPOINT, isLoggedIn, logout } from '../shared.js';


let loggedIn = await isLoggedIn();

addCookieBanner();
if (!loggedIn) {
    window.location.href = LOGIN_URL;
}

document.getElementById("create-short-url").addEventListener("submit", createShortUrl);

document.addEventListener("click", function (event) {
  if (event.target.classList.contains("close-button")) {
    parent = event.target.parentElement;
    parent.classList.add("fade-out");
    parent.addEventListener("animationend", () => {
        parent.remove();
    });
  }  
});

const LOGOUT_BUTTON = document.getElementById("logout-button");
LOGOUT_BUTTON.addEventListener("click",  () => {
    logout();
})