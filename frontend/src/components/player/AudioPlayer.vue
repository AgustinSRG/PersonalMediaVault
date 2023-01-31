<template>
  <div
    class="audio-player player-settings-no-trap"
    :class="{
      'player-min': minPlayer,
      'no-controls': !showControls,
      'full-screen': fullscreen,
    }"
    @mousemove="playerMouseMove"
    @click="clickPlayer"
    @mousedown="hideContext"
    @touchstart.passive="hideContext"
    @dblclick="toggleFullScreen"
    @mouseleave="mouseLeavePlayer"
    @mouseup="playerMouseUp"
    @touchmove="playerMouseMove"
    @touchend.passive="playerMouseUp"
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

    <PlayerEncodingPending
      v-if="!loading && !audioURL && audioPending"
      :mid="mid"
      :tid="audioPendingTask"
      :res="-1"
    ></PlayerEncodingPending>

    <div
      class="player-subtitles-container"
      :class="{ 'controls-hidden': !showControls }"
    >
      <div
        class="player-subtitles"
        v-if="subtitles"
        v-html="subtitles"
        :class="{
          'player-subtitles-s': subtitlesSize === 's',
          'player-subtitles-m': subtitlesSize === 'm',
          'player-subtitles-l': subtitlesSize === 'l',
          'player-subtitles-xl': subtitlesSize === 'xl',
          'player-subtitles-xxl': subtitlesSize === 'xxl',

          'player-subtitles-bg-0': subtitlesBg === '0',
          'player-subtitles-bg-25': subtitlesBg === '25',
          'player-subtitles-bg-50': subtitlesBg === '50',
          'player-subtitles-bg-75': subtitlesBg === '75',
          'player-subtitles-bg-100': subtitlesBg === '100',
        }"
      ></div>
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

        <div class="player-time-label-container" :class="{'in-album': !!next || !!prev}" v-if="!minPlayer">
          <span
            >{{ renderTime(currentTime) }} / {{ renderTime(duration) }}</span
          >
          <span v-if="currentTimeSlice" class="times-slice-name"><b class="separator"> - </b>{{currentTimeSliceName}}</span>
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
          class="player-btn player-settings-no-trap"
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
      @click="clickTimeline"
      @mousedown="grabTimeline"
      @toutchstart.passive="grabTimeline"
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
        v-for="ts in timeSlices"
        :key="ts"
        class="player-timeline-split"
        :class="{ 'start-split': ts.start <= 0 }"
        :style="{ left: getTimelineBarWidth(ts.start, duration) }"
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
      <div v-if="tooltipTimeSlice" class="player-tooltip-text">
        {{ tooltipTimeSlice }}
      </div>
    </div>

    <AudioPlayerConfig
      v-model:shown="displayConfig"
      v-model:speed="speed"
      v-model:loop="loop"
      v-model:nextend="nextend"
      v-model:animcolors="animationColors"
      v-model:subsize="subtitlesSize"
      v-model:subbg="subtitlesBg"
      v-model:subhtml="subtitlesHTML"
      :rtick="internalTick"
      :metadata="metadata"
      @update:animcolors="onUpdateAnimColors"
      @update:subhtml="onUpdateSubHTML"
      @update:nextend="onUpdateNextEnd"
      @enter="enterControls"
      @leave="leaveControls"
    ></AudioPlayerConfig>

    <PlayerTopBar
      v-if="metadata"
      :mid="mid"
      :metadata="metadata"
      :shown="showControls"
      :fullscreen="fullscreen"
      v-model:expanded="expandedTitle"
      v-model:albumexpanded="expandedAlbum"
      :inalbum="inalbum"
      @clickplayer="clickControls"
    ></PlayerTopBar>

    <PlayerContextMenu
      type="audio"
      v-model:shown="contextMenuShown"
      :x="contextMenuX"
      :y="contextMenuY"
      v-model:loop="loop"
      :url="audioURL"
      @stats="openStats"
      v-model:sliceloop="sliceLoop"
      :hasslices="timeSlices && timeSlices.length > 0"
    ></PlayerContextMenu>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import VolumeControl from "./VolumeControl.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import PlayerEncodingPending from "./PlayerEncodingPending.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import {
  findTimeSlice,
  normalizeTimeSlices,
  renderTimeSeconds,
} from "../../utils/time-utils";
import { isTouchDevice } from "@/utils/touch";
import AudioPlayerConfig from "./AudioPlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import { GetAssetURL } from "@/utils/request";
import { useVModel } from "../../utils/vmodel";
import { MediaController } from "@/control/media";
import { SubtitlesController } from "@/control/subtitles";
import {
  sanitizeSubtitlesHTML,
  getUniqueSubtitlesLoadTag,
} from "@/utils/subtitles-html";
import { htmlToText } from "@/utils/text";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { KeyboardManager } from "@/control/keyboard";
import { AuthController } from "@/control/auth";
import { AppPreferences } from "@/control/app-preferences";

