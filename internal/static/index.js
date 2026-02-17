import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'


/// template the api url
const API_URL = "http://localhost:8080";
const SHORT_URL_PATH = "api/v1/shorturl";
const SHORT_URL_ENDPONT = new URL(SHORT_URL_PATH, API_URL);

const INDEX_CREATE_URL_SUBMIT_BUTTON_ID = "index-create-url-submit-button";
 
function createSuccessfulLinkBox(destinationUrl, shortUrl) {
    const successfulLinkCreateDiv = document.createElement("div");
    successfulLinkCreateDiv.classList.add("successful-link-create");

    const successfulLinkCreateHeader = document.createElement("h4");
    successfulLinkCreateHeader.classList.add("successful-link-create-header");
    successfulLinkCreateHeader.textContent = "Successfully created short url!";

    const closeButton = document.createElement("button");
    closeButton.classList.add("close-button");
    closeButton.ariaLabel = "Close";

    const closeSymbol = document.createElement("span");
    closeSymbol.classList.add("close-symbol");
    closeSymbol.setAttribute("aria-hidden", "true");
    closeSymbol.innerHTML = "&times;";

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


    closeButton.appendChild(closeSymbol);
    shortUrlCreated.appendChild(shortUrlP);
    shortUrlCreated.appendChild(copyButton);
    successfulLinkCreateDiv.appendChild(successfulLinkCreateHeader);
    successfulLinkCreateDiv.appendChild(closeButton);
    successfulLinkCreateDiv.appendChild(shortUrlCreated);
    successfulLinkCreateDiv.appendChild(arrowImg);
    successfulLinkCreateDiv.appendChild(destinationUrlSpan);

    return successfulLinkCreateDiv;
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
    // <span id="login-form-submit-button-loader" class="login-form-submit-button-loader" hidden></span>
    const spinnerSpan = document.createElement("span");
    spinnerSpan.classList.add("spinner")
    return spinnerSpan
}

function changeButtonToLoading(buttonId) {
    const button = document.getElementById(buttonId);
    button.innerText = null;
    // button.innerText = "Loading...";
    button.disabled = true;
}

function changeButtonToNormal(button, textContent) {
    clearChildren(button);
    button.textContent = textContent;
    button.disabled = false;
}


async function onSubmit(event) {
    event.preventDefault();
    changeButtonToLoading(INDEX_CREATE_URL_SUBMIT_BUTTON_ID);
    const submittingButton = event.submitter;
    submittingButton.appendChild(createSpinner());



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
        } else if (response.status == 401) {
            // show error pop up
            console.log("Bad req");
        } else {
            // show error pop up
            console.log("Something did not go well");
        }

    }).catch(error => {
        // show error pop up 
        console.log(error);
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
