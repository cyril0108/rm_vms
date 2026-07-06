
import {
    AX,
    URLAPIPath,
} from "./base"


const shmMetrics = function() {
    return AX.get(URLAPIPath("shmmetrics"))
}


export default shmMetrics;