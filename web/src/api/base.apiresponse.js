import Logger from '@/utils/log';

const logger = Logger.withPrefix("[APIResponse]");
logger.log("init")

import DEFINE from "@/utils/define"

const APIResponse = function(response) {

    if( response instanceof APIResponse ) {
        return response
    }

    if (!(this instanceof APIResponse)) {
        return new APIResponse(response);
    }

    const payload = response.data;
    const statusCode = response.status;

    if( typeof payload != "object") {
        logger.log("receiving abnormal payload.", Object.assign(payload))
        payload = {}
    }

    let message = payload.message;

    DEFINE(this)
    .static("raw", response)
    .static("data", payload.data)
    .static("message", message)
    .static("statusCode", statusCode)
    .property("forbidden", {
        get: function() {
            return statusCode == 401
        }
    })
    .property("success", {
        get: function() {
            switch(statusCode) {
            case 200:
            case 201:
            case 202:
            case 203:
            case 204:
            case 205:
            case 206:
            case 207:
                return true
            default:
            }
            return false
        }
    })

    return this

}

export default APIResponse