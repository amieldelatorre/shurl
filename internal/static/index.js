import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'
import { createCloseButton, changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, ERROR_CONTAINER, changeButtonToFailed, SHORT_URL_ENDPONT, DEFAULT_HEADERS, HEADER_IDEMPOTENCY_KEY, TIMEOUT_IDS } from './shared.js';

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
            TIMEOUT_IDS.push(setTimeout(() => {
                copyButton.textContent = "copy";
            }, 5000));
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
