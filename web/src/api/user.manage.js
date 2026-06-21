
import {
    AX,
    URLAPIPath
} from "./base"


import log from '@/utils/log';
const logger = log.withPrefix("[api][user.manage]");

const list = function() {
    return AX.get(URLAPIPath("admin", "users"))
    // .then(response=>{

    //     logger.log(response)
    //     if(response && response.data && response.data.Data) {

    //         let d = response.data.Data;
    //         logger.log(d)

    //     }

    // })
}

const mgt = {
    list
}


export default mgt;