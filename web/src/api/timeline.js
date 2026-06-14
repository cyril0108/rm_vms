
import {
    AX,
    URLAPIPath,
    URLHostPath,
} from "./base"


const timeline = function(camID, from, to) {
    return AX.get(URLAPIPath("cameras", camID, "timeline", from, to))
}


export default timeline;