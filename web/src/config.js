// Reads and forms basic config for WebUI

import DEFINE from "./utils/define"
import log from "./utils/log"

const logger = log.withPrefix("[config]");

const Host = function(port) {
    return "//" + window.location.hostname + ":" + port + "";
}

const HTTPS = function(location) {
    p = location.protocol
    m = p.match(/^https/);
    return m ? m.length : false
}

const WebSocket = function(tls) {
    return tls ? "wss:" : "ws:"
}

const CONFIG = function() {

    if(window && window.____API_WEB_CONFIG____) {

        const C = {}

        const host = Host()

        let port = ____API_WEB_CONFIG____.apiPort

        DEFINE(C)
        .static("apiPort", port)
        .static("hostUrl", Host(port))
        .static("https", HTTPS(window.location))
        .static("wsProtocol", ()=>{ return WebSocket(C.https) })

        return C;

    } else {
        logger.error("Failed to load Web UI config.")
        return {}
    }

}

export default CONFIG()