// template environment variables
export const ALLOW_REGISTRATION = {{.allowRegistration}};
export const ALLOW_LOGIN = {{.allowLogin}};
export const ALLOW_ANONYMOUS = {{.allowAnonymous}};

/// template the api url
export const API_URL = "{{.apiUrl}}";
export const LOGIN_URL = new URL("_/login", API_URL);
export const HOME_URL = new URL("/", API_URL);
export const DASHBOARD_URL = new URL("_/dashboard", API_URL);

export const SHORT_URL_PATH = "api/v1/shorturl";
export const SHORT_URL_ENDPONT = new URL(SHORT_URL_PATH, API_URL);

export const USER_URL_PATH = "api/v1/user";
export const USER_URL_ENDPONT = new URL(USER_URL_PATH, API_URL);

export const LOGIN_URL_PATH = "api/v1/auth/login";
export const LOGIN_URL_ENDPOINT = new URL(LOGIN_URL_PATH, API_URL);

export const LOGOUT_URL_PATH = "api/v1/auth/logout";
export const LOGOUT_URL_ENDPOINT = new URL(LOGOUT_URL_PATH, API_URL);

export const VALIDATE_URL_PATH = "api/v1/auth/validate";
export const VALIDATE_URL_ENDPOINT = new URL(VALIDATE_URL_PATH, API_URL);

export const INFO_BANNER_CONTAINER_ID = "info-banner";
export const INFO_BANNER_CONTAINER = document.getElementById(INFO_BANNER_CONTAINER_ID);

export const NOTIFICATION_CONTAINER_ID = "notification-container";
export const NOTIFICATION_CONTAINER = document.getElementById(NOTIFICATION_CONTAINER_ID);
export const GENERIC_SERVER_ERROR_MESSAGE = "Something went wrong with the server, please try again later";
export const TOO_MANY_REQUESTS_MESSAGE = "Too many requests. Please try again in a few minutes."

export const PAGE_LOADING_CONTAINER_ID = "page-loading";
export const PAGE_LOADING_CONTAINER = document.getElementById(PAGE_LOADING_CONTAINER_ID);

export const SUCCESS_BUTTON_CLASS = "success-button";
export const ERROR_BUTTON_CLASS = "error-button";
export const BUTTON_NORMAL_TEXT = "Submit";
export const DEFAULT_BUTTON_DISABLED_AFTER_SUBMIT_MS = 1500;

export const CONTENT_TYPE_JSON = "application/json";
export const HEADER_CONTENT_TYPE = "Content-Type";
export const HEADER_IDEMPOTENCY_KEY = "X-Idempotency-Key";
export const DEFAULT_HEADERS = {
    [HEADER_CONTENT_TYPE]: CONTENT_TYPE_JSON
}

export const TIMEOUT_IDS = [];


export class FetchResponse {
    json;
    body;
    isJson;
    statusCode;
    isError;
    error;
}

export function createCloseButton() {
    const closeButton = document.createElement("button");
    closeButton.classList.add("close-button");
    closeButton.ariaLabel = "Close";

    const closeSymbol = document.createElement("span");
    closeSymbol.classList.add("close-symbol");
    closeSymbol.setAttribute("aria-hidden", "true");
    closeSymbol.innerHTML = "&times;";
    
    closeButton.appendChild(closeSymbol);
    return closeButton;
}

export function createErrorBox(messages) {
    const box = document.createElement("div");
    box.classList.add("error-notification");
    box.appendChild(createCloseButton());

    const errList = document.createElement("ul");
    box.appendChild(errList);

    for (let m of messages) {
        let li = document.createElement("li");
        li.textContent = m;
        errList.appendChild(li);
    }

    return box
}

export function createSuccessBox(messages) {
    const box = document.createElement("div");
    box.classList.add("success-notification");
    box.appendChild(createCloseButton());

    const succcessList = document.createElement("ul");
    box.appendChild(succcessList);

    for (let m of messages) {
        let li = document.createElement("li");
        li.textContent = m;
        succcessList.appendChild(li);
    }

    return box
}

// function used to simulate long tasks
export function sleep(ms) {
    return new Promise(resolve => TIMEOUT_IDS.push(setTimeout(resolve, ms)));
}

export function clearAllTimeouts() {
    TIMEOUT_IDS.forEach(id => clearTimeout(id));
}

export function clearChildren(node) {
    while (node.firstChild) {
        node.removeChild(node.firstChild);
    }
}

export function createSpinner() {
    const spinnerSpan = document.createElement("span");
    spinnerSpan.classList.add("spinner")
    return spinnerSpan
}

export function changeButtonToLoading(button) {
    button.innerText = null;
    button.disabled = true;
    button.appendChild(createSpinner());
}

export function changeButtonToNormal(button, textContent) {
    clearChildren(button);
    button.textContent = textContent;
    button.disabled = false;
}

