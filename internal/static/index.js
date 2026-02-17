import { v7 as uuidv7 } from 'https://cdn.jsdelivr.net/npm/uuid@13.0.0/+esm'

const API_URL = "http://localhost:8080";
const SHORT_URL_PATH = "api/v1/shorturl";
const SHORT_URL_ENDPONT = new URL(SHORT_URL_PATH, API_URL);
 
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
        }).catch((err) => {
            console.error("failed to copy", err)
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

    const parent = document.getElementById("index-success-links");
    parent.prepend(successfulLinkCreateDiv);
}

async function onSubmit(event) {
    event.preventDefault();
    //disable button

    const destinationUrlInput = document.getElementById("index-create-url-input");
    const destinationUrl = destinationUrlInput.value;

    const data = JSON.stringify({
        destination_url: destinationUrl
    });

    console.log(data);
    createSuccessfulLinkBox(destinationUrl, "http://localhost/_nothere");
    await fetch(SHORT_URL_ENDPONT, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Idempotency-Key": uuidv7(),
        },
        body: data
    }).then(async response => {
        if (response.ok) {
            destinationUrlInput.value = "";
            // change button to success
            // show url
                // copy to clipboard
            // enable button on closing pop up
        } else {
            // show error
            //enable button
        }

    }).catch(error => {
        //mention error somehow
        //enable button
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
