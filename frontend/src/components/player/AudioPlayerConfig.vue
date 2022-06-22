<template>
  <div
    class="audio-player-config"
    :class="{ hidden: !shown }"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
    @mouseenter="enterConfig"
    @mouseleave="leaveConfig"
  >
    <table v-if="page === ''">
      <tr>
        <td>
          <i class="fas fa-repeat icon-config"></i>
          <b>{{ $t("Loop") }}</b>
        </td>
        <td class="td-right">
          <ToggleSwitch v-model:val="loopState"></ToggleSwitch>
        </td>
      </tr>
      <tr class="tr-button" tabindex="0" @click="goToSpeeds">
        <td>
          <i class="fas fa-gauge icon-config"></i>
          <b>{{ $t("Playback speed") }}</b>
        </td>
        <td class="td-right">
          {{ renderSpeed(speed) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>
      <tr class="tr-button" tabindex="0" @click="goToAnimStyles">
        <td>
          <i class="fas fa-chart-column icon-config"></i>
          <b>{{ $t("Animation style") }}</b>
        </td>
        <td class="td-right">
          {{ renderAnimStyle(animcolors) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>
    </table>
    <table v-if="page === 'speed'">
      <tr class="tr-button" tabindex="0" @click="goBack">
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Playback speed") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="s in speeds"
        :key="s"
        class="tr-button"
        tabindex="0"
        @click="changeSpeed(s)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': s !== speed }"
          ></i>
          {{ renderSpeed(s) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
    <table v-if="page === 'anim'">
      <tr class="tr-button" tabindex="0" @click="goBack">
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Animation style") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="s in animStyles"
        :key="s"
        class="tr-button"
        tabindex="0"
        @click="setAnimStyle(s)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': s !== animcolors }"
          ></i>
          {{ renderAnimStyle(s) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";
import ToggleSwitch from "../utils/ToggleSwitch.vue";

export default defineComponent({
  components: { ToggleSwitch },
  name: "AudioPlayerConfig",
  emits: [
    "update:shown",
    "update:loop",
    "update:speed",
    "update:animcolors",
    "enter",
    "leave",
  ],
  props: {
    shown: Boolean,
    metadata: Object,
    loop: Boolean,
    speed: Number,
    animcolors: String,
    rtick: Number,
  },
  setup(props) {
    return {
      shownState: useVModel(props, "shown"),
      loopState: useVModel(props, "loop"),
      speedState: useVModel(props, "speed"),
      animColorsState: useVModel(props, "animcolors"),
    };
  },
  data: function () {
    return {
      page: "",
      speeds: [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2],
      animStyles: ["gradient", "", "none"],
    };
  },
  methods: {
    enterConfig: function () {
      this.$emit("enter");
    },

    leaveConfig: function () {
      this.$emit("leave");
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    goBack: function () {
      this.page = "";
    },

    changeSpeed: function (s) {
      this.speedState = s;
    },

    goToSpeeds: function () {
      this.page = "speed";
    },

    goToAnimStyles: function () {
      this.page = "anim";
    },

    renderSpeed: function (speed: number) {
      if (speed > 1) {
        return Math.floor(speed * 100) + "%";
      } else if (speed < 1) {
        return Math.floor(speed * 100) + "%";
      } else {
        return this.$t("Normal");
      }
    },

    renderAnimStyle: function (s) {
      switch (s) {
        case "gradient":
          return this.$t('Gradient');
        case "none":
          return this.$t('None');
        default:
          return this.$t('Monochrome');
      }
    },

    setAnimStyle: function (s) {
      this.animColorsState = s;
    },
  },
  mounted: function () {
  },
  beforeUnmount: function () {},
  watch: {
    shown: function () {
      this.page = "";
    },
  },
});
</script>

<style>
.audio-player-config {
  position: absolute;
  bottom: 80px;
  right: 8px;
  background: rgba(0, 0, 0, 0.8);
  padding-top: 8px;
  padding-bottom: 8px;
  max-height: calc(50% - 80px);
  min-width: 240px;
  max-width: calc(100% - 16px);
  overflow-y: auto;
}

.player-min .audio-player-config {
  font-size: small;
  bottom: 55px;
}

.audio-player-config.hidden {
  display: none;
}

.audio-player-config table {
  width: 100%;
  border-spacing: 0; /* Removes the cell spacing via CSS */
  border-collapse: collapse; /* Optional - if you don't want to have double border where cells touch */
}

.audio-player-config table td {
  padding: 0.75rem 1rem;
  text-align: left;
  vertical-align: middle;
}

.audio-player-config .tr-button {
  cursor: pointer;
}

.audio-player-config .arrow-config {
  margin-left: 0.5rem;
}

.audio-player-config .icon-config {
  margin-right: 0.5rem;
  width: 24px;
}

.audio-player-config .tr-button:hover {
  background: rgba(255, 255, 255, 0.1);
}

.audio-player-config .td-right {
  text-align: right;
}

.audio-player-config .check-uncheck {
  visibility: hidden;
}

/* Custom scroll bar */


/* width */

.audio-player-config::-webkit-scrollbar {
    width: 5px;
    height: 3px;
}


/* Track */

.audio-player-config::-webkit-scrollbar-track {
    background: #bdbdbd;
}


/* Handle */

.audio-player-config::-webkit-scrollbar-thumb {
    background: #757575;
}
</style>