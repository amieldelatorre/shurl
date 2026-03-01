import { changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, changeButtonToFailed, LOGIN_URL_ENDPOINT, DEFAULT_HEADERS, HOME_URL, sleep, addCookieBanner, ALLOW_LOGIN, INFO_BANNER_CONTAINER, LOGOUT_URL_ENDPOINT, isLoggedIn } from '../shared.js';

const LOGIN_FORM = document.getElementById("login-form");
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

async function checkLoggedin() {
    if (!(await isLoggedIn())) {
        LOGIN_FORM.inert = true;
        let result = await fetchWithRetry(
            LOGOUT_URL_ENDPOINT,
            "POST",
            {},
            {}
        )

        if (!result.isError) {
            LOGIN_FORM.inert = false;
            return
        }

        // Chose not to handle timeout explicitly, it should be retryable anyway and means something is wrong with the server.
        if (result.isJson && result.json)
            NOTIFICATION_CONTAINER.prepend(createErrorBox(["Error trying to invalidate expired token", result.json.error, "Please try logging in again to refresh. If error persists, please try again later"]));
        else
            NOTIFICATION_CONTAINER.prepend(createErrorBox(["Error trying to invalidate expired token", "Please try logging in again to refresh. If error persists, please try again later"]));

        LOGIN_FORM.inert = false;
        return;
    }

    window.location.href = HOME_URL;
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
await checkLoggedin();

if (!ALLOW_LOGIN) {
    LOGIN_FORM.inert = true;

    const loginDisabledBanner = document.createElement("div");
    loginDisabledBanner.id = "login-disabled-banner";
    loginDisabledBanner.classList.add("login-disabled-banner");
    loginDisabledBanner.classList.add("content-disabled-banner");

    const loginDisabledBannerText = document.createElement("p");
    loginDisabledBannerText.classList.add("login-disabled-banner-text");
    loginDisabledBannerText.innerHTML = `Log in has been disabled by the administrator. <a href="${HOME_URL}">Go to home</a>`;
    loginDisabledBanner.append(loginDisabledBannerText);

    INFO_BANNER_CONTAINER.append(loginDisabledBanner);
}
