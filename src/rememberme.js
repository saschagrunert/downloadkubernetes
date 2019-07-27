let requests = require('./myrequests.js');

let memoryButtonID = 'remember-me';
let memoryButton = document.getElementById(memoryButtonID);
let cookieRequest = new Request(requests.endpoint("/cookie"));
let forgetRequest = new Request(requests.endpoint("/forget"));

function setForgetButton() {
    memoryButton.removeEventListener('click', remembeMeClickHandler);
    memoryButton.innerText = 'Forget me';
    memoryButton.addEventListener('click', forgetMeClickHandler);
}

function setRememberButton() {
    memoryButton.removeEventListener('click', forgetMeClickHandler);
    memoryButton.innerText = 'Remember me';
    memoryButton.addEventListener('click', remembeMeClickHandler);
}

function remembeMeClickHandler(evt) {
    evt.preventDefault();
    memoryButton.disabled = true;
    fetch(cookieRequest, requests.options())
        .then(() => {
            setForgetButton();
            memoryButton.disabled = false;
        })
        .catch(err => console.log(err))
}

function forgetMeClickHandler(evt) {
    evt.preventDefault()
    memoryButton.disabled = true
    fetch(forgetRequest, requests.options())
    .then(() => {
        document.cookie = 'downloadkubernetes=; expires=Thu, 01 Jan 1970 00:00:00 GMT;'
        location.reload();
        setRememberButton();
        memoryButton.disabled = false;
    })
}

// BpLnfgDsc2WD8F2qNfHK5a84jjJkwz

function initializeMemoryButton() {
    if (document.cookie.includes('downloadkubernetes')) {
        setForgetButton();
        return
    }
    setRememberButton();
}

initializeMemoryButton();

