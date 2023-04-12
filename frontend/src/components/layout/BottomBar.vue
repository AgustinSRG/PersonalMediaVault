<template>
  <div class="bottom-bar">
    <div
      class="bottom-bar-option bottom-bar-option-list"
      tabindex="0"
      @click="goPrev"
      @keydown="clickOnEnter"
    >
      <i class="fas fa-backward-step"></i><span> {{ $t("Previous") }}</span>
    </div>
    <div
      class="bottom-bar-option bottom-bar-option-media"
      :class="{ selected: focus === 'left' }"
      tabindex="0"
      @click="clickLeft"
      @keydown="clickOnEnter"
    >
      <i class="fas fa-photo-film"></i><span> {{ $t("Media") }}</span>
    </div>
    <div
      class="bottom-bar-option bottom-bar-option-list"
      :class="{ selected: focus === 'right' }"
      tabindex="0"
      @click="clickRight"
      @keydown="clickOnEnter"
    >
      <i class="fas fa-list"></i><span> {{ $t("List") }}</span>
    </div>
    <div
      class="bottom-bar-option bottom-bar-option-list"
      tabindex="0"
      @click="goNext"
      @keydown="clickOnEnter"
    >
      <i class="fas fa-forward-step"></i><span> {{ $t("Next") }}</span>
    </div>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";

export default defineComponent({
  name: "BottomBar",
  data: function () {
    return {
      focus: AppStatus.CurrentFocus,
    };
  },
  methods: {
    onStatusUpdate: function () {
      this.focus = AppStatus.CurrentFocus;
    },

    clickLeft: function () {
      AppStatus.FocusLeft();
    },

    clickRight: function () {
      AppStatus.FocusRight();
    },

    goNext: function () {
      AppEvents.Emit("media-go-next");
    },

    goPrev: function () {
      AppEvents.Emit("media-go-prev");
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
    this.$options.updateStatusH = this.onStatusUpdate.bind(this);
    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.updateStatusH
    );
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.updateStatusH
    );
  },
});
</script>
