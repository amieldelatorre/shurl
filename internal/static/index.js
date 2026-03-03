import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'
import { createCloseButton, changeButtonToLoading, changeButtonToSuccess, changeButtonToNormal, BUTTON_NORMAL_TEXT, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, changeButtonToFailed, SHORT_URL_ENDPONT, DEFAULT_HEADERS, HEADER_IDEMPOTENCY_KEY, TIMEOUT_IDS, addCookieBanner, isLoggedIn, sleep, ALLOW_ANONYMOUS, LOGIN_URL, ALLOW_LOGIN, ALLOW_REGISTRATION, INFO_BANNER_CONTAINER } from './shared.js';

const CREATE_URL_FORM_ID = "index-create-url-form";
const CREATE_URL_FORM = document.getElementById(CREATE_URL_FORM_ID);


function createSuccessfulLinkBox(destinationUrl, shortUrl, expires_at) {
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

    const expiryP = document.createElement("p");
    expiryP.textContent = `Expiry: ${expires_at}`;

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
    successfulLinkCreateDiv.appendChild(expiryP);
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
        const successfulLinkCreateDiv = createSuccessfulLinkBox(result.json.destination_url, result.json.url, result.json.expires_at);
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
        NOTIFICATION_CONTAINER.prepend(createErrorBox([result.json.error]));
    else
        NOTIFICATION_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));

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
});

let loggedIn = await isLoggedIn();

addCookieBanner();
if (!loggedIn && !ALLOW_ANONYMOUS && !ALLOW_LOGIN && !ALLOW_REGISTRATION) {
    CREATE_URL_FORM.inert = true;


    const applicationReadonlyBanner = document.createElement("div");
    applicationReadonlyBanner.id = "application-readonly-banner";
    applicationReadonlyBanner.classList.add("application-readonly-banner");
    applicationReadonlyBanner.classList.add("content-disabled-banner");

    const applicationReadonlyBannerText = document.createElement("p");
    applicationReadonlyBannerText.classList.add("application-readonly-banner-text");
    applicationReadonlyBannerText.innerHTML = `Administrator has disabled login, registration and and anonymous short url creation. Already logged in users will still be able to create short urls for a while. Existing short url redirects will still work.`;
    applicationReadonlyBanner.append(applicationReadonlyBannerText);

    INFO_BANNER_CONTAINER.append(applicationReadonlyBanner);
} else if (!loggedIn && !ALLOW_ANONYMOUS) {
    CREATE_URL_FORM.inert = true;
    NOTIFICATION_CONTAINER.prepend(createErrorBox(["Not logged in, redirecting to login page in 1 second"]));
    await sleep(1000);
    window.location.href = LOGIN_URL;
}

