<template>
  <div
    tabindex="-1"
    class="image-player"
    :class="{
      'player-min': minPlayer,
      'no-controls': !showcontrols,
      'full-screen': fullscreen,
      'bg-black': background === 'black',
      'bg-white': background === 'white',
    }"
    @mousemove="playerMouseMove"
    @click="clickPlayer"
    @dblclick="toggleFullScreen"
    @mouseleave="mouseLeavePlayer"
    @touchmove="playerMouseMove"
    @keydown="onKeyPress"
    @contextmenu="onContextMenu"
    @wheel="onMouseWheel"
  >
    <div class="image-scroller" @mousedown="grabScroll">
      <img
        v-if="imageURL"
        :src="imageURL"
        :key="rtick"
        @load="onImageLoaded"
        @error="onImageLoaded"
        :style="{
          width: imageWidth,
          height: imageHeight,
          top: imageTop,
          left: imageLeft,
        }"
      />
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
      v-if="!loading && !imageURL && imagePending"
      :mid="mid"
      :tid="imagePendingTask"
      :res="currentResolution"
    ></PlayerEncodingPending>

    <div
      class="player-controls"
      :class="{ hidden: !showcontrols }"
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

        <button disabled type="button" :title="$t('Play')" class="player-btn">
          <i class="fas fa-play"></i>
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

        <ScaleControl
          ref="scaleControl"
          :min="minPlayer"
          :width="minPlayer ? 70 : 100"
          v-model:fit="fit"
          v-model:scale="scale"
          v-model:expanded="scaleShown"
          @update:scale="onUserScaleUpdated"
          @update:fit="onUserFitUpdated"
          @enter="enterTooltip('scale')"
          @leave="leaveTooltip('scale')"
        ></ScaleControl>
      </div>

      <div class="player-controls-right">
        <button
          v-if="canwrite"
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
      v-if="helpTooltip === 'scale'"
      class="player-tooltip player-helptip-left"
    >
      {{ $t("Scale") }} ({{ fit ? $t("Fit") : renderScale(scale) }})
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

    <ImagePlayerConfig
      v-model:shown="displayConfig"
      v-model:resolution="currentResolution"
      v-model:background="background"
      @update:resolution="onResolutionUpdated"
      @update:background="onBackgroundChanged"
      @update-autonext="setupAutoNextTimer"
      :rtick="internalTick"
      :metadata="metadata"
      @enter="enterControls"
      @leave="leaveControls"
    ></ImagePlayerConfig>

    <PlayerTopBar
      v-if="metadata"
      :mid="mid"
      :metadata="metadata"
      :shown="showcontrols"
      :fullscreen="fullscreen"
      v-model:expanded="expandedTitle"
      v-model:albumexpanded="expandedAlbum"
      :inalbum="inalbum"
      @clickplayer="clickControls"
    ></PlayerTopBar>

    <PlayerContextMenu
      type="image"
      v-model:shown="contextMenuShown"
      :x="contextMenuX"
      :y="contextMenuY"
      v-model:fit="fit"
      @update:fit="onUserFitUpdated"
      :url="imageURL"
      v-model:controls="showControls"
      @close="focusPlayer"
    ></PlayerContextMenu>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import ScaleControl from "./ScaleControl.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import PlayerEncodingPending from "./PlayerEncodingPending.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { isTouchDevice } from "@/utils/touch";
import ImagePlayerConfig from "./ImagePlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import { GetAssetURL } from "@/utils/request";
import { useVModel } from "../../utils/vmodel";

const SCALE_RANGE = 2;
const SCALE_RANGE_PERCENT = SCALE_RANGE * 100;
const SCALE_STEP = 0.1 / SCALE_RANGE;

