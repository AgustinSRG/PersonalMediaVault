<template>
  <div class="main-container">
    <EmptyPlayer
      v-if="test === 0"
      :mid="0"
      :status="'404'"
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
    ></EmptyPlayer>
    <VideoPlayer
      v-if="test === 1"
      :mid="0"
      :metadata="{
        width: 1080,
        height: 720,
        title: 'Media asset title here (very long name)',
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
    <AudioPlayer
      v-if="test === 2"
      :mid="0"
      :metadata="{
        width: 0,
        height: 0,
        title: 'Media asset title here (very long name)',
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
    <ImagePlayer
      v-if="test === 3"
      :mid="0"
      :metadata="{
        width: 1280,
        height: 1308,
        title: 'Media asset title here (very long name)',
        fps: 0,
        url: '/testvid/test.jpg',
        ready: true,
        encoded: true,
        resolutions: [
          {
            width: 480,
            height: 491,
            fps: 30,
            url: '/testvid/test-480.jpg',
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
    ></ImagePlayer>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import VideoPlayer from "./components/player/VideoPlayer.vue";
import AudioPlayer from "./components/player/AudioPlayer.vue";
import ImagePlayer from "./components/player/ImagePlayer.vue";
import EmptyPlayer from "./components/player/EmptyPlayer.vue";

@Options({
  components: {
    VideoPlayer,
    AudioPlayer,
    ImagePlayer,
    EmptyPlayer,
  },
  data: function () {
    return {
      test: 0,
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
    window["TestReload"] = function () {
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
