import { createSuccessBox, fetchWithRetry, createErrorBox, GENERIC_SERVER_ERROR_MESSAGE, NOTIFICATION_CONTAINER, LOGIN_URL, createShortUrl, addCookieBanner, USER_SHORT_URL_ENDPONT, isLoggedIn, logout, clearChildren, USER_SHORT_URL_ENDPONT_WITH_ID, TIMEOUT_IDS } from '../shared.js';


let loggedIn = await isLoggedIn();
const SHORT_URLS_TABLE_HEAD = document.getElementById("short-urls-table-heading");
const SHORT_URLS_TABLE_BODY = document.getElementById("short-urls-table-body");
let CURRENT_SHORT_URLS = {};

function getPageUrlParam() {
  let urlParams = new URLSearchParams(window.location.search);
  let page = Number(urlParams.get("page"));
  if (!Number.isInteger(page) || page === 0)
    page = 1;
  return page;
}

function getSizeUrlParam() {
  let urlParams = new URLSearchParams(window.location.search);
  let size = Number(urlParams.get("size"));
  if (!Number.isInteger(size) || size === 0)
    size = 20;
  return size;
}

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
  const tableHeaders = ["#", "Short URL", "Destination URL", "Created", "Expiry", "Actions"];
  
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

    const idx = CURRENT_SHORT_URLS.items.findIndex(item => item.id === id);
    if (idx !== -1) {
      CURRENT_SHORT_URLS.items.splice(idx, 1);
    }

    CURRENT_SHORT_URLS.total -= 1;

    renderShortUrls(CURRENT_SHORT_URLS, CURRENT_SHORT_URLS.total);

    // const allTableBodyRows = document.querySelectorAll("tbody tr");
    // const allShortUrlsCounter = document.querySelectorAll("p.short-urls-counter");
    // allShortUrlsCounter.forEach((p) => {
    //   let start = shortUrls.page * shortUrls.size - shortUrls.size + 1;
    //   let end = start + allTableBodyRows.length;
    //   p.textContent = `${start} - ${end} of ${shortUrls.total}`
    // });

    return
  }

  if (result.isJson && result.json)
    NOTIFICATION_CONTAINER.prepend(createErrorBox(result.json.errors));
  else
    NOTIFICATION_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));
  return;
}

function renderShortUrls(shortUrls) {
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
  let elementNum = shortUrls.page * shortUrls.size - shortUrls.size + 1;
  for (let h of shortUrls.items) {
    let r = document.createElement("tr");

    let num = document.createElement("td");
    num.textContent = elementNum;
    elementNum++;
    r.appendChild(num);
    
    let shortUrl = document.createElement("td");
    shortUrl.textContent = h.url;
    shortUrl.title = "click to copy";
    shortUrl.classList.add("pointer-cursor");
    shortUrl.addEventListener("click", () => {
      navigator.clipboard.writeText(h.url).then(() => {
        shortUrl.textContent = "copied!";
        TIMEOUT_IDS.push(setTimeout(() => {
          shortUrl.textContent = h.url;
        }, 5000));
      }).catch((error) => {
           NOTIFICATION_CONTAINER.prepend(createErrorBox(["Failed to copy url"]));
           console.log(error);
        });
    });
    r.appendChild(shortUrl);

    let destinationUrl = document.createElement("td");
    destinationUrl.textContent = h.destination_url;
    destinationUrl.title = h.destination_url;
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
    deleteActionBtn.title = "delete";
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

  const allShortUrlsCounter = document.querySelectorAll("p.short-urls-counter");
  allShortUrlsCounter.forEach((p) => {
    let currlength = shortUrls.page * shortUrls.size;
    let start = currlength - shortUrls.size + 1;

    if (currlength > shortUrls.total)
      currlength = shortUrls.total;
    if (shortUrls.total === 0)
      start = 0;

    p.textContent = `${start} - ${currlength} of ${shortUrls.total}`
  });
}

export async function getShortUrls(page=1, size=20) {
  let url = USER_SHORT_URL_ENDPONT();
  url.searchParams.set("page", page);
  url.searchParams.set("size", size);
  let result = await fetchWithRetry(
    url,
    "GET",
  )

  if (!result.isError) {
    CURRENT_SHORT_URLS = result.json;
    return result.json;
  }

  // Chose not to handle timeout explicitly, it should be retryable anyway and means something is wrong with the server.
  if (result.isJson && result.json)
    NOTIFICATION_CONTAINER.prepend(createErrorBox(result.json.errors));
  else
    NOTIFICATION_CONTAINER.prepend(createErrorBox([GENERIC_SERVER_ERROR_MESSAGE]));
  return;
}

function togglePreviousButtons() {
  const allPreviousButtons = document.querySelectorAll("button.table-nav-previous-button");
  allPreviousButtons.forEach(btn => {
    if (CURRENT_SHORT_URLS.page <= 1) {
      btn.disabled = true;
    } else {
      btn.disabled = false;
    }
  });
}

function toggleNextButtons() {
  const allNextButtons = document.querySelectorAll("button.table-nav-next-button");
  allNextButtons.forEach(btn => {
    if (!CURRENT_SHORT_URLS.next) {
      btn.disabled = true;
    } else {
      btn.disabled = false;
    }
  });
}

let shortUrls = await getShortUrls(getPageUrlParam(), getSizeUrlParam());
renderShortUrls(shortUrls);
togglePreviousButtons();
toggleNextButtons();


const allPreviousButtons = document.querySelectorAll("button.table-nav-previous-button");
allPreviousButtons.forEach(btn => {
  btn.addEventListener("click", async () => {
    let previousPage = getPageUrlParam() - 1;
    let res = await getShortUrls(previousPage, getSizeUrlParam());
    renderShortUrls(res);

    let newUrl = new URL(window.location.href);
    newUrl.searchParams.set("page", previousPage);
    history.pushState(null, null, newUrl);

    togglePreviousButtons();
    toggleNextButtons();
  });
});

const allNextButtons = document.querySelectorAll("button.table-nav-next-button");
allNextButtons.forEach(btn => {
  btn.addEventListener("click", async () => {
    let nextPage = getPageUrlParam() + 1;
    let res = await getShortUrls(nextPage, getSizeUrlParam());
    renderShortUrls(res);

    let newUrl = new URL(window.location.href);
    newUrl.searchParams.set("page", nextPage);
    history.pushState(null, null, newUrl);

    togglePreviousButtons();
    toggleNextButtons();
  });
});