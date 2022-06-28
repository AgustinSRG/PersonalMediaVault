<template>
  <div
    class="album-body-item-options-menu"
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
    @mousedown="stopPropagationEvent"
    @touchstart="stopPropagationEvent"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
  >
    <div
      v-if="mindex > 0"
      tabindex="0"
      @click="moveMediaUp"
      class="album-body-item-options-menu-btn"
    >
      <i class="fas fa-arrow-up"></i> {{ $t("Move up") }}
    </div>
    <div
      v-if="mindex < mlength - 1"
      tabindex="0"
      @click="moveMediaDown"
      class="album-body-item-options-menu-btn"
    >
      <i class="fas fa-arrow-down"></i> {{ $t("Move down") }}
    </div>
    <div
      tabindex="0"
      @click="removeMedia"
      class="album-body-item-options-menu-btn"
    >
      <i class="fas fa-trash-alt"></i> {{ $t("Remove from the album") }}
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "AlbumContextMenu",
  emits: ["update:shown", "move-up", "move-down", "media-remove"],
  props: {
    shown: Boolean,

    mindex: Number,
    mlength: Number,

    x: Number,
    y: Number,
  },
  setup(props) {
    return {
      shownState: useVModel(props, "shown"),
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

    moveMediaUp: function () {
      this.$emit("move-up");
      this.hide();
    },

    moveMediaDown: function () {
      this.$emit("move-up");
      this.hide();
    },

    removeMedia: function () {
      this.$emit("media-remove");
      this.hide();
    },

    hide: function () {
      this.shownState = false;
    },

    computeDimensions: function () {
      const pageWidth = window.innerWidth;
      const pageHeight = window.innerHeight;

      const x = this.x;
      const y = this.y;

      if (y > pageHeight / 2) {
        let bottom = pageHeight - y;
        let right = pageWidth - x;

        let maxWidth = pageWidth - right;

        let maxHeight = pageHeight - bottom;

        this.top = "auto";
        this.left = "auto";
        this.right = right + "px";
        this.bottom = bottom + "px";
        this.width = "auto";
        this.maxWidth = maxWidth + "px";
        this.maxHeight = maxHeight + "px";
      } else {
        let top = y;
        let right = pageWidth - x;

        let maxWidth = pageWidth - right;

        let maxHeight = pageHeight - top;

        this.top = top + "px";
        this.left = "auto";
        this.right = right + "px";
        this.bottom = "auto";
        this.width = "auto";
        this.maxWidth = maxWidth + "px";
        this.maxHeight = maxHeight + "px";
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
  },
});
</script>

<style>
.album-body-item-options-menu {
  position: fixed;
  display: flex;
  flex-direction: column;
  background: rgba(0, 0, 0, 0.9);
  padding: 0.25rem 0;
  z-index: 110;
}

.album-body-item-options-menu.hidden {
  display: none;
}

.album-body-item-options-menu-btn {
  cursor: pointer;
  padding: 1rem 1rem;
  white-space: nowrap;
}

.album-body-item-options-menu-btn i {
  width: 24px;
  margin-right: 0.5rem;
}

.album-body-item-options-menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
