/**
 * MIGHT BE BROKEN UP INTO MULTIPLE FILES IN THE FUTURE
 */

import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'


/// template the api url
const API_URL = "{{.apiUrl}}";
const SHORT_URL_PATH = "api/v1/shorturl";
const SHORT_URL_ENDPONT = new URL(SHORT_URL_PATH, API_URL);

const ERROR_CONTAINER_ID = "error-container";
const ERROR_CONTAINER = document.getElementById(ERROR_CONTAINER_ID);
const GENERIC_SERVER_ERROR_MESSAGE = "Something went wrong with the server, please try again later";
const TOO_MANY_REQUESTS_MESSAGE = "Too many requests. Please try again in a few minutes."

const SUCCESS_BUTTON_CLASS = "success-button";
const ERROR_BUTTON_CLASS = "error-button";
const BUTTON_NORMAL_TEXT = "Submit";
const DEFAULT_BUTTON_DISABLED_AFTER_SUBMIT_MS = 1500;

const CONTENT_TYPE_JSON = "application/json";
const HEADER_CONTENT_TYPE = "Content-Type";
const HEADER_IDEMPOTENCY_KEY = "X-Idempotency-Key";
const DEFAULT_HEADERS = {
    [HEADER_CONTENT_TYPE]: CONTENT_TYPE_JSON
}


class FetchResponse {
    json;
    body;
    isJson;
    statusCode;
    isError;
    error;
}

function createCloseButton() {
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
 
function createSuccessfulLinkBox(destinationUrl, shortUrl) {
    const successfulLinkCreateDiv = document.createElement("div");
    successfulLinkCreateDiv.classList.add("successful-link-create");

    const successfulLinkCreateHeader = document.createElement("h4");
    successfulLinkCreateHeader.classList.add("successful-link-create-header");
    successfulLinkCreateHeader.textContent = "Successfully created short url!";

    const shortUrlCreated = document.createElement("div");
    shortUrlCreated.classList.add("short-url-created");

    const shortUrlP = document.createElement("p");
    shortUrlP.textContent = shortUrl;

    const copyButton = document.createElement("button");
    copyButton.classList.add("copy");
    copyButton.textContent = "copy";
    copyButton.addEventListener("click", () => {
        const neighbourShortUrl = copyButton.previousElementSibling.textContent;
        navigator.clipboard.writeText(neighbourShortUrl).then(() => {
            copyButton.textContent = "copied!";
            setTimeout(() => {
                copyButton.textContent = "copy";
            }, 5000);
        }).catch((error) => {
            console.error("failed to copy", error);
        });
    })

    const arrowImg = document.createElement("img");
    arrowImg.classList.add("arrow-to-destination-svg");
    arrowImg.src = "/_/assets/right-arrow-circle-svgrepo-com.svg";

    const destinationUrlSpan = document.createElement("span");
    destinationUrlSpan.classList.add("destination-url");
    destinationUrlSpan.innerHTML = destinationUrl;

    shortUrlCreated.appendChild(shortUrlP);
    shortUrlCreated.appendChild(copyButton);
    successfulLinkCreateDiv.appendChild(successfulLinkCreateHeader);
    successfulLinkCreateDiv.appendChild(createCloseButton());
    successfulLinkCreateDiv.appendChild(shortUrlCreated);
    successfulLinkCreateDiv.appendChild(arrowImg);
    successfulLinkCreateDiv.appendChild(destinationUrlSpan);

    return successfulLinkCreateDiv;
}

function createErrorBox(messages) {
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
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
}

function clearChildren(node) {
    while (node.firstChild) {
        node.removeChild(node.firstChild);
    }
}

function createSpinner() {
    const spinnerSpan = document.createElement("span");
    spinnerSpan.classList.add("spinner")
    return spinnerSpan
}

function changeButtonToLoading(button) {
    button.innerText = null;
    button.disabled = true;
    button.appendChild(createSpinner());
}

function changeButtonToNormal(button, textContent) {
    clearChildren(button);
    button.textContent = textContent;
    button.disabled = false;
}

function changeButtonToSuccess(button, fn) {
    clearChildren(button);
    button.disabled = true;

    const img = document.createElement("img");
    img.classList.add("success-svg");
    img.src = "/_/assets/check-circle-svgrepo-com.svg";
    button.appendChild(img);

    button.classList.add(SUCCESS_BUTTON_CLASS);
    removeClassAfterTimeout(button, SUCCESS_BUTTON_CLASS, DEFAULT_BUTTON_DISABLED_AFTER_SUBMIT_MS, fn);
}

function changeButtonToFailed(button, fn) {
    clearChildren(button);
    button.disabled = true;

    const img = document.createElement("img");
    img.classList.add("error-svg");
    img.src = "/_/assets/error-svgrepo-com.svg";
    button.appendChild(img);
    
    button.classList.add(ERROR_BUTTON_CLASS);
    removeClassAfterTimeout(button, ERROR_BUTTON_CLASS, DEFAULT_BUTTON_DISABLED_AFTER_SUBMIT_MS, fn);
}

function removeClassAfterTimeout(button, classToRemove, ms, fn) {
    // using set timeout here and not sleep because this should happen in the background
    setTimeout(() => {
        button.classList.remove(classToRemove);
        fn();
    }, ms);
}


async function onSubmit(event) {
    event.preventDefault();
    const submittingButton = event.submitter;
    changeButtonToLoading(submittingButton);

    const destinationUrlInput = document.getElementById("index-create-url-input");
    const destinationUrl = destinationUrlInput.value;

    const data = JSON.stringify({
        destination_url: destinationUrl
    });

    let result = await fetchWithRetry(
        SHORT_URL_ENDPONT, 
        "POST",
        {
            ...DEFAULT_HEADERS,
            [HEADER_IDEMPOTENCY_KEY]: uuidv7(),
        },
        data
    )

    if (!result.isError) {
        const successfulLinkCreateDiv = createSuccessfulLinkBox(result.json.destination_url, result.json.url);
        const parent = document.getElementById("index-success-links");
        parent.prepend(successfulLinkCreateDiv);

        changeButtonToSuccess(submittingButton, () => {
            changeButtonToNormal(submittingButton, BUTTON_NORMAL_TEXT);
        });

        destinationUrlInput.value = "";
        return;
    }

    if (result.isJson)
        ERROR_CONTAINER.prepend(createErrorBox([result.error.error]));
    else
        ERROR_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));

    changeButtonToFailed(submittingButton, () => {
        changeButtonToNormal(submittingButton, BUTTON_NORMAL_TEXT);
    });
    return;
}

async function fetchWithRetry(url, method, headers, body, maxAttempts = 3, retryBaseDelay = 150) {
    let result = new FetchResponse();

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
        try {
            let response = await fetch(url, {
                method: method,
                headers: headers,
                body: body
            });

            if (response.ok) {
                result.isError = false;
                result.statusCode = response.status;
                result = addResponseBody(result, response);
                return result;
            } 
            result.isError = true;
            
            if (!isRetryable(response.statusCode)) {
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

async function addResponseBody(result, fetchResult) {
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

function isRetryable(statusCode) {
    if (statusCode == 429) {
        return true;
    }
}

document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("index-create-url-form").addEventListener("submit", onSubmit);
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
