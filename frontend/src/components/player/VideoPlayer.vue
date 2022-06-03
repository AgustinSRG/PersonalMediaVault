<template>
  <div
    tabindex="0"
    class="video-player"
    :class="{ 'player-min': minPlayer, 'no-controls': !showControls }"
    @mousemove="playerMouseMove"
    @click="clickPlayer"
    @dblclick="toggleFullScreen"
    @mouseleave="mouseLeavePlayer"
    @mouseup="playerMouseUp"
    @touchmove="playerMouseMove"
    @touchend="playerMouseUp"
    @keydown="onKeyPress"
  >
    <video
      v-if="videoURL"
      :src="videoURL"
      playsinline
      webkit-playsinline
      x-webkit-airplay="allow"
      :muted="muted"
      :volume.prop="volume"
      :playbackRate.prop="speed"
      @ended="onEnded"
      @timeupdate="onVideoTimeUpdate"
      @canplay="onCanPlay"
      @loadedmetadata="onLoadMetaData"
      @waiting="onWaitForBuffer(true)"
      @playing="onWaitForBuffer(false)"
    ></video>

    <div class="player-feeback-container">
      <div
        class="player-feedback player-feedback-play"
        key="play"
        v-if="feedback === 'play'"
        @animationend="onFeedBackAnimationEnd"
      >
        <div><i class="fas fa-play"></i></div>
      </div>
      <div
        class="player-feedback player-feedback-pause"
        key="pause"
        v-if="feedback === 'pause'"
        @animationend="onFeedBackAnimationEnd"
      >
        <div><i class="fas fa-pause"></i></div>
      </div>
    </div>

    <div class="player-loader" v-if="loading">
      <div class="player-lds-ring">
        <div></div>
        <div></div>
        <div></div>
        <div></div>
      </div>
    </div>

    <div
      class="player-controls"
      :class="{ hidden: !showControls }"
      @click="clickControls"
      @dblclick="stopPropagationEvent"
      @mouseenter="enterControls"
      @mouseleave="leaveControls"
    >
      <div class="player-controls-left">
        <button
          v-if="!!next || !!prev"
          :disabled="!prev"
          type="button"
          :title="$t('Previous')"
          class="player-btn"
          @click="goPrev"
        >
          <i class="fas fa-backward-step"></i>
        </button>

        <button
          v-if="!playing"
          type="button"
          :title="$t('Play')"
          class="player-btn player-play-btn"
          @click="togglePlay"
        >
          <i class="fas fa-play"></i>
        </button>
        <button
          v-if="playing"
          type="button"
          :title="$t('Pause')"
          class="player-btn player-play-btn"
          @click="togglePlay"
        >
          <i class="fas fa-pause"></i>
        </button>

        <button
          v-if="!!next || !!prev"
          :disabled="!next"
          type="button"
          :title="$t('Next')"
          class="player-btn"
          @click="goNext"
        >
          <i class="fas fa-forward-step"></i>
        </button>

        <VolumeControl
          ref="volumeControl"
          :min="minPlayer"
          :width="80"
          v-model:muted="muted"
          v-model:volume="volume"
          v-model:expanded="volumeShown"
          @update:volume="onUserVolumeUpdated"
          @update:muted="onUserMutedUpdated"
        ></VolumeControl>

        <div class="player-time-label-container" v-if="!minPlayer">
          <span
            >{{ renderTime(currentTime) }} / {{ renderTime(duration) }}</span
          >
        </div>
      </div>

      <div class="player-controls-right">
        <button
          type="button"
          :title="$t('Player Configuration')"
          class="player-btn"
          @click="showConfig"
        >
          <i class="fas fa-cog"></i>
        </button>

        <button
          v-if="!fullScreen"
          type="button"
          :title="$t('Full screen')"
          class="player-btn player-expand-btn"
          @click="toggleFullScreen"
        >
          <i class="fas fa-expand"></i>
        </button>
        <button
          v-if="fullScreen"
          type="button"
          :title="$t('Exit full screen')"
          class="player-btn player-expand-btn"
          @click="toggleFullScreen"
        >
          <i class="fas fa-compress"></i>
        </button>
      </div>
    </div>

    <div
      class="player-timeline"
      :class="{ hidden: !showControls }"
      @mouseenter="enterControls"
      @mouseleave="mouseLeaveTimeline"
      @mousemove="mouseMoveTimeline"
      @dblclick="stopPropagationEvent"
      @click="stopPropagationEvent"
      @mousedown="grabTimeline"
    >
      <div class="player-timeline-back"></div>
      <div
        class="player-timeline-buffer"
        :style="{ width: getTimelineBarWidth(bufferedTime, duration) }"
      ></div>
      <div
        class="player-timeline-current"
        :style="{ width: getTimelineBarWidth(currentTime, duration) }"
      ></div>
      <div
        class="player-timeline-thumb"
        :style="{ left: getTimelineThumbLeft(currentTime, duration) }"
      ></div>
    </div>

    <VideoPlayerConfig
      v-model:shown="displayConfig"
      v-model:speed="speed"
      v-model:loop="loop"
      v-model:resolution="currentResolution"
      @update:resolution="onResolutionUpdated"
      :rtick="internalTick"
      :metadata="metadata"
      @enter="enterControls"
      @leave="leaveControls"
    ></VideoPlayerConfig>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import VolumeControl from "./VolumeControl.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { isTouchDevice } from "@/utils/touch";
