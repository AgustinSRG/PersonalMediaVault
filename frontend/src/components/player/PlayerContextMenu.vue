<template>
  <div
    class="player-context-menu"
    :class="{
      hidden: !shown,
    }"
    :style="{
      top: top,
      left: left,
      right: right,
      bottom: bottom,
      width: width,
      'max-width': maxWidth,
      'max-height': maxHeight,
    }"
    tabindex="-1"
    @mousedown="stopPropagationEvent"
    @touchstart="stopPropagationEvent"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
  >
    <table class="player-context-menu-table">
      <tr
        v-if="type === 'video' || type === 'audio'"
        class="tr-button"
        tabindex="0"
        @click="toggleLoop"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-repeat icon-config"></i>
          <span class="context-entry-title">{{ $t("Loop") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': !loop }"></i>
        </td>
      </tr>
      <tr
        v-if="type === 'image'"
        class="tr-button"
        tabindex="0"
        @click="toggleFit"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-magnifying-glass icon-config"></i>
          <span class="context-entry-title">{{ $t("Fit image") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': !fit }"></i>
        </td>
      </tr>

      <tr
        v-if="type === 'image'"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="toggleControls"
      >
        <td>
          <i class="fas fa-eye-slash icon-config"></i>
          <span class="context-entry-title">{{ $t("Hide controls") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': controls }"></i>
        </td>
      </tr>

      <tr v-if="url" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="download">
        <td>
          <i class="fas fa-download icon-config"></i>
          <span class="context-entry-title">{{ $t("Download") }}</span>
        </td>
        <td class="td-right"></td>
      </tr>

      <tr v-if="url" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="refreshMedia">
        <td>
          <i class="fas fa-sync-alt icon-config"></i>
          <span class="context-entry-title">{{ $t("Refresh") }}</span>
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
  </div>
</template>

<script lang="ts">
import { MediaController } from "@/control/media";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "PlayerContextMenu",
  emits: ["update:shown", "update:loop", "update:controls", "update:fit"],
  props: {
    shown: Boolean,
    type: String,
    x: Number,
    y: Number,

    url: String,

    loop: Boolean,
    fit: Boolean,
    controls: Boolean,
  },
  setup(props) {
    return {
      shownState: useVModel(props, "shown"),
      loopState: useVModel(props, "loop"),
      fitState: useVModel(props, "fit"),
      controlsState: useVModel(props, "controls"),
    };
  },
  data: function () {
    return {
      top: "",
      left: "",
      right: "",
      bottom: "",

      width: "",

      maxWidth: "",
      maxHeight: "",
    };
  },
  methods: {
    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    toggleLoop: function () {
      this.loopState = !this.loopState;
      this.shownState = false;
    },

    toggleFit: function () {
      this.fitState = !this.fitState;
      this.shownState = false;
    },

    toggleControls: function () {
      this.controlsState = !this.controlsState;
      this.shownState = false;
    },

    refreshMedia: function () {
      MediaController.Load();
      this.shownState = false;
    },

    download: function () {
      this.shownState = false;
      const link = document.createElement("a");
      link.target = "_blank";
      link.rel = "noopener noreferrer";
      link.href = this.url;
      link.click();
    },

    hide: function () {
      this.shownState = false;
    },

    computeDimensions: function () {
      const pageWidth = window.innerWidth;
      const pageHeight = window.innerHeight;

      const x = this.x;
      const y = this.y;

      let top = y;
      let left = x;

      let maxWidth = pageWidth - left;

      let maxHeight = pageHeight - top;

      this.top = top + "px";
      this.left = left + "px";
      this.right = "auto";
      this.bottom = "auto";
      this.width = "auto";
      this.maxWidth = maxWidth + "px";
      this.maxHeight = maxHeight + "px";
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },
  },
  mounted: function () {
    this.computeDimensions();

    this.$options.hideHandler = this.hide.bind(this);

    document.addEventListener("mousedown", this.$options.hideHandler);
    document.addEventListener("touchstart", this.$options.hideHandler);
  },
  beforeUnmount: function () {
    document.removeEventListener("mousedown", this.$options.hideHandler);
    document.removeEventListener("touchstart", this.$options.hideHandler);
  },
  watch: {
    x: function () {
      this.computeDimensions();
    },
    y: function () {
      this.computeDimensions();
    },
    shown: function () {
      if (this.shown) {
        nextTick(() => {
          this.$el.focus();
        });
      }
    },
  },
});
</script>

<style>
.player-context-menu {
  position: fixed;
  background: rgba(0, 0, 0, 0.8);
  overflow: auto;
  z-index: 10;
  min-width: 380px;
  padding-top: 0.3rem;
  padding-bottom: 0.3rem;
  opacity: 1;
  transition: opacity 0.1s;
}

.player-min .player-context-menu  {
  min-width: fit-content;
}

.player-context-menu.hidden {
  transition: opacity 0.1s, visibility 0.1s;
  opacity: 0;
  pointer-events: none;
  visibility: hidden;
}

.full-screen .player-context-menu {
  z-index: 110;
}

.player-context-menu-table {
  width: 100%;
  border-spacing: 0; /* Removes the cell spacing via CSS */
  border-collapse: collapse; /* Optional - if you don't want to have double border where cells touch */
}

.player-context-menu-table td {
  padding: 1rem 0.75rem;
  text-align: left;
  vertical-align: middle;
  white-space: nowrap;
}

.player-context-menu-table .tr-button {
  cursor: pointer;
}

.player-context-menu-table .tr-button:hover {
  background: rgba(255, 255, 255, 0.1);
}

.player-context-menu-table .td-right {
  text-align: right;
}

.player-context-menu-table .check-uncheck {
  visibility: hidden;
}

.player-context-menu-table .icon-config {
  width: 24px;
  margin-right: 0.5rem;
}

/* Custom scroll bar */

/* width */

.player-context-menu::-webkit-scrollbar {
  width: 5px;
  height: 3px;
}

/* Track */

.player-context-menu::-webkit-scrollbar-track {
  background: #bdbdbd;
}

/* Handle */

.player-context-menu::-webkit-scrollbar-thumb {
  background: #757575;
}

.context-entry-title {
  font-size: small;
}

</style>
