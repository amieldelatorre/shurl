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


async function onSubmit(event) {
    event.preventDefault();
    const submittingButton = event.submitter;
    changeButtonToLoading(submittingButton);

    const destinationUrlInput = document.getElementById("index-create-url-input");
    const destinationUrl = destinationUrlInput.value;

    const data = JSON.stringify({
        destination_url: destinationUrl
    });

    // Simulate network wait
    await sleep(1000);
    await fetch(SHORT_URL_ENDPONT, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Idempotency-Key": uuidv7(),
        },
        body: data
    }).then(async response => {
        if (response.ok) {
            const v = await response.json();
            const successfulLinkCreateDiv = createSuccessfulLinkBox(v.destination_url, v.url);
            const parent = document.getElementById("index-success-links");
            parent.prepend(successfulLinkCreateDiv);

            destinationUrlInput.value = "";
            changeButtonToNormal(submittingButton, "Submit");
        } else if (response.status == 400) {
            const v = await response.json()
            ERROR_CONTAINER.prepend(createErrorBox([v.error]));
            changeButtonToNormal(submittingButton, "Submit");
            return
        } else {
            ERROR_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));
            changeButtonToNormal(submittingButton, "Submit");
            return
        }

    }).catch(error => {
        console.log(error);
        ERROR_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));
        changeButtonToNormal(submittingButton, "Submit");
        return
    })
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
