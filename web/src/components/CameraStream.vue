<template>
  <div class="camera-container">
    <video 
      :id="'player-' + camId" 
      ref="videoPlayer" 
      autoplay 
      muted 
      playsinline
      class="video-player"
    ></video>
    
    <div v-if="isConnecting" class="overlay">
      Connecting to Camera {{ camId }}...
    </div>
    <button v-if="isConnecting&&ws" @click="reconnect">Reconnect</button>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import JMuxer from 'jmuxer';

import config from "@/config"

const wshosturl = "wss:" + config.hostUrl;

console.log("ws host:", wshosturl);

const props = defineProps({
  camId: {
    type: String,
    required: true
  },
  // wsHost: {
  //   type: String,
  //   default: wshosturl
  // }
});

const videoPlayer = ref(null);
const isConnecting = ref(true);
let jmuxer = null;
let ws = null;

onMounted(() => {
  initPlayer();
});

onBeforeUnmount(() => {
  cleanup();
});

const reconnect = () => {
  cleanup()
  initPlayer()
}

const makeMuxer = function(mode) {

  console.log(`[makeMuxer] start with mode: ${mode}`);

  return new JMuxer({
    node: 'player-' + props.camId,
    mode: mode,
    flushingTime: 0,        // 0 = ultra-low latency. Flushes frames instantly.
    clearBuffer: true,      // Keeps memory usage low over long periods
    fps: 30,                // Fallback FPS, though the H.264 stream usually dictates this
    debug: false,
    onError: (data) => {
      console.error(`[JMuxer Cam ${props.camId}] Error:`, data);
    }
  });
}

const initPlayer = () => {
  //  Initialize JMuxer
  // jmuxer = makeMuxer("both");
  let jmuxer;
  let cnt = 0;
  let mode = "video"; // Default to video only

  //  Initialize the WebSocket
  const wsUrl = `${wshosturl}/ws/stream/${props.camId}`;
  ws = new WebSocket(wsUrl);

  // CRITICAL: Tell the browser we expect raw binary data, not text
  ws.binaryType = 'arraybuffer';

  ws.onopen = () => {
    console.log(`[WS] Connected to Camera ${props.camId}`);
    isConnecting.value = false;
  };

  ws.onmessage = (event) => {

    const headerView = new Uint8Array(event.data, 0, 1);
    const mediaType = headerView[0];

    // Wrap the incoming ArrayBuffer
    const payloadBuffer = event.data.slice(1);
    const payload = new Uint8Array(payloadBuffer);

    // Extract the payload (everything after index 0)
    // const payload = data.subarray(1);

    if(!jmuxer) {

      cnt++;

      if( mediaType===1 ) {

        mode = 'both'
        jmuxer = makeMuxer(mode);

      } else {

        if(cnt > 9) {
          jmuxer = makeMuxer(mode);
        }

      }


    } else {

      // Route to the correct JMuxer buffer
      if (mediaType === 0) {
        // Video Packet (H.264 Annex-B)
        jmuxer.feed({
          video: payload
        });
      } else if (mediaType === 1) {
        // Audio Packet (Typically AAC)
        jmuxer.feed({
          audio: payload
        });
      } else {
        console.warn("Unknown media type received:", mediaType);
      }

    }

  };

  ws.onclose = () => {
    console.log(`[WS] Disconnected from Camera ${props.camId}`);
    isConnecting.value = true;
    // Optional: Add reconnection logic here
  };
};

const cleanup = () => {
  if (ws) {
    ws.close();
    ws = null;
  }
  if (jmuxer) {
    jmuxer.destroy();
    jmuxer = null;
  }
};
</script>

<style>
.camera-container {
  position: relative;
  width: 100%;
  background: #000;
  aspect-ratio: 16 / 9;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-family: monospace;
}
</style>