<template>
  <div class="card h-100 bg-dark border-secondary camera-card">
    <div class="card-body d-flex flex-column gap-2">

      <div class="d-flex justify-content-between align-items-start">
        <h5 class="card-title text-primary mb-0 d-flex align-items-center gap-2">
          <IconCamera />
          {{ camera.name || `Camera ${camera.id}` }}
        </h5>
        <span :class="['badge', camera.active ? 'bg-success' : 'bg-danger']">
        </span>
      </div>

      <CameraProfile v-if="camera.main_stream" :camerasp="camera.main_stream"></CameraProfile>
      <CameraProfile v-if="camera.sub_stream" :camerasp="camera.sub_stream"></CameraProfile>

      <div class="mt-auto pt-3">
        <button class="btn btn-outline-primary btn-sm w-100">
          View Live Stream
        </button>
      </div>

    </div>

  </div>
</template>

<script setup>
import { defineProps } from 'vue'
import IconCamera from './icons/IconCamera.vue'
import CameraProfile from './CameraProfile.vue'

// Pure presentational component
const props = defineProps({
  camera: {
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