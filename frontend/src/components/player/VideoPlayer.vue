<template>
  <div
    tabindex="0"
    class="video-player"
    :class="{
      'player-min': minPlayer,
      'no-controls': !showControls,
      'full-screen': fullscreen,
    }"
    @mousemove="playerMouseMove"
    @click="clickPlayer"
    @mousedown="hideContext"
    @touchstart="hideContext"
    @dblclick="toggleFullScreen"
    @mouseleave="mouseLeavePlayer"
    @mouseup="playerMouseUp"
    @touchmove="playerMouseMove"
    @touchend="playerMouseUp"
    @keydown="onKeyPress"
    @contextmenu="onContextMenu"
  >
    <video
      v-if="videoURL"
      :src="videoURL"
      :key="rtick"
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

    <PlayerEncodingPending v-if="!loading && !videoURL && videoPending" :mid="mid" :tid="videoPendingTask" :res="currentResolution"></PlayerEncodingPending>

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
          @mouseenter="enterTooltip('prev')"
          @mouseleave="leaveTooltip('prev')"
        >
          <i class="fas fa-backward-step"></i>
        </button>

        <button
          v-if="!playing"
          type="button"
          :title="$t('Play')"
          class="player-btn player-play-btn"
          @click="togglePlay"
          @mouseenter="enterTooltip('play')"
          @mouseleave="leaveTooltip('play')"
        >
          <i class="fas fa-play"></i>
        </button>
        <button
          v-if="playing"
          type="button"
          :title="$t('Pause')"
          class="player-btn player-play-btn"
          @click="togglePlay"
          @mouseenter="enterTooltip('pause')"
          @mouseleave="leaveTooltip('pause')"
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
          @mouseenter="enterTooltip('next')"
          @mouseleave="leaveTooltip('next')"
        >
          <i class="fas fa-forward-step"></i>
        </button>

        <VolumeControl
          ref="volumeControl"
          :min="minPlayer"
          :width="minPlayer ? 50 : 80"
          v-model:muted="muted"
          v-model:volume="volume"
          v-model:expanded="volumeShown"
          @update:volume="onUserVolumeUpdated"
          @update:muted="onUserMutedUpdated"
          @enter="enterTooltip('volume')"
          @leave="leaveTooltip('volume')"
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
          :title="$t('Manage albums')"
          class="player-btn"
          @click="manageAlbums"
          @mouseenter="enterTooltip('albums')"
          @mouseleave="leaveTooltip('albums')"
        >
          <i class="fas fa-list-ol"></i>
        </button>

        <button
          type="button"
          :title="$t('Player Configuration')"
          class="player-btn"
          @click="showConfig"
          @mouseenter="enterTooltip('config')"
          @mouseleave="leaveTooltip('config')"
        >
          <i class="fas fa-cog"></i>
        </button>

        <button
          v-if="!fullscreen"
          type="button"
          :title="$t('Full screen')"
          class="player-btn player-expand-btn"
          @click="toggleFullScreen"
          @mouseenter="enterTooltip('full-screen')"
          @mouseleave="leaveTooltip('full-screen')"
        >
          <i class="fas fa-expand"></i>
        </button>
        <button
          v-if="fullscreen"
          type="button"
          :title="$t('Exit full screen')"
          class="player-btn player-expand-btn"
          @click="toggleFullScreen"
          @mouseenter="enterTooltip('full-screen-exit')"
          @mouseleave="leaveTooltip('full-screen-exit')"
        >
          <i class="fas fa-compress"></i>
        </button>
      </div>
    </div>

    <div
      v-if="helpTooltip === 'play'"
      class="player-tooltip player-helptip-left"
    >
      {{ $t("Play") }}
    </div>
    <div
      v-if="helpTooltip === 'pause'"
      class="player-tooltip player-helptip-left"
    >
      {{ $t("Pause") }}
    </div>

    <div
      v-if="prev && helpTooltip === 'prev'"
      class="player-tooltip player-helptip-left"
    >
      <PlayerMediaChangePreview
        :media="prev"
        :next="false"
      ></PlayerMediaChangePreview>
    </div>

    <div
      v-if="next && helpTooltip === 'next'"
      class="player-tooltip player-helptip-left"
    >
      <PlayerMediaChangePreview
        :media="next"
        :next="true"
      ></PlayerMediaChangePreview>
    </div>

    <div
      v-if="helpTooltip === 'volume'"
      class="player-tooltip player-helptip-left"
    >
      {{ $t("Volume") }} ({{ muted ? $t("Muted") : renderVolume(volume) }})
    </div>

    <div
      v-if="!displayConfig && helpTooltip === 'config'"
      class="player-tooltip player-helptip-right"
    >
      {{ $t("Player Configuration") }}
    </div>

    <div
      v-if="!displayConfig && helpTooltip === 'albums'"
      class="player-tooltip player-helptip-right"
    >
      {{ $t("Manage albums") }}
    </div>

    <div
      v-if="helpTooltip === 'full-screen'"
      class="player-tooltip player-helptip-right"
    >
      {{ $t("Full screen") }}
    </div>
    <div
      v-if="helpTooltip === 'full-screen-exit'"
      class="player-tooltip player-helptip-right"
    >
      {{ $t("Exit full screen") }}
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
      @touchstart="grabTimeline"
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

    <div
      v-if="tooltipShown"
      class="player-tooltip"
      :style="{ left: tooltipX + 'px' }"
    >
      <div v-if="tooltipImage && !tooltipImageInvalid">
        <img
          class="player-tooltip-image"
          :src="tooltipImage"
          @error="onTooltipImageError"
        />
      </div>
      <div class="player-tooltip-text">{{ tooltipText }}</div>
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

    <PlayerTopBar
      v-if="metadata"
      :mid="mid"
      :metadata="metadata"
      :shown="showControls"
      :fullscreen="fullscreen"
      v-model:expanded="expandedTitle"
      @update:expanded="pause"
      v-model:albumexpanded="expandedAlbum"
      :inalbum="inalbum"
    ></PlayerTopBar>

    <PlayerContextMenu
      type="video"
      v-model:shown="contextMenuShown"
      :x="contextMenuX"
      :y="contextMenuY"
      v-model:loop="loop"
      :url="videoURL"
    ></PlayerContextMenu>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import VolumeControl from "./VolumeControl.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerEncodingPending from "./PlayerEncodingPending.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { renderTimeSeconds } from "../../utils/time-utils";
