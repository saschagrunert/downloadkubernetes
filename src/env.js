function isDev() {
    return window.location.hostname === "localhost"
}

module.exports = {isDev};