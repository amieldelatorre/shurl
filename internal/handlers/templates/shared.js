/// template the api url
export const API_URL = "{{.apiUrl}}";
export const SHORT_URL_PATH = "api/v1/shorturl";
export const SHORT_URL_ENDPONT = new URL(SHORT_URL_PATH, API_URL);

export const ERROR_CONTAINER_ID = "error-container";
export const ERROR_CONTAINER = document.getElementById(ERROR_CONTAINER_ID);
export const GENERIC_SERVER_ERROR_MESSAGE = "Something went wrong with the server, please try again later";
export const TOO_MANY_REQUESTS_MESSAGE = "Too many requests. Please try again in a few minutes."

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

// function used to simulate long tasks
export function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
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
    setTimeout(() => {
        button.classList.remove(classToRemove);
        fn();
    }, ms);
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