export function changeButtonToSuccess(button, fn) {
    clearChildren(button);
    button.disabled = true;

    const img = document.createElement("img");
    img.classList.add("success-svg");
    img.src = "/_/assets/check-circle-svgrepo-com.svg";
    button.appendChild(img);

    button.classList.add(SUCCESS_BUTTON_CLASS);
    removeClassAfterTimeout(button, SUCCESS_BUTTON_CLASS, DEFAULT_BUTTON_DISABLED_AFTER_SUBMIT_MS, fn);
}

export function changeButtonToFailed(button, fn) {
    clearChildren(button);
    button.disabled = true;

    const img = document.createElement("img");
    img.classList.add("error-svg");
    img.src = "/_/assets/error-svgrepo-com.svg";
    button.appendChild(img);
    
    button.classList.add(ERROR_BUTTON_CLASS);
    removeClassAfterTimeout(button, ERROR_BUTTON_CLASS, DEFAULT_BUTTON_DISABLED_AFTER_SUBMIT_MS, fn);
}

export function removeClassAfterTimeout(button, classToRemove, ms, fn) {
    // using set timeout here and not sleep because this should happen in the background
    TIMEOUT_IDS.push(setTimeout(() => {
        button.classList.remove(classToRemove);
        fn();
    }, ms));
}

export async function fetchWithRetry(url, method, headers, body, maxAttempts = 3, retryBaseDelay = 150, defaultTimeoutMs = 2500) {
    let result = new FetchResponse();

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
        try {
            let response = await fetch(url, {
                method: method,
                headers: headers,
                body: body,
                signal: AbortSignal.timeout(defaultTimeoutMs)
            });

            if (response.ok) {
                result.isError = false;
                result.statusCode = response.status;
                result = addResponseBody(result, response);
                return result;
            } 
            result.isError = true;
            
            if (!isRetryable(response.status)) {
                result.statusCode = response.status;
                result = await addResponseBody(result, response);
                return result;
            }

            // having trouble figuring out how to differentiate between error types, so its all just retryable
            result = addResponseBody(result, response);

            if (attempt != maxAttempts - 1) {
                let delay = retryBaseDelay * Math.pow(2, attempt);
                let jitter = Math.floor(Math.random() * (delay/4));

                await sleep(delay + jitter);
            }

        } catch (error) {
            // having trouble figuring out how to differentiate between error types, so its all just retryable
            result.isError = true;
            result.error = error;
        }
    }
    return result;
}

export async function addResponseBody(result, fetchResult) {
    const contentType = fetchResult.headers.get(HEADER_CONTENT_TYPE);
    if (contentType == CONTENT_TYPE_JSON) {
        result.isJson = true;
        result.json = await fetchResult.json();
    }
    else {
        result.isJson = false;
        result.body = await fetchResult.text();
    }

    return result;
}

export function isRetryable(statusCode) {
    if (statusCode == 429 || statusCode == 500) {
        return true;
    }
    return false;
}

export function addCookieBanner() {
    const cookieBannerKey = "authentication-cookie-banner";
    let cookieBannerAccepted = localStorage.getItem(cookieBannerKey);
    if (cookieBannerAccepted === "accepted") {
        return
    }

    const cookieBanner = document.createElement("div");
    cookieBanner.id = "cookie-banner";
    cookieBanner.classList.add("cookie-banner");

    const cookieBannerText = document.createElement("p");
    cookieBannerText.classList.add("cookie-banner-text");
    cookieBannerText.textContent = "This site uses essential cookies for login and security.";

    cookieBanner.appendChild(cookieBannerText);
    
    const cookieBannerAcceptButton = document.createElement("button");
    cookieBannerAcceptButton.id = "cookie-banner-accept-button";
    cookieBannerAcceptButton.textContent = "Got it";

    cookieBanner.appendChild(cookieBannerAcceptButton);

    cookieBannerAcceptButton.onclick = () => {
        localStorage.setItem(cookieBannerKey, "accepted");
        cookieBanner.remove();
    }

    INFO_BANNER_CONTAINER.append(cookieBanner);
}

export async function isLoggedIn() {
    PAGE_LOADING_CONTAINER.hidden = false;
    
    let result = await fetchWithRetry(
        VALIDATE_URL_ENDPOINT,
        "GET"
    );

    PAGE_LOADING_CONTAINER.hidden = true;
    // everything else, even connection errors is false
    return (!result.isError && result.statusCode == 200);
}

export async function logout() {
    PAGE_LOADING_CONTAINER.hidden = false;
    let result = await fetchWithRetry(
        LOGOUT_URL_ENDPOINT,
        "POST",
        {},
        {}
    );

    if (!result.isError) {
        NOTIFICATION_CONTAINER.prepend(createSuccessBox(["Successfully logged out, redirecting to login page in 1 second"]));
        await sleep(1000);
        window.location.href = LOGIN_URL;
    }

    PAGE_LOADING_CONTAINER.hidden = true;
}
