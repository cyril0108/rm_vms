
import config from '@/config'
import {
    UserNeedLogin,
    UpdateAX,
    AX,
} from './base.ax'
import axios from 'axios'

console.log("[API] config", config)
console.log("[API] https", config.https)

const HostBase = config.hostUrl;
const APIBase = "api";

const URLJoin = function(...comps) {
    return comps.join("/");
}

const URLHostPath = function(...comps) {
    return URLJoin.bind(null, [HostBase]).apply(null, comps)
}

const URLAPIPath = function(...comps) {
    return URLHostPath.bind(null, [APIBase]).apply(null, comps)
}

export {
    UserNeedLogin,
    UpdateAX,
    AX,
    axios,
    HostBase,
    APIBase,
    URLJoin,
    URLHostPath,
    URLAPIPath,
}