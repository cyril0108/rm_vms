
import {
    AX,
    URLHostPath,
} from "./base"


const shmMetrics = function() {
    return AX.get(URLHostPath("health", "shm", "metrics"))
}


export default shmMetrics;