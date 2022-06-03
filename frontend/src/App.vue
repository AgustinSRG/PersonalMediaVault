<template>
  <div class="main-container">
    <VideoPlayer
      :mid="0"
      :metadata="{
        width: 1080,
        height: 720,
        fps: 30,
        url: '/testvid/video.mp4',
        ready: true,
        encoded: true,
        resolutions: [
          {
            width: 480,
            height: 360,
            fps: 30,
            url: '/testvid/video-480.mp4',
            ready: true,
          },
        ],
      }"
      :rtick="0"
      :next="{}"
      :prev="null"
    ></VideoPlayer>
  </div>

  <HelloWorld msg="Welcome to Your Vue.js + TypeScript App" />
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import VolumeControl from "./components/player/VolumeControl.vue";
import VideoPlayer from "./components/player/VideoPlayer.vue";

@Options({
  components: {
    VideoPlayer,
    VolumeControl,
  },
  data: function () {
    return {
      test: 1,
      volume: 0.5,
      muted: false,
    };
  },
  methods: {
    runTest: function () {
      this.test++;
    },
  },
})
export default class App extends Vue {}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}

.main-container {
  display: flex;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

/* Custom scroll bar */


/* width */

::-webkit-scrollbar {
    width: 5px;
    height: 3px;
}


/* Track */

::-webkit-scrollbar-track {
    background: #bdbdbd;
}


/* Handle */

::-webkit-scrollbar-thumb {
    background: #757575;
}
</style>
