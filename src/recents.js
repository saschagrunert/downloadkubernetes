let requests = require('./myrequests.js');

// if we have a cookie then let's ask for recents
function getRecents() {
    let cookies = document.cookie.split(';').filter(cookie => cookie.trim().startsWith('downloadkubernetes='));
    if (cookies.length != 1) {
        return;
    }

    let recentRequest = requests.endpoint('/recent-downloads');

    let options = requests.options();
    options['cache'] = 'no-cache';
    fetch(recentRequest, requests.options())
        .then((response) => {
            if (!response.ok) {
                console.log(response);
                return;
            }
            response.json().then(data => console.log(data));
        })
}

function doSomething(data) {
    console.log(JSON.parse(data))
}

module.exports = {
    "fetch": getRecents,
};
