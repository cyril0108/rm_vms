<template>
  <div class="video-player-view">
    <h2>NVR Playback</h2>
    <input type="number" :value="cameraid" @change="onCameraIdChange">
    <!-- <input type="date" :value="selectDayStr" @change="onSelectDayChange"> -->

    <VDatePicker
      v-model="selectedDate"
      :attributes="calendarAttributes"
      @did-move="handleDidMove"
      color="blue"
    >
      <template #default="{ togglePopover, inputValue }">
            <button 
              class="nvr-date-btn" 
              @click="togglePopover"
            >
              📅 {{ inputValue || 'Select Date' }}
            </button>
      </template>
    </VDatePicker>

    <video src="" class="player-video"></video>
    <Timeline 
      ref="timelineRef"
      :items="timelineItems" 
      :options="timelineOptions"
      :initialTime="currentPlaybackTime"
      @timechange="handleScrubbing"
      @seek="handleUserSeek"
    />
  </div>
</template>

<script setup>

import Logger from '@/utils/log';

const log = Logger.withPrefix("[Player]");

import { ref, computed, watch, onMounted } from 'vue';
import Timeline from './Timeline.vue';

import TimelineDefaults from '@/data/timeline'

import API from '@/api';
import {
  WebTime,
  APITime,
  APIDayRange,
  WebTimelineBoundaries,
  ToDateStr,
  FormatDuration,
} from '@/utils/time'

// Reference to the child component
const timelineRef = ref(null);
const currentPlaybackTime = ref(new Date());

let selectedDate = ref(new Date());
const monthlySummaries = ref([]);

// let selectDay = ref(new Date);
// const selectDayStr = computed(() => {
//   return ToDateStr(selectDay.value);
// });

let cameraid = ref(1);

/// =========================================
/// == On Mount
/// =========================================

onMounted(() => {

  const initDate = selectedDate.value;

  log.log("[onMounted]", initDate)

  fetchMonthlyData(initDate.getFullYear(), initDate.getMonth());

});


/// =========================================
/// == Date Selection
/// =========================================
const handleDidMove = async (pages)=> {
  const page = pages[0]; 

  log.log("[handleDidMove]", page.year, page.month)

  fetchMonthlyData(page.year, page.month - 1);
}

const fetchMonthlyData = async (year, monthIdx) => {

  try {

    log.log("[fetchMonthlyData]", year, monthIdx)

    // first and last days of the target month
    const firstDayOfMonth = new Date(year, monthIdx, 1);
    const lastDayOfMonth = new Date(year, monthIdx + 1, 0);

    const from = firstDayOfMonth.getTime() / 1000;
    const to = lastDayOfMonth.getTime() / 1000;

    // Calling the Go API we defined in Phase 1
    const apires = await API.playback.dailySummary(cameraid.value, from, to);

log.log("daily summary", apires.data)

    // Using your unified APIResponse format
    if (apires.success) {
      monthlySummaries.value = apires.data; // [{ date: '2026-06-21', total_seconds: 3600 }, ...]
    }

  } catch (err) {
    console.error("Failed to fetch calendar data", err);
  }

};

const calendarAttributes = computed(() => {
  return monthlySummaries.value.map(dayInfo => {
    return {
      key: dayInfo.date,       // Unique key for VCalendar tracking
      dates: new Date(dayInfo.date), // The specific day to place the attribute on

      // Shows a little dot under the date number
      dot: {
        color: 'blue', 
      },

      // Native VCalendar popover on hover
      popover: {
        label: FormatDuration(dayInfo.total_seconds),
        visibility: 'hover',
        hideIndicator: true,
      }
    };
  });
});


/// =========================================
/// == Timeline
/// =========================================

const fetchTimeline = function(day) {

  let ll = log.lin("[fetchTimeline]");
  ll.log(day);

  day = APIDayRange(day);

  API.playback.timeline(cameraid.value, day.start, day.end)
  .then(apires=>{

    if( apires.success ) {

      let data = apires.data

      ll.log("data", data);

      let list = data.timelines ?? []

      updateTimelineItems(list);

    }

  })

}

fetchTimeline(selectedDate.value);

const Timeline2Items = function(apitl) {

  let start = APITime(apitl.start_time).WebTime().Timeline();
  let end = APITime(apitl.end_time).WebTime().Timeline();

  return {
    id: apitl.start_time,
    content: '',
    start: start,
    end: end,
    type: 'background',
    className: 'continuous-record'
  }
}

const updateTimelineItems = function(list) {

  let ll = log.lin("[updateTimelineItems]");
  ll.log("new list length", list.length)

  while(timelineItems.value.length) {
    timelineItems.value.shift(); 
  }

  list.map((v)=>{
    timelineItems.value.push(Timeline2Items(v))
  })

}


// Define motion events or continuous recording blocks
const timelineItems = ref([]);

// Configure the view for a 24-hour period
const timelineOptionsByDate = function(date) {
  let bounaries = WebTimelineBoundaries(date);
  return {
    ...TimelineDefaults,
    ...bounaries,
  }
}

const updateTimelineBounds = function(date) {
  let bounaries = WebTimelineBoundaries(date);
  timelineOptions.value = {
    ...timelineOptions.value,
    ...bounaries,
  };
};

const timelineOptions = ref(timelineOptionsByDate(selectedDate.value));

// Handle the user dragging the playhead
const handleScrubbing = (properties) => {

  let ll = log.lin("[handleScrubbing]");

  const scrubbedTime = properties.time;
  // ll.log("User is scrubbing to:", scrubbedTime);

  // Send this timestamp over WebSocket to your Go backend to fetch the new video chunk


};

const handleUserSeek = (date) => {

  let ll = log.lin("[handleUserSeek]");

  currentPlaybackTime.value = date;

  // Send Unix timestamp to Go Backend
  const timestampMs = date.getTime();
  ll.log(`Commanding NVR to seek to: ${timestampMs}`);

  // Example WebSocket Payload:
  // ws.send(JSON.stringify({ command: 'SEEK', timestamp: timestampMs }));

};

watch(selectedDate, (newDate) => {
  let ll = log.lin("[selectedDate changed]");
  
  // newDate is already a valid Date object provided by VCalendar
  updateTimelineBounds(newDate);
  fetchTimeline(newDate);
});

const onCameraIdChange = (e)=> {

  let id = parseInt(e.target.value)
  if(!isNaN(id)) {
    cameraid.value = id
    fetchTimeline(selectedDate.value);
  }

}

const updatePlayheadFromVideoSync = (videoTimestampMs) => {
  const newTime = new Date(videoTimestampMs);
  currentPlaybackTime.value = newTime;
  if (timelineRef.value) {
    timelineRef.value.setPlayheadTime(newTime); // Just updates UI visually
  }
}

</script>

<style>
  video.player-video {
    width: 100%;
    min-height: 300px;
    min-width: 300px;
    border: 1px solid;
  }
</style>