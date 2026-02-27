import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'
import { changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, ERROR_CONTAINER, changeButtonToFailed, USER_URL_ENDPONT, DEFAULT_HEADERS, HEADER_IDEMPOTENCY_KEY, API_URL, LOGIN_URL, sleep, ALLOW_REGISTRATION, ALLOW_LOGIN, ALLOW_ANONYMOUS } from '../shared.js';

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
        ERROR_CONTAINER.prepend(createErrorBox(validationErrors));
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
        ERROR_CONTAINER.prepend(createErrorBox([result.json.error]));
    else
        ERROR_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));

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

if (!ALLOW_REGISTRATION && ALLOW_ANONYMOUS) 
    window.location.href = API_URL;
else if (!ALLOW_REGISTRATION && ALLOW_LOGIN)
    window.location.href = LOGIN_URL;
