<template>
  <div
    class="player-scale-control"
    :class="{ expanded: expanded, 'player-min': min }"
    :style="{ width: computeFullWidth(width, min, expanded) }"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
  >
    <button class="player-scale-btn" @click="clickOnScaleButton">
      <i class="fas fa-magnifying-glass"></i>
    </button>
    <div
      class="player-scale-btn-expand"
      :class="{ hidden: !expanded }"
      :style="{ width: computeBarContainerWidth(width) }"
      @mousedown="grabScale"
      @touchstart="grabScale"
    >
      <div
        class="player-scale-bar-container"
        :style="{ width: computeBarContainerInnerWidth(width) }"
      >
        <div
          class="player-scale-bar"
          :style="{ width: getScaleBarWidth(width) }"
        ></div>
        <div
          class="player-scale-current"
          :style="{ width: getScaleBarCurrentWidth(width, scale, fit) }"
        ></div>
        <div
          class="player-scale-thumb"
          :style="{ left: getScaleThumbLeft(width, scale, fit) }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";
import { isTouchDevice } from "../../utils/touch";

export default defineComponent({
  name: "ScaleControl",
  emits: ["update:scale", "update:fit", "update:expanded", "enter", "leave"],
  props: {
    width: Number,
    scale: Number,
    min: Boolean,
    fit: Boolean,
    expanded: Boolean,
  },
  setup(props) {
    return {
      scaleState: useVModel(props, "scale"),
      fitState: useVModel(props, "fit"),
      expandedState: useVModel(props, "expanded"),
    };
  },
  data: function () {
    return {
      scaleGrabbed: false,
    };
  },
  methods: {
    onEnter: function () {
      this.$emit("enter");
      this.showScaleBar();
    },
    onLeave: function () {
      this.$emit("leave");
    },
    computeFullWidth: function (
      width: number,
      min: boolean,
      expanded: boolean
    ) {
      let margins = 40;
      let barWidth = width;
      let btnWidth = 40;

      if (min) {
        btnWidth = 24;
      }

      return btnWidth + (expanded ? barWidth + margins : (margins / 2)) + "px";
    },
    computeBarContainerWidth(width: number) {
      let margins = 32;
      return width + margins + "px";
    },
    computeBarContainerInnerWidth(width: number) {
      let margins = 16;
      return width + margins + "px";
    },
    clickOnScaleButton() {
      this.fitState = !this.fitState;
    },
    getScaleBarWidth(width: number) {
      return width + 16 + "px";
    },
    getScaleBarCurrentWidth(width: number, scale: number, fit: boolean) {
      let actualScale = scale;

      if (fit) {
        actualScale = 0;
      }

      actualScale = Math.max(0, Math.min(1, actualScale));

      return Math.floor(actualScale * width) + "px";
    },
    getScaleThumbLeft(width: number, scale: number, fit: boolean) {
      return this.getScaleBarCurrentWidth(width, scale, fit);
    },
    showScaleBar: function () {
      this.expandedState = true;
    },
    hideScaleBar: function () {
      if (isTouchDevice()) {
        return;
      }
      this.expandedState = false;
    },
    grabScale(e) {
      this.scaleGrabbed = true;
      if (e.touches && e.touches.length > 0) {
        this.modifyScaleByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.modifyScaleByMouse(e.pageX, e.pageY);
      }
    },
    dropScale(e) {
      if (!this.scaleGrabbed) {
        return;
      }
      this.scaleGrabbed = false;
      if (e.touches && e.touches.length > 0) {
        this.modifyScaleByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.modifyScaleByMouse(e.pageX, e.pageY);
      }
    },
    moveScale(e) {
      if (!this.scaleGrabbed) {
        return;
      }
      if (e.touches && e.touches.length > 0) {
        this.modifyScaleByMouse(e.touches[0].pageX, e.touches[0].pageY);
      } else {
        this.modifyScaleByMouse(e.pageX, e.pageY);
      }
    },
    modifyScaleByMouse: function (x, y) {
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
        this.changeScale(0);
      } else {
        var p = x - offsetX;
        var vol = Math.min(1, p / this.width);
        this.changeScale(vol);
      }
    },
    changeScale: function (z: number) {
      this.fitState = false;
      this.scaleState = z;
    },
  },
  mounted: function () {
    if (isTouchDevice()) {
      this.expandedState = true;
    }

    this.$options.dropScaleHandler = this.dropScale.bind(this);
    document.addEventListener("mouseup", this.$options.dropScaleHandler);
    document.addEventListener("touchend", this.$options.dropScaleHandler);

    this.$options.moveScaleHandler = this.moveScale.bind(this);

    document.addEventListener("mousemove", this.$options.moveScaleHandler);
    document.addEventListener("touchmove", this.$options.moveScaleHandler);
  },
  beforeUnmount: function () {
    document.removeEventListener("mouseup", this.$options.dropScaleHandler);
    document.removeEventListener("touchend", this.$options.dropScaleHandler);

    document.removeEventListener("mousemove", this.$options.moveScaleHandler);
    document.removeEventListener("touchmove", this.$options.moveScaleHandler);
  },
});
</script>

<style>
.player-scale-control {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  height: 40px;
  overflow: hidden;
  transition: width 0.3s;
}

.player-scale-control.player-min {
  height: 32px;
}

.player-scale-btn {
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
  margin-right: 8px;
}

.player-min .player-scale-btn {
  width: 24px;
  height: 24px;
  font-size: 14px;
}

.player-scale-btn:hover {
  color: white;
}

.player-scale-btn:focus {
  outline: none;
}

.player-scale-btn-expand {
  position: relative;
  display: flex;
  height: 40px;
  align-items: center;
  cursor: pointer;
  overflow: hidden;
  transition: width 0.3s;
}

.player-min .player-scale-btn-expand {
  height: 32px;
}

.player-scale-btn-expand.hidden {
  width: 0 !important;
}

.player-scale-bar-container {
  position: relative;
  overflow: hidden;
  height: 15px;
}

.player-scale-bar {
  background: rgba(255, 255, 255, 0.25);
  height: 3px;
  position: absolute;
  top: 6px;
  left: 0;
}

.player-scale-current {
  background: white;
  height: 3px;
  position: absolute;
  top: 6px;
  left: 0;
}

.player-scale-thumb {
  position: absolute;
  top: 0;
  border-radius: 50%;
  background: white;
  width: 15px;
  height: 15px;
}
</style>