import VideoPlayerConfig from "./VideoPlayerConfig.vue";

export default defineComponent({
  components: {
    VolumeControl,
    VideoPlayerConfig,
  },
  name: "VideoPlayer",
  emits: ["gonext", "goprev"],
  props: {
    mid: Number,
    metadata: Object,
    rtick: Number,

    next: Object,
    prev: Object,
  },
  data: function () {
    return {
      playing: false,
      loading: true,

      videoURL: "",
      videoPending: false,
      videoPendingTask: 0,

      minPlayer: false,
      fullScreen: false,
      displayConfig: false,

      currentTime: 0,
      duration: 0,
      bufferedTime: 0,
      ended: false,
      timelineGrabbed: false,
      lastTimeChangedEvent: 0,

      showControls: true,
      lastControlsInteraction: Date.now(),
      mouseInControls: false,

      loop: false,

      currentResolution: -1,

      volume: 1,
      muted: 0,
      volumeShown: isTouchDevice(),
      internalTick: 0,

      speed: 1,

      feedback: "",
    };
  },
  methods: {
    showConfig: function (e) {
      this.displayConfig = !this.displayConfig;
      e.stopPropagation();
    },

    onResolutionUpdated: function () {
      PlayerPreferences.SetResolutionIndex(
        this.metadata,
        this.currentResolution
      );
      this.setVideoURL();
    },

    clickControls: function (e) {
      this.displayConfig = false;
      e.stopPropagation();
    },

    goNext: function () {
      if (this.next) {
        this.$emit("gonext");
      }
    },

    goPrev: function () {
      if (this.prev) {
        this.$emit("goprev");
      }
    },

    onUserVolumeUpdated() {
      PlayerPreferences.SetVolume(this.volume);
    },

    changeVolume: function (v: number) {
      this.volume = v;
      this.onUserVolumeUpdated();
    },

    onUserMutedUpdated() {
      PlayerPreferences.SetMuted(this.muted);
    },

    toggleMuted: function () {
      this.muted = !this.muted;
      this.onUserMutedUpdated();
    },

    /* Player events */

    onLoadMetaData: function () {
      this.duration = this.getVideoElement().duration;

      this.getVideoElement().currentTime = this.currentTime;
      this.getVideoElement().playbackRate = this.speed;
    },
    onVideoTimeUpdate: function () {
      if (this.loading) return;
      this.currentTime = this.getVideoElement().currentTime;
      this.duration = this.getVideoElement().duration;
      if (Date.now() - this.lastTimeChangedEvent > 5000) {
        PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
        this.lastTimeChangedEvent = Date.now();
      }
    },
    onCanPlay: function () {
      this.loading = false;
      if (!this.playing) {
        return;
      }
      var promise = this.getVideoElement().play();
      if (promise) {
        promise.catch(
          function () {
            this.playing = false;
          }.bind(this)
        );
      }
    },
    onWaitForBuffer: function (b) {
      this.loading = b;
    },
    onEnded: function () {
      this.loading = false;
      if (this.loop) {
        this.getVideoElement().currentTime = 0;
      } else {
        this.pause();
        this.$emit("ended");
        this.ended = true;
      }
    },

    playerMouseUp: function (e) {
      if (this.timelineGrabbed) {
        this.onTimelineSkip(e.pageX, e.pageY);
        this.timelineGrabbed = false;
      }
    },
    playerMouseMove: function (e) {
      this.interactWithControls();

      if (this.timelineGrabbed) {
        this.onTimelineSkip(e.pageX, e.pageY);
      }
    },
    mouseLeavePlayer: function () {
      if (!this.playing) return;
      this.showControls = false;
      this.volumeShown = isTouchDevice();
      this.displayConfig = false;
    },

    checkPlayerSize() {
      var width = this.$el.getBoundingClientRect().width;

      if (width < 480) {
        this.minPlayer = true;
      } else {
        this.minPlayer = false;
      }
    },

    tick() {
      this.checkPlayerSize();

      if (this.showControls && !this.mouseInControls && this.playing) {
        if (Date.now() - this.lastControlsInteraction > 2000) {
          this.showControls = false;
          this.volumeShown = false;
          this.displayConfig = false;
        }
      }

      var video = this.getVideoElement();

      if (video && video.buffered.length > 0) {
        this.bufferedTime = video.buffered.end(video.buffered.length - 1);
      } else {
        this.bufferedTime = 0;
      }
    },

    getVideoElement() {
      return this.$el.querySelector("video");
    },

    interactWithControls() {
      this.showControls = true;
      this.lastControlsInteraction = Date.now();
    },

    enterControls: function () {
      this.mouseInControls = true;
    },

    leaveControls: function () {
      this.mouseInControls = false;
      this.volumeShown = isTouchDevice();
    },

    togglePlay() {
      if (this.playing) {
        this.pause();
        this.feedback = "pause";
      } else {
        this.play();
        this.feedback = "play";
      }

      this.displayConfig = false;
    },

    clickPlayer: function () {
      this.togglePlay();
      this.interactWithControls();
    },

    play: function () {
      var video = this.getVideoElement();
      this.playing = true;
      if (video) {
        video.play();
      }
    },
    pause: function () {
      var video = this.getVideoElement();
      this.playing = false;

      if (video) {
        video.pause();
      }
      this.interactWithControls();
    },

    toggleFullScreen: function () {
      this.fullScreen = !this.fullScreen;
      if (this.fullScreen) {
        openFullscreen();
      } else {
        closeFullscreen();
      }
    },
    onExitFullScreen: function () {
      if (!document.fullscreenElement) {
        this.fullScreen = false;
      }
    },
    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    /* Timeline */

    grabTimeline: function (event) {
      this.timelineGrabbed = true;
      this.onTimelineSkip(event.pageX, event.pageY);
    },
    getTimelineBarWidth: function (time, duration) {
      if (duration > 0) {
        return Math.min((time / duration) * 100, 100) + "%";
      } else {
        return "0";
      }
    },
    getTimelineThumbLeft: function (time, duration) {
      if (duration > 0) {
        return "calc(" + Math.min((time / duration) * 100, 100) + "% - 7px)";
      } else {
        return "-7px";
      }
    },
    onTimelineSkip: function (x: number) {
      var offset = this.$el
        .querySelector(".player-timeline-back")
        .getBoundingClientRect().left;
      var width =
        this.$el.querySelector(".player-timeline-back").getBoundingClientRect()
          .width || 1;
      if (x < offset) {
        this.setTime(0);
      } else {
        var p = x - offset;
        var tP = Math.min(1, p / width);
        this.setTime(tP * this.duration);
      }
    },
    mouseLeaveTimeline: function () {
      this.tooltipShown = false;
      this.leaveControls();
    },
    mouseMoveTimeline: function (event) {
      /*var x = event.pageX;
      var offset = this.$el
        .querySelector(".player-timeline-back")
        .getBoundingClientRect().left;
      var width =
        this.$el.querySelector(".player-timeline-back").getBoundingClientRect()
          .width || 1;

      var time;
      if (x < offset) {
        time = 0;
      } else {
        var p = x - offset;
        var tP = Math.min(1, p / width);
        time = tP * this.duration;
      }*/

      nextTick(this.tick.bind(this));
    },
    renderTime: function (s) {
      if (isNaN(s) || !isFinite(s)) {
        s = 0;
      }
      s = Math.floor(s);
      var hours = 0;
      var minutes = 0;
      if (s >= 3600) {
        hours = Math.floor(s / 3600);
        s = s % 3600;
      }
      if (s > 60) {
        minutes = Math.floor(s / 60);
        s = s % 60;
      }
      var r = "";

      if (s > 9) {
        r = "" + s + r;
      } else {
        r = "0" + s + r;
      }

      if (minutes > 9) {
        r = "" + minutes + ":" + r;
      } else {
        r = "0" + minutes + ":" + r;
      }

      if (hours > 0) {
        if (hours > 9) {
          r = "" + hours + ":" + r;
        } else {
          r = "0" + hours + ":" + r;
        }
      }

      return r;
    },

    setTime: function (time, save) {
      time = Math.max(0, time);
      time = Math.min(time, this.duration);
      this.currentTime = time;

      var video = this.getVideoElement();

      if (video) {
        video.currentTime = time;
      }

      if (save) {
        PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
        this.lastTimeChangedEvent = Date.now();
      }
      if (time < this.duration) {
        this.ended = false;
      }
    },

    onKeyPress: function (event) {
      var catched = true;
      switch (event.key) {
        case "M":
        case "m":
          this.toggleMuted();
          this.showVolume();
          break;
        case "E":
        case "e":
          this.toggleExpand();
          break;
        case " ":
        case "K":
        case "k":
          this.togglePlay();
          break;
        case "ArrowUp":
          this.changeVolume(Math.min(1, this.volume + 0.05));
          this.volumeShown = true;
          break;
        case "ArrowDown":
          this.changeVolume(Math.max(0, this.volume - 0.05));
          this.volumeShown = true;
          break;
        case "F":
        case "f":
          this.toggleFullScreen();
          break;
        case "J":
        case "j":
        case "ArrowRight":
          if (!this.live) {
            this.setTime(this.currentTime + 5, true);
          }
          break;
        case "L":
        case "l":
        case "ArrowLeft":
          if (!this.live) {
            this.setTime(this.currentTime - 5, true);
          }
          break;
        case ".":
          if (!this.playing && !this.live) {
            this.setTime(this.currentTime - 1 / 30);
          }
          break;
        case ",":
          if (!this.playing && !this.live) {
            this.setTime(this.currentTime + 1 / 30);
          }
          break;
        case "Home":
          if (!this.live) {
            this.setTime(0, true);
          }
          break;
        case "End":
          if (!this.live) {
            this.setTime(this.duration, true);
          }
          break;
        case "PageDown":
          this.goPrev();
          break;
        case "PageUp":
          this.goNext();
          break;
        default:
          catched = false;
      }

      if (catched) {
        event.preventDefault();
        event.stopPropagation();
        this.interactWithControls();
      }
    },

    initializeVideo() {
      if (!this.metadata) {
        return;
      }
      this.currentTime = PlayerPreferences.GetInitialTime(this.mid);
      this.duration = 0;
      this.speed = 1;
      this.loop = false;
      this.currentResolution = PlayerPreferences.GetResolutionIndex(
        this.metadata
      );
      this.loading = true;
      this.playing = true;
      this.setVideoURL();
    },

    setVideoURL() {
      if (!this.metadata) {
        this.videoURL = "";
      }

      if (this.currentResolution < 0) {
        if (this.metadata.encoded) {
          this.videoURL = this.metadata.url;
          this.videoPending = false;
          this.videoPendingTask = 0;
        } else {
          this.videoURL = "";
          this.videoPending = true;
          this.videoPendingTask = this.metadata.task;
        }
      } else {
        if (
          this.metadata.resolutions &&
          this.metadata.resolutions.length > this.currentResolution
        ) {
          let res = this.metadata.resolutions[this.currentResolution];
          if (res.ready) {
            this.videoURL = res.url;
            this.videoPending = false;
            this.videoPendingTask = 0;
          } else {
            this.videoURL = "";
            this.videoPending = true;
            this.videoPendingTask = res.task;
          }
        } else {
          this.videoURL = "";
          this.videoPending = true;
          this.videoPendingTask = 0;
        }
      }
    },

    onFeedBackAnimationEnd: function () {
      this.feedback = "";
    },
  },
  mounted: function () {
    // Load player preferences
    this.muted = PlayerPreferences.PlayerMuted;
    this.volume = PlayerPreferences.PlayerVolume;

    this.$options.timer = setInterval(this.tick.bind(this), 100);

    this.$options.exitFullScreenListener = this.onExitFullScreen.bind(this);
    document.addEventListener(
      "fullscreenchange",
      this.$options.exitFullScreenListener
    );
    document.addEventListener(
      "webkitfullscreenchange",
      this.$options.exitFullScreenListener
    );
    document.addEventListener(
      "mozfullscreenchange",
      this.$options.exitFullScreenListener
    );
    document.addEventListener(
      "MSFullscreenChange",
      this.$options.exitFullScreenListener
    );

    this.initializeVideo();
  },
  beforeUnmount: function () {
    clearInterval(this.$options.timer);

    document.removeEventListener(
      "fullscreenchange",
      this.$options.exitFullScreenListener
    );
    document.removeEventListener(
      "webkitfullscreenchange",
      this.$options.exitFullScreenListener
    );
    document.removeEventListener(
      "mozfullscreenchange",
      this.$options.exitFullScreenListener
    );
    document.removeEventListener(
      "MSFullscreenChange",
      this.$options.exitFullScreenListener
    );
  },
  watch: {
    rtick: function () {
      this.internalTick++;
      this.initializeVideo();
    },
    videoURL: function () {
      if (this.videoURL) {
        this.loading = true;
      }
    },
  },
});
</script>

