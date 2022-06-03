<template>
  <div class="main-container">
    <VideoPlayer v-if="test === 1"
      :mid="0"
      :metadata="{
        width: 1080,
        height: 720,
        fps: 30,
        url: '/testvid/video.mp4',
        video_previews: '/testvid/thumb_{INDEX}.jpg',
        video_previews_interval: 3,
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
          {
            width: 240,
            height: 144,
            fps: 30,
            url: '',
            ready: false,
            task: 1,
          },
        ],
      }"
      :rtick="rtick"
      :next="{
        id: 2,
        type: 2,
        title: '',
        thumbnail: '/testvid/thumb_1.jpg',
        duration: 120,
        width: 800,
        height: 600,
        fps: 30,
      }"
      :prev="{
        id: 2,
        type: 3,
        title: '',
        thumbnail: '',
        duration: 120,
        width: 800,
        height: 600,
        fps: 30,
      }"
    ></VideoPlayer>
    <AudioPlayer v-if="test === 2"
      :mid="0"
      :metadata="{
        width: 0,
        height: 0,
        fps: 0,
        url: '/testvid/audio.mp3',
        ready: true,
        encoded: true,
        resolutions: [],
      }"
      :rtick="rtick"
      :next="{
        id: 2,
        type: 2,
        title: '',
        thumbnail: '/testvid/thumb_1.jpg',
        duration: 120,
        width: 800,
        height: 600,
        fps: 30,
      }"
      :prev="{
        id: 2,
        type: 3,
        title: '',
        thumbnail: '',
        duration: 120,
        width: 800,
        height: 600,
        fps: 30,
      }"
    ></AudioPlayer>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import VideoPlayer from "./components/player/VideoPlayer.vue";
import AudioPlayer from "./components/player/AudioPlayer.vue";

@Options({
  components: {
    VideoPlayer,
    AudioPlayer,
  },
  data: function () {
    return {
      test: 2,
      rtick: 0,
      volume: 0.5,
      muted: false,
    };
  },
  methods: {
    runTest: function () {
      this.test++;
    },
  },
  mounted: function () {
    window["TestReload"] = function() {
      this.rtick++;
    }.bind(this);
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
</style>
