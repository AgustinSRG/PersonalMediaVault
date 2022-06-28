<template>
  <div
    tabindex="0"
    class="audio-player"
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
    <audio
      :src="audioURL || undefined"
      playsinline
      webkit-playsinline
      x-webkit-airplay="allow"
      crossorigin="anonymous"
      :muted="muted"
      :volume.prop="volume"
      :playbackRate.prop="speed"
      @ended="onEnded"
      @timeupdate="onVideoTimeUpdate"
      @canplay="onCanPlay"
      @loadedmetadata="onLoadMetaData"
      @waiting="onWaitForBuffer(true)"
      @playing="onWaitForBuffer(false)"
      @play="setupAudioRenderer"
      @pause="clearAudioRenderer"
    ></audio>

    <canvas v-if="audioURL"></canvas>

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
      <div class="player-tooltip-text">{{ tooltipText }}</div>
    </div>

    <AudioPlayerConfig
      v-model:shown="displayConfig"
      v-model:speed="speed"
      v-model:loop="loop"
      v-model:animcolors="animationColors"
      :rtick="internalTick"
      :metadata="metadata"
      @update:animcolors="onUpdateAnimColors"
      @enter="enterControls"
      @leave="leaveControls"
    ></AudioPlayerConfig>

    <PlayerTopBar
      v-if="metadata"
      :mid="mid"
      :metadata="metadata"
      :shown="showControls"
      :fullscreen="fullscreen"
      :expanded="expandedTitle"
      :albumexpanded="expandedAlbum"
    ></PlayerTopBar>

    <PlayerContextMenu
      type="audio"
      v-model:shown="contextMenuShown"
      :x="contextMenuX"
      :y="contextMenuY"
      v-model:loop="loop"
      :url="audioURL"
    ></PlayerContextMenu>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import VolumeControl from "./VolumeControl.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { renderTimeSeconds } from "../../utils/time-utils";
