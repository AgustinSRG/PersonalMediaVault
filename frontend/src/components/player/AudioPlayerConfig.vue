<template>
    <div
        class="audio-player-config"
        tabindex="-1"
        :class="{ hidden: !shown }"
        role="dialog"
        :aria-hidden="!shown"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @mouseenter="enterConfig"
        @mouseleave="leaveConfig"
        @keydown="keyDownHandle"
    >
        <table v-if="page === ''">
            <tr v-if="!isShort">
                <td>
                    <i class="fas fa-repeat icon-config"></i>
                    <b>{{ $t("Loop") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="loopState"></ToggleSwitch>
                </td>
            </tr>
            <tr v-if="!isShort">
                <td>
                    <i class="fas fa-forward icon-config"></i>
                    <b>{{ $t("Auto next") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="nextEndState"></ToggleSwitch>
                </td>
            </tr>
            <tr v-if="isShort" class="tr-button" tabindex="0" @click="goToAutoNext" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-forward icon-config"></i>
                    <b>{{ $t("Auto next") }}</b>
                </td>
                <td class="td-right">
                    {{ renderAutoNext(autoNext) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>
            <tr class="tr-button" tabindex="0" @click="goToSpeeds" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-gauge icon-config"></i>
                    <b>{{ $t("Playback speed") }}</b>
                </td>
                <td class="td-right">
                    {{ renderSpeed(speed) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>
            <tr class="tr-button" tabindex="0" @click="goToAnimStyles" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-chart-column icon-config"></i>
                    <b>{{ $t("Animation style") }}</b>
                </td>
                <td class="td-right">
                    {{ renderAnimStyle(animColors) }}
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

            <tr v-if="metadata.subtitles && metadata.subtitles.length > 0 && subtitles">
                <td>
                    <i class="fas fa-closed-captioning icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Allow HTML") }})</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="subHTMLState"></ToggleSwitch>
                </td>
            </tr>
        </table>
        <table v-if="page === 'speed'">
            <tr class="tr-button" tabindex="0" @click="goBack" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Playback speed") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr v-for="s in speeds" :key="s" class="tr-button" tabindex="0" @click="changeSpeed(s)" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== speed }"></i>
                    {{ renderSpeed(s) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>
        <table v-if="page === 'anim'">
            <tr class="tr-button" tabindex="0" @click="goBack" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Animation style") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr v-for="s in animStyles" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="setAnimStyle(s)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== animColors }"></i>
                    {{ renderAnimStyle(s) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>

        <table v-if="page === 'subtitles'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeSubtitle('')">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': '' !== subtitles }"></i>
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
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': sub.id !== subtitles }"></i>
                    {{ sub.name }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>

        <table v-if="page === 'subSizes'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Size") }}) </b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr v-for="s in subtitlesSizes" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="updateSubtitleSize(s)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== subSize }"></i>
                    {{ renderSubtitleSize(s) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>

        <table v-if="page === 'subBackground'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
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
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== subBackground }"></i>
                    {{ renderSubtitleBackground(s) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>

        <table v-if="page === 'auto-next'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Auto next") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr v-for="b in autoNextOptions" :key="b" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeAutoNext(b)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': b !== autoNext }"></i>
                    {{ renderAutoNext(b) }}
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
    name: "AudioPlayerConfig",
    emits: [
        "update:shown",
        "update:loop",
        "update:nextEnd",
        "update:speed",
        "update:animColors",
        "update:subSize",
        "update:subBackground",
        "update:subHTML",
        "update-auto-next",
        "enter",
        "leave",
    ],
    props: {
        shown: Boolean,
        metadata: Object,
        loop: Boolean,
        nextEnd: Boolean,
        speed: Number,
        animColors: String,
        subSize: String,
        subBackground: String,
        subHTML: Boolean,
        rTick: Number,
        isShort: Boolean,
    },
    setup(props) {
        return {
            shownState: useVModel(props, "shown"),
            loopState: useVModel(props, "loop"),
            nextEndState: useVModel(props, "nextEnd"),
            speedState: useVModel(props, "speed"),
            animColorsState: useVModel(props, "animColors"),
            subSizeState: useVModel(props, "subSize"),
            subBackgroundState: useVModel(props, "subBackground"),
            subHTMLState: useVModel(props, "subHTML"),
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

            autoNext: PlayerPreferences.ImageAutoNext,
            autoNextOptions: [0, 3, 5, 10, 15, 20, 25, 30],
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
            this.page = "subSizes";
            this.focus();
        },

        goToSubBackgrounds: function () {
            this.page = "subBackground";
            this.focus();
        },

        goToAutoNext: function () {
            this.page = "auto-next";
            this.focus();
        },

        changeAutoNext: function (b) {
            this.autoNext = b;
            PlayerPreferences.SetImageAutoNext(b);
            this.$emit("update-auto-next");
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

        renderAutoNext: function (s: number) {
            if (!isNaN(s) && isFinite(s) && s > 0) {
                if (s === 1) {
                    return s + " " + this.$t("second");
                } else {
                    return s + " " + this.$t("seconds");
                }
            } else {
                return this.$t("Disabled");
            }
        },

        updateSubtitleBackground: function (s: string) {
            this.subBackgroundState = s;
            PlayerPreferences.SetSubtitlesBackground(s);
        },

        renderSubtitle: function (subId: string, rTick: number) {
            if (rTick < 0 || !this.metadata || !this.metadata.subtitles || !subId) {
                return this.$t("No subtitles");
            }

            for (const sub of this.metadata.subtitles) {
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
        this._handles = Object.create(null);
        this.subtitles = PlayerPreferences.SelectedSubtitles;
        this._handles.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
    },
    beforeUnmount: function () {
        this._handles.focusTrap.destroy();
    },
    watch: {
        shown: function () {
            this.page = "";
            if (this.shown) {
                this._handles.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this._handles.focusTrap.deactivate();
            }
        },
    },
});
</script>
