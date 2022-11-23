<template>
  <div
    class="empty-player"
    :class="{
      'player-min': minPlayer,
      'full-screen': fullscreen,
    }"
    @dblclick="toggleFullScreen"
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
import { defineComponent } from "vue";

import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { useVModel } from "../../utils/vmodel";
import { KeyboardManager } from "@/control/keyboard";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";

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

    onKeyPress: function (event: KeyboardEvent): boolean {
      if (AuthController.Locked || !AppStatus.IsPlayerVisible() || !event.key || event.ctrlKey) {
        return false;
      } 
      let caught = true;
      switch (event.key) {
        case "F":
        case "f":
          if (event.altKey || event.shiftKey) {
            caught = false;
          } else {
            this.toggleFullScreen();
          }
          break;
        case "PageDown":
        case "ArrowLeft":
          if (this.prev) {
            this.goPrev();
          } else {
            caught = false;
          }
          break;
        case "PageUp":
        case "ArrowRight":
          if (this.next) {
            this.goNext();
          } else {
            caught = false;
          }
          break;
        default:
          caught = false;
      }

      return caught;
    },
  },
  mounted: function () {
    // Load player preferences

    this.$options.timer = setInterval(
      this.tick.bind(this),
      Math.floor(1000 / 30)
    );

    this.$options.keyHandler = this.onKeyPress.bind(this);
    KeyboardManager.AddHandler(this.$options.keyHandler, 100);

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
    KeyboardManager.RemoveHandler(this.$options.keyHandler);
  },
  watch: {
    rtick: function () {
      this.expandedTitle = false;
    },
  },
});
</script>
