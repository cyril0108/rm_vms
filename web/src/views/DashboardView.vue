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
  .then(apires=>{

    let data = apires.data;
    if (apires.success) {

      // ll.log("data", data);
      let list = data;
      if( list && list.length > 0) {
        list.sort((a,b)=>{ return a.id - b.id })
      }

      cameras.value = list;

    } else {

      ll.error("response", apires);

    }

  })


}

const UpdateData = function() {

  UpdateCameras()

  logger.log("updating...")
  metricsLoader.busy()

  API.shmMetrics()
  .then(apires=>{

    if(apires.success) {

      metrics.value = apires.data;
      lastUpdate.value = new Date;

    } else {

      logger.error("response", apires);

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
