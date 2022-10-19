<template>
  <div
    class="audio-player-config"
    tabindex="-1"
    :class="{ hidden: !shown }"
    role="dialog"
    :aria-hidden="!shown"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
    @mouseenter="enterConfig"
    @mouseleave="leaveConfig"
    @keydown="keyDownHandle"
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
      <tr>
        <td>
          <i class="fas fa-forward icon-config"></i>
          <b>{{ $t("Auto next") }}</b>
        </td>
        <td class="td-right">
          <ToggleSwitch v-model:val="nextendState"></ToggleSwitch>
        </td>
      </tr>
      <tr
        class="tr-button"
        tabindex="0"
        @click="goToSpeeds"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-gauge icon-config"></i>
          <b>{{ $t("Playback speed") }}</b>
        </td>
        <td class="td-right">
          {{ renderSpeed(speed) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>
      <tr
        class="tr-button"
        tabindex="0"
        @click="goToAnimStyles"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-chart-column icon-config"></i>
          <b>{{ $t("Animation style") }}</b>
        </td>
        <td class="td-right">
          {{ renderAnimStyle(animcolors) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>

      <tr
        v-if="metadata.subtitles && metadata.subtitles.length > 0"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goToSubtitles"
      >
        <td>
          <i class="fas fa-closed-captioning icon-config"></i>
          <b>{{ $t("Subtitles") }}</b>
        </td>
        <td class="td-right">
          {{ renderSubtitle(subtitles, rtick) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>
      <tr
        v-if="metadata.subtitles && metadata.subtitles.length > 0 && subtitles"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goToSubSizes"
      >
        <td>
          <i class="fas fa-closed-captioning icon-config"></i>
          <b>{{ $t("Subtitles") }} ({{ $t("Size") }})</b>
        </td>
        <td class="td-right">
          {{ renderSubtitleSize(subsize) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>

      <tr
        v-if="metadata.subtitles && metadata.subtitles.length > 0 && subtitles"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goToSubBackgrounds"
      >
        <td>
          <i class="fas fa-closed-captioning icon-config"></i>
          <b>{{ $t("Subtitles") }} ({{ $t("Background") }})</b>
        </td>
        <td class="td-right">
          {{ renderSubtitleBackground(subbg) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>

      <tr
        v-if="metadata.subtitles && metadata.subtitles.length > 0 && subtitles"
      >
        <td>
          <i class="fas fa-closed-captioning icon-config"></i>
          <b>{{ $t("Subtitles") }} ({{ $t("Allow HTML") }})</b>
        </td>
        <td class="td-right">
          <ToggleSwitch v-model:val="subhtmlState"></ToggleSwitch>
        </td>
      </tr>
    </table>
    <table v-if="page === 'speed'">
      <tr
        class="tr-button"
        tabindex="0"
        @click="goBack"
        @keydown="clickOnEnter"
      >
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
        @keydown="clickOnEnter"
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
      <tr
        class="tr-button"
        tabindex="0"
        @click="goBack"
        @keydown="clickOnEnter"
      >
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
        @keydown="clickOnEnter"
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

    <table v-if="page === 'subtitles'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
      >
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Subtitles") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeSubtitle('')"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': '' !== subtitles }"
          ></i>
          {{ renderSubtitle("", rtick) }}
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="sub in metadata.subtitles"
        :key="sub.id"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeSubtitle(sub.id)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': sub.id !== subtitles }"
          ></i>
          {{ sub.name }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>

    <table v-if="page === 'subsizes'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
      >
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Subtitles") }} ({{ $t("Size") }}) </b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="s in subtitlesSizes"
        :key="s"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="updateSubtitleSize(s)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': s !== subsize }"
          ></i>
          {{ renderSubtitleSize(s) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>

    <table v-if="page === 'subbg'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
      >
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Subtitles") }} ({{ $t("Background") }}) </b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="s in subtitlesBackgrounds"
        :key="s"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="updateSubtitleBackground(s)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': s !== subbg }"
          ></i>
          {{ renderSubtitleBackground(s) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { SubtitlesController } from "@/control/subtitles";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  components: { ToggleSwitch },
  name: "AudioPlayerConfig",
  emits: [
    "update:shown",
    "update:loop",
    "update:nextend",
    "update:speed",
    "update:animcolors",
    "update:subsize",
    "update:subbg",
    "update:subhtml",
    "enter",
    "leave",
  ],
  props: {
    shown: Boolean,
    metadata: Object,
    loop: Boolean,
    nextend: Boolean,
    speed: Number,
    animcolors: String,
    subsize: String,
    subbg: String,
    subhtml: Boolean,
    rtick: Number,
  },
  setup(props) {
    return {
      shownState: useVModel(props, "shown"),
      loopState: useVModel(props, "loop"),
      nextendState: useVModel(props, "nextend"),
      speedState: useVModel(props, "speed"),
      animColorsState: useVModel(props, "animcolors"),
      subsizeState: useVModel(props, "subsize"),
      subbgState: useVModel(props, "subbg"),
      subhtmlState: useVModel(props, "subhtml"),
    };
  },
  data: function () {
    return {
      page: "",
      speeds: [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2],
      animStyles: ["gradient", "", "none"],

      subtitles: "",

      subtitlesSizes: ["s", "m", "l", "xl", "xxl"],
      subtitlesBackgrounds: ["100", "75", "50", "25", "0"],
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

    focus: function () {
      nextTick(() => {
        this.$el.focus();
      });
    },

    goBack: function () {
      this.page = "";
      this.focus();
    },

    changeSpeed: function (s) {
      this.speedState = s;
    },

    goToSpeeds: function () {
      this.page = "speed";
      this.focus();
    },

    goToAnimStyles: function () {
      this.page = "anim";
      this.focus();
    },

    goToSubtitles: function () {
      this.page = "subtitles";
      this.focus();
    },

    goToSubSizes: function () {
      this.page = "subsizes";
      this.focus();
    },

    goToSubBackgrounds: function () {
      this.page = "subbg";
      this.focus();
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
          return this.$t("Gradient");
        case "none":
          return this.$t("None");
        default:
          return this.$t("Monochrome");
      }
    },

    setAnimStyle: function (s) {
      this.animColorsState = s;
    },

    changeSubtitle: function (s) {
      this.subtitles = s;
      PlayerPreferences.SetSubtitles(s);
      SubtitlesController.OnSubtitlesChanged();
    },

    renderSubtitleSize: function (s: string) {
      switch (s) {
        case "s":
          return this.$t("Small");
        case "l":
          return this.$t("Large");
        case "xl":
          return this.$t("Extra large");
        case "xxl":
          return this.$t("Extra extra large");
        default:
          return this.$t("Medium");
      }
    },

    updateSubtitleSize: function (s: string) {
      this.subsizeState = s;
      PlayerPreferences.SetSubtitlesSize(s);
    },

    renderSubtitleBackground: function (s: string) {
      switch (s) {
        case "0":
          return this.$t("Transparent");
        case "25":
          return this.$t("Translucid") + " (75%)";
        case "50":
          return this.$t("Translucid") + " (50%)";
        case "75":
          return this.$t("Translucid") + " (25%)";
        default:
          return this.$t("Opaque");
      }
    },

    updateSubtitleBackground: function (s: string) {
      this.subbgState = s;
      PlayerPreferences.SetSubtitlesBackground(s);
    },

    renderSubtitle: function (subId: string, rtick: number) {
      if (rtick < 0 || !this.metadata || !this.metadata.subtitles || !subId) {
        return this.$t("No subtitles");
      }

      for (let sub of this.metadata.subtitles) {
        if (sub.id === subId) {
          return sub.name;
        }
      }

      return this.$t("No subtitles");
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },

    close: function () {
      this.shownState = false;
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.subtitles = PlayerPreferences.SelectedSubtitles;
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
  },
  beforeUnmount: function () {
    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    shown: function () {
      this.page = "";
      if (this.shown) {
        this.$options.focusTrap.activate();
        nextTick(() => {
          this.$el.focus();
        });
      } else {
        this.$options.focusTrap.deactivate();
      }
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