<style>
.video-player {
  background: black;
  color: white;

  display: block;
  position: relative;
  overflow: hidden;
  width: 100%;
  height: 100%;
  -webkit-touch-callout: none;
  /* iOS Safari */
  -webkit-user-select: none;
  /* Safari */
  -khtml-user-select: none;
  /* Konqueror HTML */
  -moz-user-select: none;
  /* Old versions of Firefox */
  -ms-user-select: none;
  /* Internet Explorer/Edge */
  user-select: none;
  /* Non-prefixed version, currently
                                  supported by Chrome, Edge, Opera and Firefox */
}

.video-player video {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.video-player:focus {
  outline: none;
}

.video-player.no-controls {
  cursor: none;
}

.player-controls {
  position: absolute;
  display: block;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 56px;
  background-color: rgba(0, 0, 0, 0.2);
  transition: opacity 0.3s;
  opacity: 1;
}

.player-min .player-controls {
  height: 32px;
}

.player-controls.hidden {
  opacity: 0;
  pointer-events: none;
}

.player-btn {
  display: block;
  width: 40px;
  height: 40px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
  outline: none;
}

.player-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.player-min .player-btn {
  width: 24px;
  height: 24px;
  font-size: 14px;
}

.player-btn:hover {
  color: white;
}

.player-btn:disabled:hover {
  color: rgba(255, 255, 255, 0.75);
}

.player-btn:focus {
  outline: none;
}

.player-controls-left {
  display: flex;
  align-items: center;
  width: 100%;
  height: 100%;
  justify-content: left;
  padding-left: 8px;
  position: absolute;
  top: 0;
  left: 0;
  overflow: hidden;
}

.player-controls-right {
  display: flex;
  align-items: center;
  height: 100%;
  width: auto;
  justify-content: right;
  padding-right: 8px;
  position: absolute;
  top: 0;
  right: 0;
}

.player-min .player-controls-right {
  padding-right: 4px;
}
.player-controls-left .player-controls-left {
  padding-left: 4px;
}

/* Player Loader */

.player-loader {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.3);
}

