<script setup>
import CameraStream from '../components/CameraStream.vue'
import Player from '../components/Player.vue'
let page = Player
</script>

<template>
  <main>
    <component :is="page"></component>
  </main>
</template>
