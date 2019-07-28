let env = require('./env.js');

function options() {
    let out = {
        credentials: 'include',
    };
    if (env.isDev()) {
        out['mode'] = 'cors';
    }
    return out;
}

// URL allows us to test in dev.
function URL(endpoint) {
    if (env.isDev()) {
        return "http://localhost:9999/app" + endpoint;
    }
    return "/app" + endpoint;
}

module.exports = {
    endpoint: URL,
    options: options,
};