<template>
  <div
    class="player-volume-control"
    :class="{ expanded: expanded, 'player-min': min }"
    :style="{ width: computeFullWidth(width, min, expanded) }"
    @mouseenter="showVolumeBar"
  >
    <button class="player-volume-btn" @click="clickOnVolumeButton">
      <i v-if="!muted && volume > 0.5" class="fas fa-volume-up"></i>
      <i v-if="!muted && volume <= 0" class="fas fa-volume-off"></i>
      <i
        v-if="!muted && volume > 0 && volume <= 0.5"
        class="fas fa-volume-down"
      ></i>
      <i v-if="muted" class="fas fa-volume-mute"></i>
    </button>
    <div
      class="player-volume-btn-expand"
      :class="{ hidden: !expanded }"
      :style="{ width: computeBarContainerWidth(width) }"
      @mousedown="grabVolume"
      @touchstart="grabVolume"
    >
      <div
        class="player-volume-bar-container"
        :style="{ width: computeBarContainerInnerWidth(width) }"
      >
        <div
          class="player-volume-bar"
          :style="{ width: getVolumeBarWidth(width) }"
        ></div>
        <div
          class="player-volume-current"
          :style="{ width: getVolumeBarCurrentWidth(width, volume, muted) }"
        ></div>
        <div
          class="player-volume-thumb"
          :style="{ left: getVolumeThumbLeft(width, volume, muted) }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "VolumeControl",
  emits: ["update:volume", "update:muted"],
  props: {
    width: Number,
    volume: Number,
    min: Boolean,
    muted: Boolean,
  },
  setup(props) {
    return {
      volumeState: useVModel(props, "volume"),
      mutedState: useVModel(props, "muted"),
    };
  },
  data: function () {
    return {
      expanded: false,
      volumeGrabbed: false,
    };
  },
  methods: {
    computeFullWidth: function (
      width: number,
      min: boolean,
      expanded: boolean
    ) {
      let margins = 24;
      let barWidth = width;
      let btnWidth = 40;

      if (min) {
        btnWidth = 24;
      }

      return btnWidth + (expanded ? barWidth + margins : 0) + "px";
    },
    computeBarContainerWidth(width: number) {
      let margins = 24;
      return width + margins + "px";
    },
    computeBarContainerInnerWidth(width: number) {
      let margins = 16;
      return width + margins + "px";
    },
    clickOnVolumeButton() {
      this.mutedState = !this.mutedState;
    },
    getVolumeBarWidth(width: number) {
      return width + "px";
    },
    getVolumeBarCurrentWidth(width: number, volume: number, muted: boolean) {
      let actualVolume = volume;

      if (muted) {
        actualVolume = 0;
      }

      actualVolume = Math.max(0, Math.min(1, actualVolume));

      return Math.floor(actualVolume * width) + "px";
    },
    getVolumeThumbLeft(width: number, volume: number, muted: boolean) {
      return this.getVolumeBarCurrentWidth(width, volume, muted);
    },
    showVolumeBar: function () {
      this.expanded = true;
    },
    hideVolumeBar: function () {
      this.expanded = false;
    },
    grabVolume(e) {
      this.volumeGrabbed = true;
      if (e.touches && e.touches.length > 0) {
        this.modifyVolumeByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.modifyVolumeByMouse(e.pageX, e.pageY);
      }
    },
    dropVolume(e) {
      if (!this.volumeGrabbed) {
        return;
      }
      this.volumeGrabbed = false;
      if (e.touches && e.touches.length > 0) {
        this.modifyVolumeByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.modifyVolumeByMouse(e.pageX, e.pageY);
      }
    },
    moveVolume(e) {
      if (!this.volumeGrabbed) {
        return;
      }
      if (e.touches && e.touches.length > 0) {
        this.modifyVolumeByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.modifyVolumeByMouse(e.pageX, e.pageY);
      }
    },
    modifyVolumeByMouse: function (x, y) {
      if (
        typeof x !== "number" ||
        typeof y !== "number" ||
        isNaN(x) ||
        isNaN(y)
      ) {
        return;
      }
      var offset = this.$el.getBoundingClientRect();

      var offsetX = offset.left + 8 + (this.min ? 24 : 40);

      if (x < offsetX) {
        this.changeVolume(0);
      } else {
        var p = x - offsetX;
        var vol = Math.min(1, p / this.width);
        this.changeVolume(vol);
      }
    },
    changeVolume: function (v: number) {
      this.mutedState = false;
      this.volumeState = v;
    },
  },
  mounted: function () {
    this.$options.dropVolumeHandler = this.dropVolume.bind(this);
    document.addEventListener("mouseup", this.$options.dropVolumeHandler);
    document.addEventListener("touchend", this.$options.dropVolumeHandler);

    this.$options.moveVolumeHandler = this.moveVolume.bind(this);

    document.addEventListener("mousemove", this.$options.moveVolumeHandler);
    document.addEventListener("touchmove", this.$options.moveVolumeHandler);
  },
  beforeUnmount: function () {
    document.removeEventListener("mouseup", this.$options.dropVolumeHandler);
    document.removeEventListener("touchend", this.$options.dropVolumeHandler);

    document.removeEventListener("mousemove", this.$options.moveVolumeHandler);
    document.removeEventListener("touchmove", this.$options.moveVolumeHandler);
  },
});
</script>

<i18n>
{
  "en": {
    "hello": "Hello i18n in SFC!"
  }
}
</i18n>

<style>
.player-volume-control {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  height: 40px;
  overflow: hidden;
  transition: width 0.5s;
}

.player-volume-control.player-min {
  height: 32px;
}

.player-volume-btn {
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

.player-min .player-volume-btn {
  width: 24px;
  height: 24px;
  font-size: 14px;
}

.player-volume-btn:hover {
  color: white;
}

.player-volume-btn:focus {
  outline: none;
}

.player-volume-btn-expand {
  position: relative;
  display: flex;
  height: 40px;
  width: 96px;
  align-items: center;
  cursor: pointer;
  padding-left: 8px;
  overflow: hidden;
  transition: width 0.5s;
}

.player-min .player-volume-btn-expand {
  height: 32px;
}

.player-volume-btn-expand.hidden {
  width: 0 !important;
}

.player-volume-bar-container {
  position: relative;
  overflow: hidden;
  width: 96px;
  height: 15px;
}

.player-volume-bar {
  background: rgba(255, 255, 255, 0.25);
  width: 80px;
  height: 3px;
  position: absolute;
  top: 6px;
  left: 0;
}

.player-volume-current {
  background: white;
  width: 40px;
  height: 3px;
  position: absolute;
  top: 6px;
  left: 0;
}

.player-volume-thumb {
  position: absolute;
  top: 0;
  left: 40px;
  border-radius: 50%;
  background: white;
  width: 15px;
  height: 15px;
}
</style>


