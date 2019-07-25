let requests = require('./myrequests.js');
let env = require('./env.js')

function copyLinkEvent(data) {
    let options = requests.options();
    options['method'] = 'POST';
    options['body'] = JSON.stringify(data);

    // We don't really care about the response unless it's in dev
    return fetch(requests.endpoint('/link-copied'), options)
        .then(response => {
            if (env.isDev()) {
                console.log(response);
            }
        });
}

module.exports = {
    event: copyLinkEvent,
}