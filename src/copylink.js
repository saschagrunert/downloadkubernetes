let requests = require('./myrequests.js');
let env = require('./env.js')

function copyLinkEvent(data) {
    console.log("copy link data:", JSON.stringify(data));
    let options = requests.options();
    options['method'] = 'POST';
    options['body'] = JSON.stringify(data);

    // We don't really care about the response unless it's in dev
    fetch(requests.endpoint('/link-copied'), options)
        .then(response => {
            if (!response.ok) {
                console.log("WOW WTF");
            }
            if (response.ok) {
                console.log("UES OSDFK")
            }
            console.log(response);
        });
}

module.exports = {
    event: copyLinkEvent,
}