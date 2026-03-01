import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'
import { changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, changeButtonToFailed, USER_URL_ENDPONT, DEFAULT_HEADERS, HEADER_IDEMPOTENCY_KEY, LOGIN_URL, sleep, ALLOW_REGISTRATION, addCookieBanner, INFO_BANNER_CONTAINER, HOME_URL, isLoggedIn } from '../shared.js';

const SIGNUP_FORM = document.getElementById("signup-form");
const EMAIL_INPUT = document.getElementById("email");
const USERNAME_INPUT = document.getElementById("username");
const PASSWORD_INPUT = document.getElementById("password");
const CONFIRM_PASSWORD_INPUT = document.getElementById("confirmpassword");

function validateSignup(data) {
    // username and email are already validated by html
    // validate passwords
    if (data.password !== data.confirm_password) {
        return ["Password does not match Confirm Password"];    
    }

    return [];
}

async function onSubmit(event) {
    event.preventDefault();
    const submittingButton = event.submitter;
    changeButtonToLoading(submittingButton);

    let email = EMAIL_INPUT.value.trim();
    let username = USERNAME_INPUT.value.trim();
    let password = PASSWORD_INPUT.value;
    let confirmPassword = CONFIRM_PASSWORD_INPUT.value;

    const data = {
        email: email,
        username: username,
        password: password,
        confirm_password: confirmPassword
    };

    let validationErrors = validateSignup(data);
    if (validationErrors.length > 0) {
        changeButtonToFailed(submittingButton, () => {
            changeButtonToNormal(submittingButton, BUTTON_NORMAL_TEXT);
        });
        NOTIFICATION_CONTAINER.prepend(createErrorBox(validationErrors));
        return;
    }

    let result = await fetchWithRetry(
        USER_URL_ENDPONT,
        "POST",
        {
            ...DEFAULT_HEADERS,
            [HEADER_IDEMPOTENCY_KEY]: uuidv7(),
        },
        JSON.stringify(data)
    )

    if (!result.isError) {
        changeButtonToSuccess(submittingButton, () => {
            changeButtonToNormal(submittingButton, BUTTON_NORMAL_TEXT);
        });

        await sleep(500);
        window.location.href = LOGIN_URL;
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
    document.getElementById("signup-form").addEventListener("submit", onSubmit);
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

if (!ALLOW_REGISTRATION) {
    SIGNUP_FORM.inert = true;

    const signupDisabledBanner = document.createElement("div");
    signupDisabledBanner.id = "signup-disabled-banner";
    signupDisabledBanner.classList.add("signup-disabled-banner");
    signupDisabledBanner.classList.add("content-disabled-banner");

    const signupDisabledBannerText = document.createElement("p");
    signupDisabledBannerText.classList.add("signup-disabled-banner-text");
    signupDisabledBannerText.innerHTML = `Sign up has been disabled by the administrator. <a href="${LOGIN_URL}">Go to login</a>`;
    signupDisabledBanner.append(signupDisabledBannerText);

    INFO_BANNER_CONTAINER.append(signupDisabledBanner);
}

if (await isLoggedIn()) {
    window.location.href = HOME_URL;
}
