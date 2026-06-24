
import {
    AX,
    URLAPIPath,
    URLHostPath,
} from "./base"

const QueryParamsFrom = (data) => {
    return new URLSearchParams(data).toString();
}

const timeline = function(camID, from, to) {
    return AX.get(URLAPIPath("cameras", camID, "timeline", from, to))
}

// /api/cameras/{cam_id}/summary?start=1000&end=2000
const dailySummary = function(camID, start, end) {
    return AX.get(URLAPIPath("cameras", camID, "summary")+"?"+QueryParamsFrom({start,end}))
}


export default {
    timeline,
    dailySummary,
};