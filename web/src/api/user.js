
import {
    UpdateAX,
    AX,
    URLAPIPath
} from "./base"

import manage from "./user.manage"

import log from '@/utils/log';
const logger = log.withPrefix("[api][login]");

const login = function(username , password) {
    return AX.post(URLAPIPath("web", "login"), {
        username,
        password,
    }).then(response=>{

        logger.log(response)
        if(response && response.data && response.data.data) {

            let d = response.data.data;
            let token = d.token
            logger.log(token)

            if( token && token.length > 0 ) {
                UpdateAX.login(token)
            }

        }

    })
}


const refresh = function(username , password) {
    return AX.post(URLAPIPath("web", "refresh"), {}).then(response=>{

        logger.log(response)
        if(response && response.data && response.data.data) {

            let d = response.data.data;
            let token = d.token
            logger.log(token)

            if( token && token.length > 0 ) {
                UpdateAX.login(token)
            }

        }

    })
}

const user = {
    login,
    refresh,
    manage,
}


export default user;