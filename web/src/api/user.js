
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
    }).then(apires=>{

        logger.log(apires)
        if( apires.success ) {

            let d = apires.data;
            let token = d.token

            logger.log(token)

            if( token && token.length > 0 ) {
                UpdateAX.login(token)
            }

        }

    })
}

const logout = function() {
    return AX.post(URLAPIPath("web", "refresh"), {
        logout: true,
    }).then(apires=>{

        logger.log(apires)
        if( apires.success ) {
            UpdateAX.logout()
        }

    })
}

const refresh = function(username , password) {
    return AX.post(URLAPIPath("web", "refresh"), {}).then(apires=>{

        logger.log(apires)
        if(apires.success) {

            let d = apires.data;
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