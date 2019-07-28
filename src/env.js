var Env = "";

function environment() {
    if (Env != "") {
        return Env;
    }
    if (window.location.hostname != "localhost"){
        return "production";
    }
    if (window.location.port === "8008") {
        return "docker";
    }
    return "dev";
}

function isDev() {
    return environment() === "dev"
}

module.exports = {isDev};