export default defineComponent({
  components: {
    VolumeControl,
    AudioPlayerConfig,
    PlayerMediaChangePreview,
    PlayerTopBar,
    PlayerContextMenu,
    PlayerEncodingPending,
  },
  name: "AudioPlayer",
  emits: [
    "gonext",
    "goprev",
    "ended",
    "update:fullscreen",
    "albums-open",
    "stats-open",
  ],
  props: {
    mid: Number,
    metadata: Object,
    rtick: Number,

    fullscreen: Boolean,

    next: Object,
    prev: Object,
    inalbum: Boolean,

    canwrite: Boolean,
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

      canSaveTime: true,

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
      tooltipTimeSlice: "",
      tooltipX: 0,
      tooltipEventX: 0,

      showControls: true,
      lastControlsInteraction: Date.now(),
      mouseInControls: false,

      loop: false,
      nextend: false,

      sliceLoop: false,

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

      requiresRefresh: false,

      subtitles: "",
      subtitlesStart: -1,
      suntitlesEnd: -1,
      subtitlesSize: "l",
      subtitlesBg: "75",
      subtitlesHTML: false,

      timeSlices: [],
      currentTimeSlice: null,
      currentTimeSliceName: "",
      currentTimeSliceStart: 0,
      currentTimeSliceEnd: 0,

      theme: AppPreferences.Theme,
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

    openStats: function () {
      this.$emit("stats-open");
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
      this.contextMenuShown = false;
      if (e) {
        e.stopPropagation();
      }
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
      const audioElement = this.getAudioElement();

      if (!audioElement) {
        return;
      }

      this.duration = audioElement.duration;

      if (
        typeof this.currentTime === "number" &&
        !isNaN(this.currentTime) &&
        isFinite(this.currentTime) &&
        this.currentTime >= 0
      ) {
        audioElement.currentTime = Math.min(this.currentTime, this.duration);
        this.updateSubtitles();
        this.updateCurrentTimeSlice();
      }

      audioElement.playbackRate = this.speed;
    },
    onVideoTimeUpdate: function () {
      if (this.loading) return;

      const audioElement = this.getAudioElement();

      if (!audioElement) {
        return;
      }

      this.currentTime = audioElement.currentTime;
      this.duration = audioElement.duration;

      if (this.canSaveTime && Date.now() - this.lastTimeChangedEvent > 5000) {
        PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
        this.lastTimeChangedEvent = Date.now();
      }

      this.updateSubtitles();
      this.updateCurrentTimeSlice();
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
            this.requiresRefresh = true;
          }.bind(this)
        );
      }
    },
    onWaitForBuffer: function (b) {
      this.loading = b;
    },
    onEnded: function () {
      if (this.canSaveTime) {
        PlayerPreferences.SetInitialTime(this.mid, 0);
      }
      this.loading = false;
      if (this.loop) {
        const audioElement = this.getAudioElement();
        if (audioElement) {
          this.getAudioElement().currentTime = 0;
        }
      } else {
        this.pause();
        this.ended = true;
        if (this.nextend) {
          this.goNext();
        }
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
      if (!this.playing || this.expandedTitle || this.expandedAlbum) return;
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

      if (
        this.showControls &&
        !this.mouseInControls &&
        this.playing &&
        !this.expandedTitle &&
        !this.expandedAlbum
      ) {
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
      if (this.requiresRefresh) {
        this.requiresRefresh = false;
        MediaController.Load();
        return;
      }
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

      if (this.canSaveTime && audio && !audio.ended) {
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

    clickTimeline: function (e) {
      this.displayConfig = false;
      this.contextMenuShown = false;
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
      this.tooltipTimeSlice = this.findTimeSliceName(time);
      this.tooltipEventX = x;

      nextTick(this.tick.bind(this));
    },
    renderTime: function (s: number): string {
      return renderTimeSeconds(s);
    },

    setTime: function (time: number, save: boolean) {
      this.currentTimeSlice = null;
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

      if (this.canSaveTime && save) {
        PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
        this.lastTimeChangedEvent = Date.now();
      }
      if (time < this.duration) {
        this.ended = false;
      }

      this.updateSubtitles();
      this.updateCurrentTimeSlice();
    },

    onKeyPress: function (event: KeyboardEvent): boolean {
      if (
        AuthController.Locked ||
        !AppStatus.IsPlayerVisible() ||
        !event.key ||
        event.ctrlKey
      ) {
        return false;
      }
      let caught = true;
      const shifting = event.shiftKey;
      switch (event.key) {
        case "A":
        case "a":
          this.manageAlbums();
          break;
        case "S":
        case "s":
          this.showConfig();
          break;
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
          if (event.altKey || shifting) {
            caught = false;
          } else {
            this.toggleFullScreen();
          }
          break;
        case "ArrowRight":
          if (shifting || event.altKey) {
            if (this.next) {
              this.goNext();
            } else {
              caught = false;
            }
          } else {
            this.setTime(this.currentTime + 5, true);
          }
          break;
        case "ArrowLeft":
          if (shifting || event.altKey) {
            if (this.prev) {
              this.goPrev();
            } else {
              caught = false;
            }
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
          if (event.altKey || shifting) {
            caught = false;
          } else {
            this.setTime(0, true);
          }
          break;
        case "End":
          if (event.altKey || shifting) {
            caught = false;
          } else {
            this.setTime(this.duration, true);
          }
          break;
        case "PageDown":
          if (this.prev) {
            this.goPrev();
          } else {
            caught = false;
          }
          break;
        case "PageUp":
          if (this.next) {
            this.goNext();
          } else {
            caught = false;
          }
          break;
        case "l":
        case "L":
          if (event.altKey || shifting) {
            caught = false;
          } else {
            this.loop = !this.loop;
            if (this.loop) {
              AppEvents.Emit("snack", this.$t("Loop enabled"));
            } else {
              AppEvents.Emit("snack", this.$t("Loop disabled"));
            }
          }
          break;
        case "b":
        case "B":
          if (this.currentTimeSlice) {
            this.setTime(this.currentTimeSlice.start, true)
          }
          break;
        case "j":
        case "J":
          if (this.currentTimeSlice) {
            this.setTime(this.currentTimeSlice.end, true)
          }
          break;
        default:
          caught = false;
      }

      if (caught) {
        this.interactWithControls();
      }

      return caught;
    },

    initializeAudio() {
      if (!this.metadata) {
        return;
      }
      this.canSaveTime = !this.metadata.force_start_beginning;
      this.timeSlices = normalizeTimeSlices(
        (this.metadata.time_slices || []).sort((a, b) => {
          if (a.time < b.time) {
            return -1;
          } else if (a.time > b.time) {
            return 1;
          } else {
            return 0;
          }
        }),
        this.metadata.duration
      );
      this.currentTimeSlice = null;
      this.currentTimeSliceName = "";
      this.currentTimeSliceStart = 0;
      this.currentTimeSliceEnd = 0;
      this.sliceLoop = false;
      this.currentTime = this.canSaveTime
        ? PlayerPreferences.GetInitialTime(this.mid)
        : 0;
      this.duration = 0;
      this.speed = 1;
      this.loop = AppStatus.CurrentAlbum < 0 || !this.nextend;
      this.loading = true;
      this.playing = true;
      this.clearAudioRenderer();
      this.setAudioURL();
    },

    onClearURL: function () {
      const audioElem = this.$el.querySelector("audio");
      if (audioElem) {
        audioElem.src = "";
      }
    },

    setAudioURL() {
      if (!this.metadata) {
        this.audioURL = "";
        this.onClearURL();
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
        this.onClearURL();
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

      ctx.fillStyle = this.theme === "light" ? "#fff" : "#000";
      ctx.fillRect(0, 0, WIDTH, HEIGHT);

      if (this.animationColors === "none") {
        return;
      }

      for (let i = 0; i < bufferLength; i++) {
        barHeight = dataArray[i];

        switch (this.animationColors) {
          case "gradient":
            {
              if (this.theme === "light") {
                let r = Math.min(255, barHeight + 80 * (i / bufferLength));
                let g = 250 * (i / bufferLength);
                let b = 180;
                ctx.fillStyle = "rgb(" + r + "," + g + "," + b + ")";
              } else {
                let r = Math.min(255, barHeight + 25 * (i / bufferLength));
                let g = 250 * (i / bufferLength);
                let b = 50;
                ctx.fillStyle = "rgb(" + r + "," + g + "," + b + ")";
              }
            }
            break;
          default:
            ctx.fillStyle =
              this.theme === "light"
                ? "rgba(0, 0, 0, 0.5)"
                : "rgba(255, 255, 255, 0.5)";
        }

        let trueHeight = Math.floor(HEIGHT * (barHeight / 255));

        ctx.fillRect(x, HEIGHT - trueHeight, barWidth, trueHeight);

        x += barWidth + 1;
      }
    },

    onFeedBackAnimationEnd: function () {
      this.feedback = "";
    },

    reloadSubtitles: function () {
      this.subtitles = "";
      this.subtitlesStart = -1;
      this.subtitlesEnd = -1;
      this.updateSubtitles();
    },

    updateSubtitles: function () {
      if (
        this.currentTime >= this.subtitlesStart &&
        this.currentTime <= this.subtitlesEnd
      ) {
        return;
      }
      const sub = SubtitlesController.GetSubtitlesLine(this.currentTime);
      if (sub) {
        if (this.subtitlesHTML) {
          const subTag = getUniqueSubtitlesLoadTag();
          this.$options.subTag = subTag;
          sanitizeSubtitlesHTML(sub.text).then((text) => {
            if (this.$options.subTag === subTag) {
              this.subtitles = text;
            }
          });
        } else {
          this.$options.subTag = "";
          this.subtitles = htmlToText(sub.text);
        }
        this.subtitlesStart = sub.start;
        this.subtitlesEnd = sub.end;
      } else {
        this.$options.subTag = "";
        this.subtitles = "";
        this.subtitlesStart = 0;
        this.subtitlesEnd = 0;
      }
    },

    onUpdateSubHTML: function () {
      PlayerPreferences.SetSubtitlesHTML(this.subtitlesHTML);
      this.reloadSubtitles();
    },

    findTimeSliceName: function (time: number) {
      const slice = findTimeSlice(this.timeSlices, time);
      if (slice) {
        return slice.name;
      } else {
        return "";
      }
    },

    updateCurrentTimeSlice: function () {
      if (this.currentTimeSlice && this.sliceLoop && this.currentTime >= this.currentTimeSlice.end) {
        this.setTime(this.currentTimeSlice.start, false);
        return;
      }
      const slice = findTimeSlice(this.timeSlices, this.currentTime);
      if (slice) {
        this.currentTimeSlice = slice;
        this.currentTimeSliceName = slice.name;
        this.currentTimeSliceStart = slice.start;
        this.currentTimeSliceEnd = slice.end;
      } else {
        this.currentTimeSlice = null;
        this.currentTimeSliceName = "";
        this.currentTimeSliceStart = 0;
        this.currentTimeSliceEnd = 0;
      }
    },

    onUpdateNextEnd: function () {
      PlayerPreferences.SetNextOnEnd(this.nextend);
    },

    themeUpdated: function () {
      this.theme = AppPreferences.Theme;
    },

    handleMediaSessionEvent: function (event: {
      action: string;
      fastSeek: boolean;
      seekTime: number;
      seekOffset: number;
    }) {
      if (!event || !event.action) {
        return;
      }
      switch (event.action) {
        case "play":
          this.play();
          break;
        case "pause":
          this.pause();
          break;
        case "nexttrack":
          if (this.next) {
            this.goNext();
          }
          break;
        case "previoustrack":
          if (this.prev) {
            this.goPrev();
          }
          break;
      }
    },
  },
  mounted: function () {
    // Load player preferences
    this.muted = PlayerPreferences.PlayerMuted;
    this.volume = PlayerPreferences.PlayerVolume;
    this.animationColors = PlayerPreferences.AudioAnimationStyle;
    this.subtitlesSize = PlayerPreferences.SubtitlesSize;
    this.subtitlesBg = PlayerPreferences.SubtitlesBackground;
    this.subtitlesHTML = PlayerPreferences.SubtitlesHTML;
    this.nextend = PlayerPreferences.NextOnEnd;

    this.$options.keyHandler = this.onKeyPress.bind(this);
    KeyboardManager.AddHandler(this.$options.keyHandler, 100);

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

    this.$options.subtitlesReloadH = this.reloadSubtitles.bind(this);
    AppEvents.AddEventListener(
      "subtitles-update",
      this.$options.subtitlesReloadH
    );

    this.$options.themeHandler = this.themeUpdated.bind(this);
    AppEvents.AddEventListener("theme-changed", this.$options.themeHandler);

    this.initializeAudio();

    if (window.navigator && window.navigator.mediaSession) {
      navigator.mediaSession.setActionHandler("play", this.handleMediaSessionEvent.bind(this));
      navigator.mediaSession.setActionHandler("pause", this.handleMediaSessionEvent.bind(this));
      navigator.mediaSession.setActionHandler("nexttrack", this.handleMediaSessionEvent.bind(this));
      navigator.mediaSession.setActionHandler("previoustrack", this.handleMediaSessionEvent.bind(this));
    }
  },
  beforeUnmount: function () {
    this.audioURL = "";
    this.onClearURL();
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

    AppEvents.RemoveEventListener(
      "subtitles-update",
      this.$options.subtitlesReloadH
    );

    AppEvents.RemoveEventListener("theme-changed", this.$options.themeHandler);

    KeyboardManager.RemoveHandler(this.$options.keyHandler);

    if (window.navigator && window.navigator.mediaSession) {
      navigator.mediaSession.setActionHandler("play", null);
      navigator.mediaSession.setActionHandler("pause", null);
      navigator.mediaSession.setActionHandler("nexttrack", null);
      navigator.mediaSession.setActionHandler("previoustrack", null);
    }
  },
  watch: {
    rtick: function () {
      this.internalTick++;
      this.expandedTitle = false;
      this.subtitles = "";
      this.subtitlesStart = -1;
      this.subtitlesEnd = -1;
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
