<template>
  <div class="text-muted small mt-2">
    <div class="d-flex justify-content-between">
      <span><strong>Worker:</strong></span>
      <span>{{ camerasp.worker_id }}</span>
    </div>
    <div class="d-flex justify-content-between">
      <span><strong>SHM Channel:</strong></span>
      <span>{{ camerasp.channel_id }}</span>
    </div>
    <div class="d-flex justify-content-between">
      <span><strong>Status:</strong></span>
      <span>{{ camerasp.status }}</span>
    </div>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'
// import IconCamera from './icons/IconCamera.vue'

// Pure presentational component
const props = defineProps({
  camerasp: {
    type: Object,
    required: true,
    default: () => {}
  }
})
</script>

<style scoped>
.camera-card {
  /* Inherits your custom $body-color from SCSS, but ensures text is readable */
  color: #cdd6f4; 
  transition: transform 0.2s ease-in-out, border-color 0.2s ease;
}

/* Subtle lift effect on hover to make the grid feel interactive */
.camera-card:hover {
  transform: translateY(-3px);
  border-color: #89b4fa !important; /* Matches your custom $primary color */
}
</style>