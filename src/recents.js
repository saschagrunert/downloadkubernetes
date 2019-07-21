let requests = require('./myrequests.js');

// if we have a cookie then let's ask for recents
function getRecents() {
    let cookies = document.cookie.split(';').filter(cookie => cookie.trim().startsWith('downloadkubernetes='));
    if (cookies.length != 1) {
        return;
    }

    let recentRequest = new Request(requests.endpoint('/recent-downloads'));

    fetch(recentRequest, requests.options()).then((response) => {
        if (!response.ok) {
            return;
        }
        let body = readEntireStream(response.body.getReader())
        console.log("recents", body);
    })
}

// readEntireStream reads the whole reader and returns the contents

/**
 * @param {ReadableStreamDefaultReader} [reader]
 */
function readEntireStream(reader) {
    let partial = ""
    reader.read().then(function processValue({done, value}) {
        if (done) {
            partial += value;
            return partial
        }
        partial += value;
        return reader.read().then(processValue);
    })
}


module.exports = {
    "fetch": getRecents,
};
