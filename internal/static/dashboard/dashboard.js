import { createSuccessBox, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, LOGIN_URL, createShortUrl, addCookieBanner, USER_SHORT_URL_ENDPONT, isLoggedIn, logout, clearChildren, USER_SHORT_URL_ENDPONT_WITH_ID } from '../shared.js';


let loggedIn = await isLoggedIn();
const SHORT_URLS_TABLE_HEAD = document.getElementById("short-urls-table-heading");
const SHORT_URLS_TABLE_BODY = document.getElementById("short-urls-table-body");
let CURRENT_SHORT_URLS = [];

addCookieBanner();
if (!loggedIn) {
    window.location.href = LOGIN_URL;
}

document.getElementById("create-short-url").addEventListener("submit", createShortUrl);

document.addEventListener("click", function (event) {
  if (event.target.classList.contains("close-button")) {
    parent = event.target.parentElement;
    parent.classList.add("fade-out");
    parent.addEventListener("animationend", () => {
        parent.remove();
    });
  }  
});

const LOGOUT_BUTTON = document.getElementById("logout-button");
LOGOUT_BUTTON.addEventListener("click",  () => {
    logout();
})

function renderTableHead() {
  const tableHeaders = ["Short URL", "Destination URL", "Created", "Expiry", "Actions"];
  
  const tableHeadersElem = document.createElement("tr");

  for (let h of tableHeaders) {
    let heading = document.createElement("th");
    heading.textContent = h;

    tableHeadersElem.appendChild(heading);
  }

  SHORT_URLS_TABLE_HEAD.appendChild(tableHeadersElem);
}

async function deleteShortUrl(id) {
  let url = USER_SHORT_URL_ENDPONT_WITH_ID(id);
  let result = await fetchWithRetry(
    url,
    "DELETE",
  );

  if (!result.isError) {
    NOTIFICATION_CONTAINER.prepend(createSuccessBox(["Successfully deleted short url", "The link may still work for a while due to caching."]));

    const idx = CURRENT_SHORT_URLS.findIndex(item => item.id === id);
    if (idx !== -1) {
      CURRENT_SHORT_URLS.splice(idx, 1);
    }

    renderShortUrls(CURRENT_SHORT_URLS);
    return
  }

  if (result.isJson && result.json)
    NOTIFICATION_CONTAINER.prepend(createErrorBox(result.json.errors));
  else
    NOTIFICATION_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));
  return;
}

function renderShortUrls(items) {
  clearChildren(SHORT_URLS_TABLE_HEAD);
  renderTableHead();
  clearChildren(SHORT_URLS_TABLE_BODY);
  
  const dateOptions = {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: true
  };

  for (let h of items) {
    let r = document.createElement("tr");
    
    let shortUrl = document.createElement("td");
    shortUrl.textContent = h.url;
    r.appendChild(shortUrl);

    let destinationUrl = document.createElement("td");
    destinationUrl.textContent = h.destination_url;
    r.appendChild(destinationUrl);

    let created = document.createElement("td");
    let createdValue = new Intl.DateTimeFormat(undefined, dateOptions).format(new Date(h.created_at));
    created.textContent = createdValue;
    created.title = h.created_at;
    r.appendChild(created);

    let expiry = document.createElement("td");
    let expiryValue = new Intl.DateTimeFormat(undefined, dateOptions).format(new Date(h.expires_at));
    expiry.textContent = expiryValue;
    expiry.title = h.expires_at;
    r.appendChild(expiry);

    let action = document.createElement("td");
    action.classList.add("table-row-actions");
    let deleteActionBtn = document.createElement("button");
    deleteActionBtn.classList.add("delete-button");
    deleteActionBtn.onclick = async () => {
      await deleteShortUrl(h.id);
    }

    let deleteActionImg = document.createElement("img");
    deleteActionImg.classList.add("delete-button-img");
    deleteActionImg.src = "/_/assets/rubbish-bin-svgrepo-com.svg"
    deleteActionBtn.appendChild(deleteActionImg);
    action.appendChild(deleteActionBtn);
    r.appendChild(action);

    SHORT_URLS_TABLE_BODY.appendChild(r);
  }
}

export async function getShortUrls(page=1, size=20) {
  let result = await fetchWithRetry(
    USER_SHORT_URL_ENDPONT,
    "GET",
  )

  if (!result.isError) {
    CURRENT_SHORT_URLS = result.json.items;
    return result.json.items;
  }

  // Chose not to handle timeout explicitly, it should be retryable anyway and means something is wrong with the server.
  if (result.isJson && result.json)
    NOTIFICATION_CONTAINER.prepend(createErrorBox(result.json.errors));
  else
    NOTIFICATION_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));
  return;
}

let shortUrls = await getShortUrls();
renderShortUrls(shortUrls);
