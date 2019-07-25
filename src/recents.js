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
                return;
            }
            response.json().then(data => {
                document.querySelectorAll('.recent-marker').forEach(marker => {
                    marker.remove()
                });
                data.map(link => {
                    let tbody = document.querySelector('tbody');
                    let el = document.querySelector('[href="'+link+'"]');
                    el.insertAdjacentElement('afterend', recentMarker());
                    let row = el.parentElement.closest('tr');
                    tbody.insertBefore(row, tbody.firstChild);
                });
            });
        })
}

// Add a thing that marks a download as recent
function recentMarker() {
    let span = document.createElement('span');
    span.classList.add('icon', 'recent-marker');

    let i = document.createElement('i');
    i.classList.add('fas', 'fa-star');
    span.appendChild(i)
    return span
}

module.exports = {
    "fetch": getRecents,
};