.player-lds-ring {
  display: inline-block;
  position: relative;
  width: 120px;
  height: 120px;
}

.player-min .player-lds-ring {
  width: 48px;
  height: 48px;
}

.player-lds-ring div {
  box-sizing: border-box;
  display: block;
  position: absolute;
  width: 110px;
  height: 110px;
  margin: 8px;
  border: 8px solid #fff;
  border-radius: 50%;
  animation: player-lds-ring 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
  border-color: #fff transparent transparent transparent;
}

.player-min .player-lds-ring div {
  width: 42px;
  height: 42px;
  margin: 4px;
  border: 4px solid #fff;
  border-color: #fff transparent transparent transparent;
}

.player-lds-ring div:nth-child(1) {
  animation-delay: -0.45s;
}

.player-lds-ring div:nth-child(2) {
  animation-delay: -0.3s;
}

.player-lds-ring div:nth-child(3) {
  animation-delay: -0.15s;
}

@keyframes player-lds-ring {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* Timeline */

.player-timeline {
  position: absolute;
  height: 20px;
  overflow: visible;
  bottom: 56px;
  left: 10px;
  width: calc(100% - 20px);
  cursor: pointer;
  transition: opacity 0.3s;
  opacity: 1;
}

.player-timeline.hidden {
  opacity: 0;
  pointer-events: none;
}

.player-min .player-timeline {
  bottom: 32px;
}

.player-timeline-back {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 5px;
  width: 100%;
  background: rgba(255, 255, 255, 0.25);
}

.player-timeline-buffer {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 5px;
  width: 0;
  background: rgba(255, 255, 255, 0.5);
}

.player-timeline-current {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 5px;
  width: 0;
  background: red;
}

.player-timeline-thumb {
  border-radius: 50%;
  width: 15px;
  height: 15px;
  background: red;
  position: absolute;
  bottom: -5px;
  left: -7px;
}

/* Player feedback */

.player-feeback-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  pointer-events: none;
  overflow: hidden;
}

@keyframes player-feedback-animation {
  0% {
    opacity: 1;
    transform: scale(0.75);
  }

  100% {
    opacity: 0;
    transform: scale(1.5);
  }
}

.player-feedback {
  animation-name: player-feedback-animation;
  animation-fill-mode: forwards;
  animation-duration: 1s;

  width: 80px;
  height: 80px;
  border-radius: 50%;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.5);
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;

  font-size: 24px;
}

.player-min .player-feedback {
  width: 64px;
  height: 64px;
  font-size: 16px;
}
</style>