import { isTouchDevice } from "@/utils/touch";
import AudioPlayerConfig from "./AudioPlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import { GetAssetURL } from "@/utils/request";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  components: {
    VolumeControl,
    AudioPlayerConfig,
    PlayerMediaChangePreview,
    PlayerTopBar,
    PlayerContextMenu,
  },
  name: "AudioPlayer",
  emits: ["gonext", "goprev", "ended", "update:fullscreen", "albums-open"],
  props: {
    mid: Number,
    metadata: Object,
    rtick: Number,

    fullscreen: Boolean,

    next: Object,
    prev: Object,
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

      audioURL: "",
      audioPending: false,
      audioPendingTask: 0,

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

      showControls: true,
      lastControlsInteraction: Date.now(),
      mouseInControls: false,

      loop: false,

      volume: 1,
      muted: false,
      volumeShown: isTouchDevice(),
      internalTick: 0,

      speed: 1,

      feedback: "",

      helpTooltip: "",

      animationColors: "",

      expandedTitle: false,
      expandedAlbum: false,

      contextMenuX: 0,
      contextMenuY: 0,
      contextMenuShown: false,
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

    onUpdateAnimColors: function () {
      PlayerPreferences.SetAudioAnimationStyle(this.animationColors);
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
      this.duration = this.getAudioElement().duration;

      this.getAudioElement().currentTime = this.currentTime;
      this.getAudioElement().playbackRate = this.speed;
    },
    onVideoTimeUpdate: function () {
      if (this.loading) return;
      this.currentTime = this.getAudioElement().currentTime;
      this.duration = this.getAudioElement().duration;
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
      var promise = this.getAudioElement().play();
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
      PlayerPreferences.SetInitialTime(this.mid, 0);
      this.loading = false;
      if (this.loop) {
        this.getAudioElement().currentTime = 0;
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

      var audio = this.getAudioElement();

      if (audio && audio.buffered.length > 0) {
        this.bufferedTime = audio.buffered.end(audio.buffered.length - 1);
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

    getAudioElement() {
      return this.$el.querySelector("audio");
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
      var audio = this.getAudioElement();
      this.playing = true;
      if (audio) {
        audio.play();
      }
    },
    pause: function () {
      var audio = this.getAudioElement();
      this.playing = false;

      if (audio) {
        audio.pause();
      }

      if (!audio.ended) {
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
      this.tooltipEventX = x;

      nextTick(this.tick.bind(this));
    },
    renderTime: function (s: number): string {
      return renderTimeSeconds(s);
    },

    setTime: function (time: number, save: boolean) {
      time = Math.max(0, time);
      time = Math.min(time, this.duration);

      if (isNaN(time) || !isFinite(time) || time < 0) {
        return;
      }

      this.currentTime = time;

      var video = this.getAudioElement();

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

    initializeAudio() {
      if (!this.metadata) {
        return;
      }
      this.currentTime = PlayerPreferences.GetInitialTime(this.mid);
      this.duration = 0;
      this.speed = 1;
      this.loop = !this.next;
      this.loading = true;
      this.playing = true;
      this.clearAudioRenderer();
      this.setAudioURL();
    },

    setAudioURL() {
      nextTick(() => {
        this.$el.focus();
      });
      if (!this.metadata) {
        this.audioURL = "";
        this.duration = 0;
        this.loading = false;
        this.clearAudioRenderer();
        this.getAudioElement().load();
        return;
      }

      if (this.metadata.encoded) {
        this.audioURL = GetAssetURL(this.metadata.url);
        this.audioPending = false;
        this.audioPendingTask = 0;
        this.getAudioElement().load();
      } else {
        this.audioURL = "";
        this.audioPending = true;
        this.audioPendingTask = this.metadata.task;
        this.duration = 0;
        this.loading = false;
        this.clearAudioRenderer();
        this.getAudioElement().load();
      }
    },

    clearAudioRenderer: function () {
      if (this.$options.rendererTimer) {
        clearInterval(this.$options.rendererTimer);
        this.$options.rendererTimer = null;
      }

      if (this.$options.audioSource) {
        this.$options.audioSource.disconnect();
      }

      if (this.$options.audioAnalyser) {
        this.$options.audioAnalyser.disconnect();
      }

      this.$options.audioAnalyser = null;
    },

    setupAudioRenderer: function () {
      if (!this.$options.audioContext) {
        this.setupAudioContext();
      }

      this.clearAudioRenderer();

      if (this.audioURL) {
        const context = this.$options.audioContext;
        const source = this.$options.audioSource;

        const analyser = context.createAnalyser();

        this.$options.audioSource = source;
        this.$options.audioAnalyser = analyser;
        source.connect(analyser);
        analyser.connect(context.destination);

        analyser.fftSize = 256;

        this.$options.rendererTimer = setInterval(
          this.audioAnimationFrame.bind(this),
          Math.floor(1000 / 30)
        );
      } else {
        this.$options.audioContext = null;
        this.$options.audioSource = null;
        this.$options.audioAnalyser = null;
      }
    },

    setupAudioContext: function () {
      const context = new AudioContext();
      this.$options.audioContext = context;
      const source = context.createMediaElementSource(this.getAudioElement());
      this.$options.audioSource = source;
    },

    closeAudioContext: function () {
      if (this.$options.audioContext) {
        this.$options.audioContext.close();
      }

      this.$options.audioContext = null;
      this.$options.audioSource = null;
    },

    audioAnimationFrame: function () {
      if (!this.playing) {
        return;
      }

      const analyser = this.$options.audioAnalyser;
      const canvas = this.$el.querySelector("canvas");

      if (!analyser || !canvas) {
        return;
      }

      const rect = this.$el.getBoundingClientRect();
      if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
      }

      const bufferLength = analyser.frequencyBinCount;

      const dataArray = new Uint8Array(bufferLength);

      const WIDTH = canvas.width;
      const HEIGHT = canvas.height;

      let barWidth = (WIDTH / bufferLength) * 2.5;
      let barHeight;
      let x = 0;

      analyser.getByteFrequencyData(dataArray);

      const ctx = canvas.getContext("2d");

      ctx.fillStyle = "#000";
      ctx.fillRect(0, 0, WIDTH, HEIGHT);

      if (this.animationColors === "none") {
        return;
      }

      for (let i = 0; i < bufferLength; i++) {
        barHeight = dataArray[i];

        switch (this.animationColors) {
          case "gradient":
            {
              let r = barHeight + 25 * (i / bufferLength);
              let g = 250 * (i / bufferLength);
              let b = 50;
              ctx.fillStyle = "rgb(" + r + "," + g + "," + b + ")";
            }
            break;
          default:
            ctx.fillStyle = "rgba(255, 255, 255, 0.5)";
        }

        let trueHeight = Math.floor(HEIGHT * (barHeight / 255));

        ctx.fillRect(x, HEIGHT - trueHeight, barWidth, trueHeight);

        x += barWidth + 1;
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
    this.animationColors = PlayerPreferences.AudioAnimationStyle;

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

    this.initializeAudio();
  },
  beforeUnmount: function () {
    this.audioURL = "";
    clearInterval(this.$options.timer);

    this.clearAudioRenderer();
    this.closeAudioContext();

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
      this.initializeAudio();
    },
    audioURL: function () {
      if (this.audioURL) {
        this.loading = true;
      }
    },
  },
});
</script>

<style>
.audio-player {
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

.audio-player canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.audio-player audio {
  visibility: hidden;
}

.audio-player:focus {
  outline: none;
}

.audio-player.no-controls {
  cursor: none;
}

.audio-player.full-screen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 100;
}
</style>