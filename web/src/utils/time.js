import { format, getUnixTime, fromUnixTime, startOfDay, endOfDay, addDays, subDays } from 'date-fns';
import DEFINE from './define';

const DateFormat = 'yyyy-MM-dd';
const TimelineFormat = "yyyy-MM-dd'T'HH:mm:ss";

/**
 * ========================================================
 * Day Ranging
 * ========================================================
 */
const APIDayRange = function(date) {

    // startOfDay and endOfDay safely handle exact boundary calculations
    let start = getUnixTime(startOfDay(date));
    let end = getUnixTime(endOfDay(date));

    return { start, end }
}

const DayRange = function(date) {

    // startOfDay and endOfDay safely handle exact boundary calculations
    let start = startOfDay(date);
    let end = endOfDay(date);

    return { start, end }
}


const APITodayRange = function() {
    return APIDayRange(new Date)
}

const WebTimelineBoundaries = function(date) {

    let range = DayRange(date);

    let o = {
        start: ToTimelineStr(range.start),
        end: ToTimelineStr(range.end),
        min: format(subDays(date, 1), "yyyy-MM-dd'T'23:00:00"),
        max: format(addDays(date, 1), "yyyy-MM-dd'T'01:00:00"),
    };

    return o;
}


/**
 * ========================================================
 * Date Time Format Converting
 * ========================================================
 */

const ToDateStr = function(date) {
    return format(date, DateFormat);
}

// getUnixTime automatically converts milliseconds to seconds
const DatetimeToAPIStamps = function(date) {
    return getUnixTime(date);
}

// fromUnixTime automatically converts seconds back to a Date object
const APIStampsToDatetime = function(stamps) {
    return fromUnixTime(stamps);
}

// Safely formats to '2026-04-15T08:30:00', handling all zero-padding
const ToTimelineStr = function(date) {
    return format(date, TimelineFormat);
}


/**
 * Object-Oriented Wrapper
 * Allows chaining or contextual operations around a specific Date.
 * Added a default parameter so Time() safely defaults to now.
 */
const WebTime = function(date = new Date()) {

    let wt = {};

    DEFINE(wt)
    .static("Native", date)
    .static("APIStamp", () => getUnixTime(date))
    .static("Timeline", () => format(date, TimelineFormat))

    return wt;

}

const APITime = function(stamps) {

    const date = fromUnixTime(stamps);

    let obj = {};

    DEFINE(obj)
    .static("Native", date)
    .static("WebTime", ()=>new WebTime(date))

    return obj;

}

const FormatDuration = (totalSeconds) => {
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    return hours > 0 ? `${hours}h ${minutes}m recorded` : `${minutes}m recorded`;
};


export {
    ToDateStr,
    WebTime,
    APITime,
    DayRange,
    APIDayRange,
    APITodayRange,
    WebTimelineBoundaries,
    FormatDuration,
}