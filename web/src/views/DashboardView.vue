<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
    <div class="row my-2">
      <div class="col-1">
        <button class="btn btn-primary me-1" @click="toogleRefresh">{{ToogleRefreshTitle}}
        </button>
        <Loader v-if="metricsLoader.loading"/>
      </div>
      <span class="col">{{lastUpdate}}</span>
    </div>
    <CameraList class="row" :cameras="cameras"></CameraList>
    <SHMMetrics class="row" :metrics="metrics"></SHMMetrics>
  </div>
</template>


<script setup>
import Logger from '@/utils/log';

const logger = Logger.withPrefix("[Dashboard]");
logger.log("init")

import { ref, computed } from 'vue';

import SHMMetrics from '@/components/SHMMetrics.vue';
import CameraList from '@/components/CameraList.vue';
import Loader from '@/components/Loader.vue';

import BusyLoader from '@/models/busy.loader'

import API from '@/api';

const metrics = ref([]);
const cameras = ref([]);
const metricsLoader = BusyLoader();

const UpdateCameras = function() {

  let ll = logger.lin("[UpdateCameras]")

  API.cameras.list()
  .then(response=>{

    let data = response.data;
    if (typeof data === "object") {

      // metrics.value = data;
      // lastUpdate.value = new Date;
      ll.log("data", data);
      cameras.value = data.Data;

    } else {

      ll.log("response", response);
      ll.log("data type", typeof response.data)

    }

  })


}

const UpdateData = function() {

  UpdateCameras()

  logger.log("updating...")
  metricsLoader.busy()

  API.shmMetrics()
  .then(response=>{

    let data = response.data;
    if (typeof data === "object" && typeof data.Data === "object") {

      metrics.value = data.Data;
      lastUpdate.value = new Date;

    } else {

      logger.log("response", response);
      logger.log("data type", typeof response.data)

    }

  })
  .finally(metricsLoader.idle)

}

let RefreshTimer = ref();

const ToogleRefreshTitle = computed(()=>{
  return RefreshTimer.value ? "Stop" : "Start"
})

const lastUpdate = ref();
const LastUpdateTime = computed(()=>{
  return lastUpdate.value
})

const toogleRefresh = function() {

  if(RefreshTimer.value) {
    clearInterval(RefreshTimer.value);
    RefreshTimer.value = undefined
  } else {
    RefreshTimer.value = setInterval(UpdateData, 5000);
  }

}

UpdateData()
toogleRefresh()

</script>

<style>
@media (min-width: 1024px) {
  .dashboard {
    min-height: 100vh;
    /*display: flex;*/
    align-items: center;
  }
}
</style>