import { isTouchDevice } from "@/utils/touch";
import VideoPlayerConfig from "./VideoPlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import { GetAssetURL } from "@/utils/request";
import { useVModel } from "../../utils/vmodel";
import { MediaController } from "@/control/media";

export default defineComponent({
  components: {
    VolumeControl,
    VideoPlayerConfig,
    PlayerMediaChangePreview,
    PlayerTopBar,
    PlayerContextMenu,
    PlayerEncodingPending,
  },
  name: "VideoPlayer",
  emits: ["gonext", "goprev", "ended", "update:fullscreen", "albums-open"],
  props: {
    mid: Number,
    metadata: Object,
    rtick: Number,

    fullscreen: Boolean,

    next: Object,
    prev: Object,
    inalbum: Boolean,
  },
  setup(props) {
    return {
      fullScreen: useVModel(props, "fullscreen"),
    };
  },
  data: function () {
    return {
      playing: false,
      loading: true,

      videoURL: "",
      videoPending: false,
      videoPendingTask: 0,

      minPlayer: false,
      displayConfig: false,

      currentTime: 0,
      duration: 0,
      bufferedTime: 0,
      ended: false,
      timelineGrabbed: false,
      lastTimeChangedEvent: 0,

      // Timeline tooltip
      tooltipShown: false,
      tooltipText: "",
      tooltipX: 0,
      tooltipEventX: 0,
      tooltipImage: "",
      tooltipImageInvalid: false,

      showControls: true,
      lastControlsInteraction: Date.now(),
      mouseInControls: false,

      loop: false,

      currentResolution: -1,

      volume: 1,
      muted: false,
      volumeShown: isTouchDevice(),
      internalTick: 0,

      speed: 1,

      feedback: "",

      helpTooltip: "",

      expandedTitle: false,
      expandedAlbum: false,

      contextMenuX: 0,
      contextMenuY: 0,
      contextMenuShown: false,

      requiresRefresh: false,
    };
  },
  methods: {
    onContextMenu: function (e) {
      this.contextMenuX = e.pageX;
      this.contextMenuY = e.pageY;
      this.contextMenuShown = true;
      e.preventDefault();
    },

    manageAlbums: function () {
      this.$emit("albums-open");
    },

    hideContext: function (e) {
      if (this.contextMenuShown) {
        e.stopPropagation();
      }
    },

    renderVolume: function (v: number): string {
      return Math.round(v * 100) + "%";
    },
    enterTooltip: function (t: string) {
      this.helpTooltip = t;
    },

    leaveTooltip: function (t: string) {
      if (t === this.helpTooltip) {
        this.helpTooltip = "";
      }
    },

    showConfig: function (e) {
      this.displayConfig = !this.displayConfig;
      e.stopPropagation();
    },

    getThumbnailForTime: function (time: number) {
      if (
        this.duration <= 0 ||
        !this.metadata ||
        !this.metadata.video_previews ||
        !this.metadata.video_previews_interval
      ) {
        return "";
      }
      let thumbCount = Math.floor(
        this.duration / this.metadata.video_previews_interval
      );

      if (thumbCount <= 0) {
        return "";
      }

      var part = Math.floor(time / this.metadata.video_previews_interval);
      if (part > thumbCount) {
        part = thumbCount;
      }

      return GetAssetURL(this.metadata.video_previews.replace("{INDEX}", "" + part));
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

      if (typeof this.currentTime === "number" && !isNaN(this.currentTime) && isFinite(this.currentTime) && this.currentTime >= 0) {
        this.getVideoElement().currentTime = Math.min(this.currentTime, this.duration);
      }

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
            this.requiresRefresh = true;
          }.bind(this)
        );
      }
    },
    onWaitForBuffer: function (b) {
      this.loading = b;
    },
    onEnded: function () {
      this.loading = false;
      PlayerPreferences.SetInitialTime(this.mid, 0);
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
        if (e.touches && e.touches.length > 0) {
          this.onTimelineSkip(e.touches[0].pageX, e.touches[0].pageY);
        } else {
          this.onTimelineSkip(e.pageX, e.pageY);
        }
        this.timelineGrabbed = false;
      }
    },
    playerMouseMove: function (e) {
      this.interactWithControls();

      if (this.timelineGrabbed) {
        if (e.touches && e.touches.length > 0) {
          this.onTimelineSkip(e.touches[0].pageX, e.touches[0].pageY);
        } else {
          this.onTimelineSkip(e.pageX, e.pageY);
        }
      }
    },
    mouseLeavePlayer: function () {
      this.timelineGrabbed = false;
      if (!this.playing) return;
      this.showControls = false;
      this.volumeShown = isTouchDevice();
      this.helpTooltip = "";
      this.displayConfig = false;
    },

    checkPlayerSize() {
      const rect = this.$el.getBoundingClientRect();
      const width = rect.width;
      const height = rect.height;

      if (width < 480 || height < 360) {
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
          this.helpTooltip = "";
          this.displayConfig = false;
        }
      }

      var video = this.getVideoElement();

      if (video && video.buffered.length > 0) {
        this.bufferedTime = video.buffered.end(video.buffered.length - 1);
      } else {
        this.bufferedTime = 0;
      }

      if (this.tooltipShown) {
        var tooltip = this.$el.querySelector(".player-tooltip");
        if (tooltip) {
          var x = this.tooltipEventX;
          var toolTipWidth = tooltip.getBoundingClientRect().width;
          var leftPlayer = this.$el.getBoundingClientRect().left;
          var widthPlayer = this.$el.getBoundingClientRect().width;

          x = x - Math.floor(toolTipWidth / 2);
          if (x + toolTipWidth > leftPlayer + widthPlayer - 20) {
            x = leftPlayer + widthPlayer - 20 - toolTipWidth;
          }
          if (x < leftPlayer + 10) {
            x = leftPlayer + 10;
          }
          this.tooltipX = x - leftPlayer;
        }
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
      this.helpTooltip = "";
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
      if (this.displayConfig || this.contextMenuShown) {
        this.displayConfig = false;
        this.contextMenuShown = false;
      } else {
        this.togglePlay();
      }
      this.interactWithControls();
    },

    play: function () {
      if (this.requiresRefresh) {
        this.requiresRefresh = false;
        MediaController.Load();
        return;
      }
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

      if (!video.ended) {
        PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
      }
      
      this.lastTimeChangedEvent = Date.now();

      this.interactWithControls();
    },

    toggleFullScreen: function () {
      if (!this.fullscreen) {
        openFullscreen();
      } else {
        closeFullscreen();
      }
      this.fullScreen = !this.fullScreen;
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

    grabTimeline: function (e) {
      this.timelineGrabbed = true;
      if (e.touches && e.touches.length > 0) {
        this.onTimelineSkip(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.onTimelineSkip(e.pageX, e.pageY);
      }
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
      var x = event.pageX;
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
      }

      this.tooltipShown = true;
      this.tooltipText = this.renderTime(time);
      var oldTooltipImage = this.tooltipImage;
      this.tooltipImage = this.getThumbnailForTime(time);
      if (oldTooltipImage !== this.tooltipImage) {
        this.tooltipImageInvalid = false;
      }
      this.tooltipEventX = x;

      nextTick(this.tick.bind(this));
    },
    renderTime: function (s: number): string {
      return renderTimeSeconds(s);
    },

    setTime: function (time, save) {
      time = Math.max(0, time);
      time = Math.min(time, this.duration);

      if (isNaN(time) || !isFinite(time) || time < 0) {
        return;
      }

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
      var shifting = event.shiftKey;
      switch (event.key) {
        case "M":
        case "m":
          this.toggleMuted();
          this.volumeShown = true;
          this.helpTooltip = "volume";
          break;
        case " ":
        case "K":
        case "k":
          this.togglePlay();
          break;
        case "ArrowUp":
          this.changeVolume(Math.min(1, this.volume + 0.05));
          this.volumeShown = true;
          this.helpTooltip = "volume";
          break;
        case "ArrowDown":
          this.changeVolume(Math.max(0, this.volume - 0.05));
          this.volumeShown = true;
          this.helpTooltip = "volume";
          break;
        case "F":
        case "f":
          this.toggleFullScreen();
          break;
        case "J":
        case "j":
        case "ArrowRight":
          if (shifting) {
            this.goNext();
          } else {
            this.setTime(this.currentTime + 5, true);
          }
          break;
        case "L":
        case "l":
        case "ArrowLeft":
          if (shifting) {
            this.goPrev();
          } else {
            this.setTime(this.currentTime - 5, true);
          }
          break;
        case ".":
          if (!this.playing) {
            this.setTime(this.currentTime - 1 / 30);
          }
          break;
        case ",":
          if (!this.playing) {
            this.setTime(this.currentTime + 1 / 30);
          }
          break;
        case "Home":
          this.setTime(0, true);
          break;
        case "End":
          this.setTime(this.duration, true);
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
      this.loop = !this.next;
      this.currentResolution = PlayerPreferences.GetResolutionIndex(
        this.metadata
      );
      this.loading = true;
      this.playing = true;
      this.setVideoURL();
    },

    setVideoURL() {
      nextTick(() => {
        this.$el.focus();
      });
      if (!this.metadata) {
        this.videoURL = "";
        this.duration = 0;
        this.loading = false;
        return;
      }

      if (this.currentResolution < 0) {
        if (this.metadata.encoded) {
          this.videoURL = GetAssetURL(this.metadata.url);
          this.videoPending = false;
          this.videoPendingTask = 0;
        } else {
          this.videoURL = "";
          this.videoPending = true;
          this.videoPendingTask = this.metadata.task;
          this.duration = 0;
          this.loading = false;
        }
      } else {
        if (
          this.metadata.resolutions &&
          this.metadata.resolutions.length > this.currentResolution
        ) {
          let res = this.metadata.resolutions[this.currentResolution];
          if (res.ready) {
            this.videoURL = GetAssetURL(res.url);
            this.videoPending = false;
            this.videoPendingTask = 0;
          } else {
            this.videoURL = "";
            this.videoPending = true;
            this.videoPendingTask = res.task;
            this.duration = 0;
            this.loading = false;
          }
        } else {
          this.videoURL = "";
          this.videoPending = true;
          this.videoPendingTask = 0;
          this.duration = 0;
          this.loading = false;
        }
      }
    },

    onFeedBackAnimationEnd: function () {
      this.feedback = "";
    },

    onTooltipImageError: function () {
      this.tooltipImageInvalid = true;
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
      this.expandedTitle = false;
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

.video-player.full-screen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 100;
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

.player-timeline-back,
.player-timeline-buffer,
.player-timeline-current {
  height: 3px;
  transition: height 0.1s;
}

.player-timeline-back {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  background: rgba(255, 255, 255, 0.25);
}

.player-timeline-buffer {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  background: rgba(255, 255, 255, 0.5);
}

.player-timeline-current {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  background: red;
}

.player-timeline:hover .player-timeline-back,
.player-timeline:hover .player-timeline-buffer,
.player-timeline:hover .player-timeline-current {
  height: 5px;
}

.player-timeline-thumb {
  border-radius: 50%;
  width: 15px;
  height: 15px;
  background: red;
  position: absolute;
  bottom: -5px;
  left: -7px;
  display: none;
}

.player-timeline:hover .player-timeline-thumb {
  display: block;
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

.player-tooltip {
  background: rgba(0, 0, 0, 0.75);
  color: white;
  padding: 0.5rem 0.75rem;
  position: absolute;
  bottom: 80px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  max-width: 50%;
}

.player-tooltip-image {
  height: 108px;
  padding-bottom: 0.5rem;
}

.player-min .player-tooltip-image {
  height: 72px;
}

.player-min .player-tooltip {
  bottom: 50px;
}

.player-helptip-left {
  left: 8px;
}

.player-helptip-right {
  right: 8px;
}
</style>