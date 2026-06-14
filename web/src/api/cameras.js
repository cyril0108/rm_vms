
import {
    AX,
    URLAPIPath,
    URLHostPath,
} from "./base"


const cameraList = function() {
    return AX.get(URLAPIPath("cameras"))
}

const CAM = {
    list: cameraList
}


export default CAM;