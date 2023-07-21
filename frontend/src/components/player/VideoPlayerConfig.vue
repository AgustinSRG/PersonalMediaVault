<template>
  <div
    class="video-player-config"
    :class="{ hidden: !shown }"
    tabindex="-1"
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
          <ToggleSwitch v-model:val="nextEndState"></ToggleSwitch>
        </td>
      </tr>
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goToSpeeds"
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
        @keydown="clickOnEnter"
        @click="goToResolutions"
      >
        <td>
          <i class="fas fa-photo-film icon-config"></i>
          <b>{{ $t("Quality") }}</b>
        </td>
        <td class="td-right">
          {{ renderResolution(resolution, rTick) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>
      <tr
        v-if="metadata.audios && metadata.audios.length > 0"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goToAudios"
      >
        <td>
          <i class="fas fa-headphones icon-config"></i>
          <b>{{ $t("Audio") }}</b>
        </td>
        <td class="td-right">
          {{ renderAudio(audioTrack, rTick) }}
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
          {{ renderSubtitle(subtitles, rTick) }}
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
          {{ renderSubtitleSize(subSize) }}
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
          {{ renderSubtitleBackground(subBackground) }}
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
          <ToggleSwitch v-model:val="subHTMLState"></ToggleSwitch>
        </td>
      </tr>

      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goToDelays"
      >
        <td>
          <i class="fas fa-clock icon-config"></i>
          <b>{{ $t("Toggle play delay") }}</b>
        </td>
        <td class="td-right">
          {{ renderToggleDelay(toggleDelay) }}
          <i class="fas fa-chevron-right arrow-config"></i>
        </td>
      </tr>
    </table>
    <table v-if="page === 'speed'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
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
        @keydown="clickOnEnter"
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
    <table v-if="page === 'resolution'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
      >
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Quality") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeResolution(-1)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': -1 !== resolution }"
          ></i>
          {{ renderResolution(-1, rTick) }}
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="(r, i) in metadata.resolutions"
        :key="i"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeResolution(i)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': i !== resolution }"
          ></i>
          {{ renderResolution(i, rTick) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>

    <table v-if="page === 'audios'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
      >
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Audio") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeAudioTrack('')"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': '' !== audioTrack }"
          ></i>
          {{ renderAudio("", rTick) }}
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="aud in metadata.audios"
        :key="aud.id"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeAudioTrack(aud.id)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': aud.id !== audioTrack }"
          ></i>
          {{ aud.name }}
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
          {{ renderSubtitle("", rTick) }}
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

    <table v-if="page === 'subSizes'">
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
            :class="{ 'check-uncheck': s !== subSize }"
          ></i>
          {{ renderSubtitleSize(s) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>

    <table v-if="page === 'subBackground'">
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
            :class="{ 'check-uncheck': s !== subBackground }"
          ></i>
          {{ renderSubtitleBackground(s) }}
        </td>
        <td class="td-right"></td>
      </tr>
    </table>

    <table v-if="page === 'time-delays'">
      <tr
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="goBack"
      >
        <td>
          <i class="fas fa-chevron-left icon-config"></i>
          <b>{{ $t("Toggle play delay") }}</b>
        </td>
        <td class="td-right"></td>
      </tr>
      <tr
        v-for="s in toggleDelayOptions"
        :key="s"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="changeToggleDelay(s)"
      >
        <td>
          <i
            class="fas fa-check icon-config"
            :class="{ 'check-uncheck': s !== toggleDelay }"
          ></i>
          {{ renderToggleDelay(s) }}
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
import { useVModel } from "../../utils/v-model";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    components: { ToggleSwitch },
    name: "VideoPlayerConfig",
    emits: [
        "update:shown",
        "update:loop",
        "update:nextEnd",
        "update:speed",
        "update:resolution",
        "update:subSize",
        "update:subBackground",
        "update:subHTML",
        "update:audioTrack",
        "enter",
        "leave",
    ],
    props: {
        shown: Boolean,
        metadata: Object,
        loop: Boolean,
        nextEnd: Boolean,
        speed: Number,
        resolution: Number,
        subSize: String,
        subBackground: String,
        subHTML: Boolean,
        rTick: Number,
        audioTrack: String,
    },
    setup(props) {
        return {
            shownState: useVModel(props, "shown"),
            loopState: useVModel(props, "loop"),
            nextEndState: useVModel(props, "nextEnd"),
            speedState: useVModel(props, "speed"),
            resolutionState: useVModel(props, "resolution"),
            subSizeState: useVModel(props, "subSize"),
            subBackgroundState: useVModel(props, "subBackground"),
            subHTMLState: useVModel(props, "subHTML"),
            audioTrackState: useVModel(props, "audioTrack"),
        };
    },
    data: function () {
        return {
            page: "",
            speeds: [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2],
            resolutions: [],

            subtitles: "",
            subtitlesSizes: ["s", "m", "l", "xl", "xxl"],
            subtitlesBackgrounds: ["100", "75", "50", "25", "0"],

            toggleDelay: PlayerPreferences.PlayerTogglePlayDelay,
            toggleDelayOptions: [0, 250, 500, 750, 1000],
        };
    },
    methods: {
        changeResolution: function (i) {
            this.resolutionState = i;
        },

        changeToggleDelay: function (d) {
            this.toggleDelay = d;
            PlayerPreferences.SetPlayerToggleDelay(d);
        },

        changeSubtitle: function (s) {
            this.subtitles = s;
            PlayerPreferences.SetSubtitles(s);
            SubtitlesController.OnSubtitlesChanged();
        },

        changeAudioTrack: function (s) {
            this.audioTrackState = s;
            PlayerPreferences.SetAudioTrack(s);
        },

        focus: function () {
            nextTick(() => {
                this.$el.focus();
            });
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
            this.focus();
        },

        changeSpeed: function (s) {
            this.speedState = s;
        },

        goToSpeeds: function () {
            this.page = "speed";
            this.focus();
        },

        goToResolutions: function () {
            this.page = "resolution";
            this.focus();
        },

        goToSubtitles: function () {
            this.page = "subtitles";
            this.focus();
        },

        goToAudios: function () {
            this.page = "audios";
            this.focus();
        },

        goToSubSizes: function () {
            this.page = "subSizes";
            this.focus();
        },

        goToSubBackgrounds: function () {
            this.page = "subBackground";
            this.focus();
        },

        goToDelays: function () {
            this.page = "time-delays";
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
        renderResolution: function (res: number, rTick: number) {
            if (rTick < 0 || !this.metadata) {
                return this.$t("Unknown");
            }
            if (res < 0) {
                return (
                    this.metadata.width +
          "x" +
          this.metadata.height +
          ", " +
          this.metadata.fps +
          " fps (" +
          this.$t("Original") +
          ")" +
          (this.metadata.encoded ? "" : " (" + this.$t("Pending") + ")")
                );
            } else {
                let resData = this.metadata.resolutions[res];
                if (resData) {
                    return (
                        resData.width +
            "x" +
            resData.height +
            ", " +
            resData.fps +
            " fps " +
            (resData.ready ? "" : " (" + this.$t("Pending") + ")")
                    );
                } else {
                    return this.$t("Unknown");
                }
            }
        },

        renderSubtitle: function (subId: string, rTick: number) {
            if (rTick < 0 || !this.metadata || !this.metadata.subtitles || !subId) {
                return this.$t("No subtitles");
            }

            for (let sub of this.metadata.subtitles) {
                if (sub.id === subId) {
                    return sub.name;
                }
            }

            return this.$t("No subtitles");
        },

        renderAudio: function (audioId: string, rTick: number) {
            if (rTick < 0 || !this.metadata || !this.metadata.audios || !audioId) {
                return "(" + this.$t("From video") + ")";
            }

            for (let aud of this.metadata.audios) {
                if (aud.id === audioId) {
                    return aud.name;
                }
            }

            return "(" + this.$t("From video") + ")";
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

        renderToggleDelay: function (d: number) {
            switch (d) {
            case 0:
                return this.$t("No delay");
            case 250:
                return "0.25 s";
            case 500:
                return "0.5 s";
            case 750:
                return "0.75 s";
            case 1000:
                return "1 s";
            default:
                return "" + d;
            }
        },

        updateSubtitleSize: function (s: string) {
            this.subSizeState = s;
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
            this.subBackgroundState = s;
            PlayerPreferences.SetSubtitlesBackground(s);
        },

        updateResolutions: function () {
            if (this.metadata && this.metadata.resolutions) {
                this.resolutions = this.metadata.resolutions.slice();
            } else {
                this.resolutions = [];
            }
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

        keyDownHandle: function (e: KeyboardEvent) {
            if (e.ctrlKey) {
                return;
            }
            if (e.key === "Escape") {
                this.close();
                e.stopPropagation();
            }
        },
    },
    mounted: function () {
        this.updateResolutions();
        this.subtitles = PlayerPreferences.SelectedSubtitles;
        this.$options.focusTrap = new FocusTrap(
            this.$el,
            this.close.bind(this),
            "player-settings-no-trap"
        );
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
                if (this.$options.focusTrap) {
                    this.$options.focusTrap.activate();
                }
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                if (this.$options.focusTrap) {
                    this.$options.focusTrap.deactivate();
                }
            }
        },
        rTick: function () {
            this.updateResolutions();
        },
    },
});
</script>