export default defineComponent({
  components: {
    ScaleControl,
    ImagePlayerConfig,
    PlayerMediaChangePreview,
    PlayerTopBar,
    PlayerContextMenu,
    PlayerEncodingPending,
  },
  name: "ImagePlayer",
  emits: [
    "gonext",
    "goprev",
    "update:fullscreen",
    "update:showcontrols",
    "albums-open",
  ],
  props: {
    mid: Number,
    metadata: Object,
    rtick: Number,

    showcontrols: Boolean,

    fullscreen: Boolean,

    next: Object,
    prev: Object,
    inalbum: Boolean,

    canwrite: Boolean,
  },
  setup(props) {
    return {
      fullScreen: useVModel(props, "fullscreen"),
      showControls: useVModel(props, "showcontrols"),
    };
  },
  data: function () {
    return {
      loading: true,

      imageURL: "",
      imagePending: false,
      imagePendingTask: 0,

      currentResolution: -1,
      width: 0,
      height: 0,

      minPlayer: false,
      displayConfig: false,

      imageTop: "0",
      imageLeft: "0",
      imageWidth: "auto",
      imageHeight: "auto",

      lastControlsInteraction: Date.now(),
      mouseInControls: false,

      scale: 0,
      fit: true,
      scaleShown: isTouchDevice(),

      background: "default",

      internalTick: 0,

      helpTooltip: "",

      expandedTitle: false,
      expandedAlbum: false,

      contextMenuX: 0,
      contextMenuY: 0,
      contextMenuShown: false,

      scrollGrabbed: false,
      scrollGrabX: 0,
      scrollGrabY: 0,
      scrollGrabTop: 0,
      scrollGrabLeft: 0,
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

    grabScroll: function (e) {
      const scroller = this.$el.querySelector(".image-scroller");

      if (!scroller) {
        return;
      }

      this.scrollGrabTop = scroller.scrollTop;
      this.scrollGrabLeft = scroller.scrollLeft;

      this.scrollGrabbed = true;
      if (e.touches && e.touches.length > 0) {
        this.scrollGrabX = e.touches[0].pageX;
        this.scrollGrabY = e.touches[0].pageY;
      } else {
        this.scrollGrabX = e.pageX;
        this.scrollGrabY = e.pageY;
      }
    },

    moveScrollByMouse: function (x: number, y: number) {
      const scroller = this.$el.querySelector(".image-scroller");

      if (!scroller) {
        return;
      }

      const rect = scroller.getBoundingClientRect();

      const maxScrollLeft = scroller.scrollWidth - rect.width;
      const maxScrollTop = scroller.scrollHeight - rect.height;

      const diffX = x - this.scrollGrabX;
      const diffY = y - this.scrollGrabY;

      scroller.scrollTop = Math.max(
        0,
        Math.min(maxScrollTop, this.scrollGrabTop - diffY)
      );
      scroller.scrollLeft = Math.max(
        0,
        Math.min(maxScrollLeft, this.scrollGrabLeft - diffX)
      );
    },

    dropScroll: function (e) {
      if (!this.scrollGrabbed) {
        return;
      }
      this.scrollGrabbed = false;

      if (e.touches && e.touches.length > 0) {
        this.moveScrollByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.moveScrollByMouse(e.pageX, e.pageY);
      }
    },

    moveScroll: function (e) {
      if (!this.scrollGrabbed) {
        return;
      }

      if (e.touches && e.touches.length > 0) {
        this.moveScrollByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.moveScrollByMouse(e.pageX, e.pageY);
      }
    },

    centerScroll: function () {
      const scroller = this.$el.querySelector(".image-scroller");

      if (!scroller) {
        return;
      }

      scroller.scrollTop =
        (scroller.scrollHeight - scroller.getBoundingClientRect().height) / 2;
      scroller.scrollLeft =
        (scroller.scrollWidth - scroller.getBoundingClientRect().width) / 2;
    },

    computeImageDimensions() {
      if (!this.imageURL) {
        return;
      }

      const scroller = this.$el.querySelector(".image-scroller");

      if (!scroller) {
        return;
      }

      const scrollerDimensions = scroller.getBoundingClientRect();

      const fitDimensions = {
        width: this.width,
        height: this.height,
        fitWidth: true,
      };

      if (scrollerDimensions.width > scrollerDimensions.height) {
        fitDimensions.fitWidth = true;
        fitDimensions.height = scrollerDimensions.height;
        fitDimensions.width =
          (scrollerDimensions.height * this.width) / this.height;

        if (fitDimensions.width > scrollerDimensions.width) {
          fitDimensions.fitWidth = false;
          fitDimensions.width = scrollerDimensions.width;
          fitDimensions.height =
            (scrollerDimensions.width * this.height) / this.width;
        }
      } else {
        fitDimensions.fitWidth = false;
        fitDimensions.width = scrollerDimensions.width;
        fitDimensions.height =
          (scrollerDimensions.width * this.height) / this.width;

        if (fitDimensions.height > scrollerDimensions.height) {
          fitDimensions.fitWidth = true;
          fitDimensions.height = scrollerDimensions.height;
          fitDimensions.width =
            (scrollerDimensions.height * this.width) / this.height;
        }
      }

      if (this.fit) {
        let top = Math.max(
          0,
          (scrollerDimensions.height - fitDimensions.height) / 2
        );
        let left = Math.max(
          0,
          (scrollerDimensions.width - fitDimensions.width) / 2
        );

        this.imageTop = Math.floor(top) + "px";
        this.imageLeft = Math.floor(left) + "px";
        this.imageWidth = Math.floor(fitDimensions.width) + "px";
        this.imageHeight = Math.floor(fitDimensions.height) + "px";
      } else {
        let width = 0;
        let height = 0;

        if (fitDimensions.fitWidth) {
          width = scrollerDimensions.width * (0.5 + this.scale * SCALE_RANGE);
          height = (width * this.height) / this.width;
        } else {
          height = scrollerDimensions.height * (0.5 + this.scale * SCALE_RANGE);
          width = (height * this.width) / this.height;
        }

        let top = Math.max(0, (scrollerDimensions.height - height) / 2);
        let left = Math.max(0, (scrollerDimensions.width - width) / 2);

        this.imageTop = Math.floor(top) + "px";
        this.imageLeft = Math.floor(left) + "px";
        this.imageWidth = Math.floor(width) + "px";
        this.imageHeight = Math.floor(height) + "px";
      }
    },

    renderScale: function (v: number): string {
      return Math.round(50 + v * SCALE_RANGE_PERCENT) + "%";
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

    onUserScaleUpdated() {
      PlayerPreferences.SetScale(this.scale);
      this.computeImageDimensions();
      nextTick(this.centerScroll.bind(this));
    },

    changeScale: function (s: number) {
      this.scale = s;
      this.onUserScaleUpdated();
    },

    onUserFitUpdated() {
      PlayerPreferences.SetFit(this.fit);
      this.computeImageDimensions();
    },

    toggleFit: function () {
      this.fit = !this.fit;
      this.onUserFitUpdated();
    },

    onBackgroundChanged() {
      PlayerPreferences.SetImagePlayerBackground(this.background);
    },

    /* Player events */

    onImageLoaded: function () {
      this.loading = false;
    },

    playerMouseMove: function () {
      this.interactWithControls();
    },
    mouseLeavePlayer: function () {
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
      this.computeImageDimensions();
    },

    interactWithControls() {
      this.lastControlsInteraction = Date.now();
    },

    enterControls: function () {
      this.mouseInControls = true;
    },

    leaveControls: function () {
      this.mouseInControls = false;
      this.helpTooltip = "";
    },

    clickPlayer: function () {
      if (this.displayConfig) {
        this.displayConfig = false;
      }
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

    onKeyPress: function (event) {
      var catched = true;
      switch (event.key) {
        case " ":
        case "K":
        case "k":
          this.toggleFit();
          this.scaleShown = true;
          this.helpTooltip = "scale";
          break;
        case "+":
          this.changeScale(Math.min(1, this.scale + SCALE_STEP));
          this.scaleShown = true;
          this.helpTooltip = "scale";
          this.fit = false;
          break;
        case "-":
          this.changeScale(Math.max(0, this.scale - SCALE_STEP));
          this.scaleShown = true;
          this.helpTooltip = "scale";
          this.fit = false;
          break;
        case "F":
        case "f":
          this.toggleFullScreen();
          break;
        case "C":
        case "c":
          this.showControls = !this.showControls;
          break;
        case "PageDown":
        case "ArrowLeft":
          this.goPrev();
          break;
        case "PageUp":
        case "ArrowRight":
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

    initializeImage() {
      if (!this.metadata) {
        return;
      }
      this.loading = true;
      this.currentResolution = PlayerPreferences.GetResolutionIndexImage(
        this.metadata
      );
      this.setImageURL();
    },

    onResolutionUpdated: function () {
      PlayerPreferences.SetResolutionIndexImage(
        this.metadata,
        this.currentResolution
      );
      this.setImageURL();
    },

    focusPlayer: function () {
      nextTick(() => {
        this.$el.focus();
      });
    },

    setImageURL() {
      nextTick(() => {
        this.$el.focus();
      });
      if (!this.metadata) {
        this.imageURL = "";
        this.loading = false;
        return;
      }

      if (this.currentResolution < 0) {
        if (this.metadata.encoded) {
          this.imageURL = GetAssetURL(this.metadata.url);
          this.imagePending = false;
          this.imagePendingTask = 0;
          this.width = this.metadata.width;
          this.height = this.metadata.height;
          this.setupAutoNextTimer();
        } else {
          this.imageURL = "";
          this.imagePending = true;
          this.imagePendingTask = this.metadata.task;
          this.loading = false;
        }
      } else {
        if (
          this.metadata.resolutions &&
          this.metadata.resolutions.length > this.currentResolution
        ) {
          let res = this.metadata.resolutions[this.currentResolution];
          if (res.ready) {
            this.imageURL = GetAssetURL(res.url);
            this.imagePending = false;
            this.imagePendingTask = 0;
            this.width = res.width;
            this.height = res.height;
            this.setupAutoNextTimer();
          } else {
            this.imageURL = "";
            this.imagePending = true;
            this.imagePendingTask = res.task;
            this.loading = false;
          }
        } else {
          this.imageURL = "";
          this.imagePending = true;
          this.imagePendingTask = 0;
          this.loading = false;
        }
      }

      this.computeImageDimensions();
      nextTick(this.centerScroll.bind(this));
    },

    setupAutoNextTimer: function () {
      if (this.$options.autoNextTimer) {
        clearTimeout(this.$options.autoNextTimer);
        this.$options.autoNextTimer = null;
      }
      const timerS = PlayerPreferences.ImageAutoNext;

      if (isNaN(timerS) || !isFinite(timerS) || timerS <= 0) {
        return;
      }

      if (!this.next) {
        return;
      }

      const ms = timerS * 1000;

      this.$options.autoNextTimer = setTimeout(() => {
        this.$options.autoNextTimer = null;
        this.goNext();
      }, ms);
    },

    onMouseWheel: function (e: WheelEvent) {
      if (e.ctrlKey) {
        e.preventDefault();
        e.stopPropagation();
        if (e.deltaY > 0) {
           this.changeScale(Math.max(0, this.scale - SCALE_STEP));
          this.scaleShown = true;
          this.helpTooltip = "scale";
          this.fit = false;
        } else {
          this.changeScale(Math.min(1, this.scale + SCALE_STEP));
          this.scaleShown = true;
          this.helpTooltip = "scale";
          this.fit = false;
        }
      }
    },
  },
  mounted: function () {
    // Load player preferences
    this.fit = PlayerPreferences.PlayerFit;
    this.scale = PlayerPreferences.PlayerScale;
    this.background = PlayerPreferences.ImagePlayerBackground;

    this.$options.timer = setInterval(
      this.tick.bind(this),
      Math.floor(1000 / 30)
    );

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

    this.$options.dropScrollHandler = this.dropScroll.bind(this);
    document.addEventListener("mouseup", this.$options.dropScrollHandler);

    this.$options.moveScrollHandler = this.moveScroll.bind(this);

    document.addEventListener("mousemove", this.$options.moveScrollHandler);

    this.initializeImage();
  },
  beforeUnmount: function () {
    this.imageURL = "";
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

    document.removeEventListener("mouseup", this.$options.dropScrollHandler);

    document.removeEventListener("mousemove", this.$options.moveScrollHandler);

    if (this.$options.autoNextTimer) {
      clearTimeout(this.$options.autoNextTimer);
      this.$options.autoNextTimer = null;
    }
  },
  watch: {
    rtick: function () {
      this.internalTick++;
      this.expandedTitle = false;
      this.initializeImage();
    },
    imageURL: function () {
      if (this.imageURL) {
        this.loading = true;
      }
    },
    next: function () {
      this.setupAutoNextTimer();
    },
  },
});
</script>

<style>
.image-player {
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

.image-player.no-controls {
  cursor: none;
}

.image-player.full-screen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 100;
}

.image-scroller {
  overflow: auto;
  position: absolute;
  top: 56px;
  bottom: 56px;
  height: auto;
  left: 0;
  width: 100%;
  cursor: move;
}

.image-player.bg-black .image-scroller {
  background: black;
}

.image-player.bg-white .image-scroller {
  background: white;
}

.player-min .image-scroller {
  top: 32px;
  bottom: 32px;
}

/* Custom scroll bar */

/* width */

.image-scroller::-webkit-scrollbar {
  width: 5px;
  height: 3px;
}

/* Track */

.image-scroller::-webkit-scrollbar-track {
  background: #bdbdbd;
}

/* Handle */

.image-scroller::-webkit-scrollbar-thumb {
  background: #757575;
}

.image-scroller img {
  position: absolute;
  pointer-events: none;
}

.no-controls .image-scroller {
  top: 0;
  bottom: 0;
}

.player-min.no-controls .image-scroller {
  top: 0;
  bottom: 0;
}

.image-player .player-tooltip {
  bottom: 64px;
}

.image-player.player-min .player-tooltip {
  bottom: 40px;
}
</style>