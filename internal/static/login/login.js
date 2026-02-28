import { changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, changeButtonToFailed, LOGIN_URL_ENDPOINT, DEFAULT_HEADERS, HOME_URL, sleep, addCookieBanner } from '../shared.js';


const EMAIL_INPUT = document.getElementById("email");
const PASSWORD_INPUT = document.getElementById("password");
const HEADER_X_AUTH_METHOD_WANTED = "X-Auth-Method-Wanted";
const HEADER_X_AUTH_METHOD_WANTED_COOKIE = "cookie";


async function onSubmit(event) {
    event.preventDefault();
    const submittingButton = event.submitter;
    changeButtonToLoading(submittingButton);

    let email = EMAIL_INPUT.value.trim();
    let password = PASSWORD_INPUT.value;

    const data = {
        email: email,
        password: password,
    };

    let result = await fetchWithRetry(
        LOGIN_URL_ENDPOINT,
        "POST",
        {
            ...DEFAULT_HEADERS,
            [HEADER_X_AUTH_METHOD_WANTED]: HEADER_X_AUTH_METHOD_WANTED_COOKIE
        },
        JSON.stringify(data)
    )

    if (!result.isError) {
        changeButtonToSuccess(submittingButton, () => {
            changeButtonToNormal(submittingButton, BUTTON_NORMAL_TEXT);
        });

        await sleep(500);
        window.location.href = HOME_URL;
        return;
    }

    // Chose not to handle timeout explicitly, it should be retryable anyway and means something is wrong with the server.
    if (result.isJson && result.json)
        NOTIFICATION_CONTAINER.prepend(createErrorBox([result.json.error]));
    else
        NOTIFICATION_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));

    changeButtonToFailed(submittingButton, () => {
        changeButtonToNormal(submittingButton, BUTTON_NORMAL_TEXT);
    });
    return;
}


document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("login-form").addEventListener("submit", onSubmit);
});

document.addEventListener("click", function (event) {
  if (event.target.classList.contains("close-button")) {
    parent = event.target.parentElement;
    parent.classList.add("fade-out");
    parent.addEventListener("animationend", () => {
        parent.remove();
    });
  }  
})

addCookieBanner();

// TODO: Check if logged in and is valid and redirect