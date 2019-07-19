let requests = require('./myrequests.js');


// if we have a cookie then let's ask for recents
function getRecents() {
    let cookies = document.cookie.split(';').filter(cookie => cookie.trim().startsWith('downloadkubernetes='));
    if (cookies.length != 1) {
        return
    }

    let recentRequest = new Request(requests.endppoints('/recent-downloads'))

    fetch(recentRequest, )
}