<template>
  <div
    class="image-player-config"
    :class="{ hidden: !shown }"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
    @mouseenter="enterConfig"
    @mouseleave="leaveConfig"
  >
    <table v-if="page === ''">
      <tr class="tr-button" tabindex="0" @click="goToResolutions">
        <td>
          <i class="fas fa-photo-film icon-config"></i>
          <b>{{ $t("Quality") }}</b>
        </td>
        <td class="td-right">
          {{ renderResolution(resolution, rtick) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>

      <tr class="tr-button" tabindex="0" @click="goToBackgrounds">
        <td>
          <i class="fas fa-palette icon-config"></i>
          <b>{{ $t("Background") }}</b>
        </td>
        <td class="td-right">
          {{ renderBackground(background) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>

      <tr class="tr-button" tabindex="0" @click="goToAutonext">
        <td>
          <i class="fas fa-clock-rotate-left icon-config"></i>
          <b>{{ $t("Auto next") }}</b>
        </td>
        <td class="td-right">
          {{ renderAutoNext(autoNext) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>

    </table>
    <table v-if="page === 'resolution'">
      <tr class="tr-button" tabindex="0" @click="goBack">
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Quality") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr class="tr-button" tabindex="0" @click="changeResolution(-1)">
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': -1 !== resolution }"
          ></i>
          {{ renderResolution(-1, rtick) }}
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="(r, i) in metadata.resolutions"
        :key="i"
        class="tr-button"
        tabindex="0"
        @click="changeResolution(i)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': i !== resolution }"
          ></i>
          {{ renderResolution(i, rtick) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
    <table v-if="page === 'background'">
      <tr class="tr-button" tabindex="0" @click="goBack">
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Background") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr v-for="b in bgOptions" :key="b" class="tr-button" tabindex="0" @click="changeBackground(b)">
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': b !== background }"
          ></i>
          {{ renderBackground(b) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
     <table v-if="page === 'autotext'">
      <tr class="tr-button" tabindex="0" @click="goBack">
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Auto next") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr v-for="b in autoNextOptions" :key="b" class="tr-button" tabindex="0" @click="changeAutoNext(b)">
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': b !== autoNext }"
          ></i>
          {{ renderAutoNext(b) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
  </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "ImagePlayerConfig",
  emits: [
    "update:shown",
    "update:resolution",
    "update:background",
    "update-autonext",
    "enter",
    "leave",
  ],
  props: {
    shown: Boolean,
    metadata: Object,
    resolution: Number,
    background: String,
    rtick: Number,
  },
  setup(props) {
    return {
      shownState: useVModel(props, "shown"),
      resolutionState: useVModel(props, "resolution"),
      backgroundState: useVModel(props, "background"),
    };
  },
  data: function () {
    return {
      page: "",
      resolutions: [],
      bgOptions: ['default', 'black', 'white'],
      autoNext: PlayerPreferences.ImageAutoNext,
      autoNextOptions: [0, 3, 5, 10, 15, 20, 25, 30],
    };
  },
  methods: {
    changeResolution: function (i) {
      this.resolutionState = i;
    },

    changeBackground: function (b) {
      this.backgroundState = b;
    },

    changeAutoNext: function (b) {
      this.autoNext = b;
      PlayerPreferences.SetImageAutoNext(b);
      this.$emit("update-autonext");
    },

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

    goToResolutions: function () {
      this.page = "resolution";
    },

    goToBackgrounds: function () {
      this.page = "background";
    },

    goToAutonext: function () {
      this.page = "autotext";
    },

    renderBackground: function (b: string) {
      switch (b) {
        case "white":
          return this.$t("White");
        case "black":
          return this.$t("Black");
        default:
          return this.$t("Default (Theme)");
      }
    },

    renderAutoNext: function (s: number) {
      if (!isNaN(s) && isFinite(s) && s > 0) {
        if (s === 1) {
          return s + " " + this.$t('second');
        } else {
          return s + " " + this.$t('seconds');
        }
      } else {
        return this.$t('Disabled');
      }
    },

    renderResolution: function (res: number, rtick: number) {
      if (rtick < 0 || !this.metadata) {
        return this.$t("Unknown");
      }
      if (res < 0) {
        return (
          this.metadata.width +
          "x" +
          this.metadata.height +
          " (" +
          this.$t("Original") +
          ")" +
          (this.metadata.encoded ? "" : "(" + this.$t("Pending") + ")")
        );
      } else {
        let resData = this.metadata.resolutions[res];
        if (resData) {
          return (
            resData.width +
            "x" +
            resData.height +
            "" +
            (resData.ready ? "" : "(" + this.$t("Pending") + ")")
          );
        } else {
          return this.$t("Unknown");
        }
      }
    },
    updateResolutions: function () {
      if (this.metadata && this.metadata.resolutions) {
        this.resolutions = this.metadata.resolutions.slice();
      } else {
        this.resolutions = [];
      }
    },
  },
  mounted: function () {
    this.updateResolutions();
  },
  beforeUnmount: function () {},
  watch: {
    shown: function () {
      this.page = "";
    },
    rtick: function () {
      this.updateResolutions();
    },
  },
});
</script>

<style>
.image-player-config {
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

.player-min .image-player-config {
  font-size: small;
  bottom: 55px;
}

.image-player-config.hidden {
  display: none;
}

.image-player-config table {
  width: 100%;
  border-spacing: 0; /* Removes the cell spacing via CSS */
  border-collapse: collapse; /* Optional - if you don't want to have double border where cells touch */
}

.image-player-config table td {
  padding: 0.75rem 1rem;
  text-align: left;
  vertical-align: middle;
}

.image-player-config .tr-button {
  cursor: pointer;
}

.image-player-config .arrow-config {
  margin-left: 0.5rem;
}

.image-player-config .icon-config {
  margin-right: 0.5rem;
  width: 24px;
}

.image-player-config .tr-button:hover {
  background: rgba(255, 255, 255, 0.1);
}

.image-player-config .td-right {
  text-align: right;
}

.image-player-config .check-uncheck {
  visibility: hidden;
}

/* Custom scroll bar */

/* width */

.image-player-config::-webkit-scrollbar {
  width: 5px;
  height: 3px;
}

/* Track */

.image-player-config::-webkit-scrollbar-track {
  background: #bdbdbd;
}

/* Handle */

.image-player-config::-webkit-scrollbar-thumb {
  background: #757575;
}
</style>