<template>
  <div
    tabindex="-1"
    class="empty-player"
    :class="{
      'player-min': minPlayer,
      'full-screen': fullscreen,
    }"
    @dblclick="toggleFullScreen"
    @keydown="onKeyPress"
  >

    <div class="player-loader" v-if="status === 'loading' || (status === 'none' && albumloading)">
      <div class="player-lds-ring">
        <div></div>
        <div></div>
        <div></div>
        <div></div>
      </div>
    </div>

    <div class="player-error-container" v-if="status === '404'">
      <div class="player-info-icon"><i class="fas fa-ban"></i></div>
      <div class="player-error">{{ $t('Media asset does not exist or was removed from the vault') }}</div>
    </div>

    <div class="player-error-container" v-if="status === 'none' && !albumloading">
      <div class="player-info-icon"><i class="fas fa-list-ol"></i></div>
      <div class="player-info">{{ $t('The album is empty') }}</div>
      <div class="player-info">{{ $t('Browse the vault in order to add media to it') }}</div>
    </div>

    <div
      class="player-controls"
      @dblclick="stopPropagationEvent"
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
          disabled
          type="button"
          :title="$t('Play')"
          class="player-btn"
        >
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
      </div>

      <div class="player-controls-right">

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


    <PlayerTopBar
      :mid="mid"
      :metadata="null"
      :shown="true"
      :fullscreen="fullscreen"
      v-model:expanded="expandedTitle"
      v-model:albumexpanded="expandedAlbum"
      :inalbum="inalbum"
    ></PlayerTopBar>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";

import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  components: {
    PlayerMediaChangePreview,
    PlayerTopBar,
  },
  name: "EmptyPlayer",
  emits: ["gonext", "goprev", "update:fullscreen"],
  props: {
    mid: Number,
    status: String,

    albumloading: Boolean,

    fullscreen: Boolean,

    rtick: Number,

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
      minPlayer: false,

      helpTooltip: "",

      expandedTitle: false,
      expandedAlbum: false,
    };
  },
  methods: {
    enterTooltip: function (t: string) {
      this.helpTooltip = t;
    },

    leaveTooltip: function (t: string) {
      if (t === this.helpTooltip) {
        this.helpTooltip = "";
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

    mouseLeavePlayer: function () {
      this.helpTooltip = "";
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
    },

    leaveControls: function () {
      this.helpTooltip = "";
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
        case "F":
        case "f":
          this.toggleFullScreen();
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
      }
    },
  },
  mounted: function () {
    // Load player preferences

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

    nextTick(() => {
        this.$el.focus();
      });
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
      this.expandedTitle = false;
      nextTick(() => {
        this.$el.focus();
      });
    },
  },
});
</script>

<style>
.empty-player {
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

.empty-player.no-controls {
  cursor: none;
}

.empty-player.full-screen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 100;
}

.player-error-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  padding: 0.5rem;
  opacity: 0.75;
}

.player-info {
  font-size: x-large;
  padding-bottom: 0.5rem;
}

.player-info-icon {
  font-size: 48px;
  padding-bottom: 1rem;
}

.player-error {
  font-size: x-large;
  padding-bottom: 0.5rem;
}

</style>