function isDev() {
    return window.location.hostname === "localhost"
}

function options() {
    let out = {
        credentials: 'include',
    }
    if (isDev()) {
        out['mode'] = 'cors'
    }
    return out
}

// URL allows us to test in dev.
function URL(endpoint) {
    if ((isDev())) {
        return "http://localhost:9999" + endpoint
    }
    return endpoint
}

module.exports = {
    endpoint: URL,
    options: options,
};