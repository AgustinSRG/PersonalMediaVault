<template>
    <div
        class="video-player-config"
        :class="{ hidden: !shown }"
        tabindex="-1"
        role="dialog"
        @click="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @mouseenter="enterConfig"
        @mouseleave="leaveConfig"
        @keydown="keyDownHandle"
    >
        <table v-if="page === ''">
            <tbody>
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

                <tr v-if="!isShort && !inAlbum">
                    <td>
                        <i class="fas fa-clock icon-config"></i>
                        <b>{{ $t("Wait after video ends") }}</b>
                    </td>
                    <td class="td-right">
                        <ToggleSwitch v-model:val="autoNextPageDelayState"></ToggleSwitch>
                    </td>
                </tr>

                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goToSpeeds">
                    <td>
                        <i class="fas fa-gauge icon-config"></i>
                        <b>{{ $t("Playback speed") }}</b>
                    </td>
                    <td class="td-right">
                        {{ renderSpeed(speed) }}
                        <i class="fas fa-chevron-right arrow-config"></i>
                    </td>
                </tr>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goToScales">
                    <td>
                        <i class="fas fa-magnifying-glass icon-config"></i>
                        <b>{{ $t("Scale") }}</b>
                    </td>
                    <td class="td-right">
                        {{ renderScale(scale) }}
                        <i class="fas fa-chevron-right arrow-config"></i>
                    </td>
                </tr>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goToResolutions">
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

                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goToDelays">
                    <td>
                        <i class="fas fa-clock icon-config"></i>
                        <b>{{ $t("Toggle play delay") }}</b>
                    </td>
                    <td class="td-right">
                        {{ renderToggleDelay(toggleDelay) }}
                        <i class="fas fa-chevron-right arrow-config"></i>
                    </td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'speed'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Playback speed") }}</b>
                    </td>
                    <td class="td-right" @click="goToCustomSpeed">
                        <a href="#playback-speed-custom" @click="goToCustomSpeed">{{ $t("Custom") }}</a>
                    </td>
                </tr>
                <tr v-for="s in speeds" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeSpeed(s)">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== speed }"></i>
                        {{ renderSpeed(s) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr v-if="!speeds.includes(speed)" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeSpeed(speed)">
                    <td>
                        <i class="fas fa-check icon-config"></i>
                        {{ $t("Custom") }}: {{ renderSpeed(speed) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'speed-custom'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goToSpeeds">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Playback speed") }} ({{ $t("Custom") }})</b>
                    </td>
                    <td class="td-right"></td>
                </tr>

                <tr>
                    <td colspan="2">
                        <input
                            v-model.number="speedNum"
                            type="range"
                            class="form-range"
                            :min="1"
                            :max="200"
                            :step="1"
                            @input="updateSpeedNum"
                        />
                    </td>
                </tr>

                <tr>
                    <td colspan="2" class="custom-size-row">
                        <input
                            v-model.number="speedNum"
                            type="number"
                            class="form-control custom-size-input"
                            :min="1"
                            :step="1"
                            @input="updateSpeedNum"
                        />
                        <b class="custom-size-unit">%</b>
                    </td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'scales'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Scale") }}</b>
                    </td>
                    <td class="td-right" @click="goToCustomScale">
                        <a href="#video-scale-custom" @click="goToCustomScale">{{ $t("Custom") }}</a>
                    </td>
                </tr>
                <tr v-for="s in scales" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeScale(s)">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== scale }"></i>
                        {{ renderScale(s) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr v-if="!scales.includes(scale)" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeScale(scale)">
                    <td>
                        <i class="fas fa-check icon-config"></i>
                        {{ $t("Custom") }}: {{ renderScale(scale) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'scale-custom'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goToScales">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Scale") }} ({{ $t("Custom") }})</b>
                    </td>
                    <td class="td-right"></td>
                </tr>

                <tr>
                    <td colspan="2">
                        <input
                            v-model.number="scaleNum"
                            type="range"
                            class="form-range"
                            :min="100"
                            :max="800"
                            :step="1"
                            @input="updateScaleNum"
                        />
                    </td>
                </tr>

                <tr>
                    <td colspan="2" class="custom-size-row">
                        <input
                            v-model.number="scaleNum"
                            type="number"
                            class="form-control custom-size-input"
                            :min="1"
                            :step="1"
                            @input="updateScaleNum"
                        />
                        <b class="custom-size-unit">%</b>
                    </td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'resolution'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Quality") }}</b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeResolution(-1)">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': -1 !== resolution }"></i>
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
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': i !== resolution }"></i>
                        {{ renderResolution(i, rTick) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'audios'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Audio") }}</b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeAudioTrack('')">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': '' !== audioTrack }"></i>
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
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': aud.id !== audioTrack }"></i>
                        {{ aud.name }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'subtitles'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Subtitles") }}</b>
                    </td>
                    <td class="td-right" @click="goToSubtitlesOptions">
                        <a href="#subtitle-options" @click="goToSubtitlesOptions">{{ $t("Style options") }}</a>
                    </td>
                </tr>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeSubtitle('')">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': '' !== effectiveSubtitles }"></i>
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
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': sub.id !== effectiveSubtitles }"></i>
                        {{ sub.name }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>

        <PlayerSubtitlesConfig v-if="page === 'subtitle-options'" @page-switch="focus" @go-back="goToSubtitles"></PlayerSubtitlesConfig>

        <table v-if="page === 'time-delays'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
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
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== toggleDelay }"></i>
                        {{ renderToggleDelay(s) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'auto-next'">
            <tbody>
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
            </tbody>
        </table>
    </div>
</template>

<script lang="ts">
import {
    getAutoNextTime,
    getSelectedSubtitles,
    getTogglePlayDelay,
    setAutoNextTime,
    setSelectedAudioTrack,
    setSelectedSubtitles,
    setTogglePlayDelay,
} from "@/control/player-preferences";
import { SubtitlesController } from "@/control/subtitles";
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "@/utils/v-model";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import { FocusTrap } from "@/utils/focus-trap";
import PlayerSubtitlesConfig from "./PlayerSubtitlesConfig.vue";
import type { MediaData, MediaResolution } from "@/api/models";

export default defineComponent({
    name: "VideoPlayerConfig",
    components: { ToggleSwitch, PlayerSubtitlesConfig },
    props: {
        shown: Boolean,
        metadata: Object as PropType<MediaData>,
        loop: Boolean,
        nextEnd: Boolean,
        speed: Number,
        scale: Number,
        resolution: Number,
        rTick: Number,
        audioTrack: String,
        isShort: Boolean,
        inAlbum: Boolean,
        autoNextPageDelay: Boolean,
    },
    emits: [
        "update:shown",
        "update:loop",
        "update:nextEnd",
        "update:speed",
        "update:scale",
        "update:resolution",
        "update:audioTrack",
        "update:autoNextPageDelay",
        "update-auto-next",
        "enter",
        "leave",
    ],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            shownState: useVModel(props, "shown"),
            loopState: useVModel(props, "loop"),
            nextEndState: useVModel(props, "nextEnd"),
            speedState: useVModel(props, "speed"),
            scaleState: useVModel(props, "scale"),
            resolutionState: useVModel(props, "resolution"),
            audioTrackState: useVModel(props, "audioTrack"),
            autoNextPageDelayState: useVModel(props, "autoNextPageDelay"),
        };
    },
    data: function () {
        return {
            page: "",
            speeds: [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2],
            scales: [1, 1.25, 1.5, 1.75, 2, 4, 8],
            resolutions: [] as MediaResolution[],

            speedNum: Math.floor(this.speed * 100),
            scaleNum: Math.floor(this.scale * 100),

            subtitles: "",
            effectiveSubtitles: "",

            subSizeCustomNum: this.subSizeCustom,

            toggleDelay: getTogglePlayDelay(),
            toggleDelayOptions: [0, 250, 500, 750, 1000],

            autoNext: getAutoNextTime(),
            autoNextOptions: [0, 3, 5, 10, 15, 20, 25, 30],
        };
    },
    watch: {
        shown: function () {
            this.page = "";
            if (this.shown) {
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },
        speed: function () {
            this.speedNum = Math.floor(this.speed * 100);
        },
        scale: function () {
            this.scaleNum = Math.floor(this.scale * 100);
        },
        rTick: function () {
            this.updateResolutions();
            this.updateEffectiveSubtitles();
        },
    },
    mounted: function () {
        this.updateResolutions();
        this.subtitles = getSelectedSubtitles();
        this.updateEffectiveSubtitles();
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        changeResolution: function (i: number) {
            this.resolutionState = i;
        },

        changeToggleDelay: function (d: number) {
            this.toggleDelay = d;
            setTogglePlayDelay(d);
        },

        changeSubtitle: function (s: string) {
            this.subtitles = s;
            this.updateEffectiveSubtitles();
            setSelectedSubtitles(s);
            SubtitlesController.OnSubtitlesChanged(s);
        },

        changeAudioTrack: function (s: string) {
            this.audioTrackState = s;
            setSelectedAudioTrack(s);
        },

        changeAutoNext: function (b: number) {
            this.autoNext = b;
            setAutoNextTime(b);
            this.$emit("update-auto-next");
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

        goBack: function () {
            this.page = "";
            this.focus();
        },

        changeSpeed: function (s: number) {
            this.speedState = s;
        },

        goToSpeeds: function () {
            this.page = "speed";
            this.focus();
        },

        changeScale: function (s: number) {
            this.scaleState = s;
        },

        goToScales: function () {
            this.page = "scales";
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

        goToAutoNext: function () {
            this.page = "auto-next";
            this.focus();
        },

        goToSubtitlesOptions: function (e?: Event) {
            if (e) {
                e.preventDefault();
                e.stopPropagation();
            }

            this.page = "subtitle-options";
            this.focus();
        },

        goToCustomSpeed: function (e?: Event) {
            if (e) {
                e.preventDefault();
                e.stopPropagation();
            }

            this.page = "speed-custom";
            this.focus();
        },

        goToCustomScale: function (e?: Event) {
            if (e) {
                e.preventDefault();
                e.stopPropagation();
            }

            this.page = "scale-custom";
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

        renderScale: function (scale: number) {
            if (scale > 1) {
                return Math.floor(scale * 100) + "%";
            } else if (scale < 1) {
                return Math.floor(scale * 100) + "%";
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
                const resData = this.metadata.resolutions[res];
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

        updateEffectiveSubtitles: function () {
            if (!this.metadata || !this.metadata.subtitles || !this.subtitles) {
                this.effectiveSubtitles = "";
                return;
            }

            for (const sub of this.metadata.subtitles) {
                if (sub.id === this.subtitles) {
                    this.effectiveSubtitles = sub.id;
                    return;
                }
            }

            if (this.subtitles && this.metadata.subtitles.length > 0) {
                this.effectiveSubtitles = this.metadata.subtitles[0].id;
            } else {
                this.effectiveSubtitles = "";
            }
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

            if (subId && this.metadata.subtitles.length > 0) {
                return this.metadata.subtitles[0].name;
            }

            return this.$t("No subtitles");
        },

        renderAudio: function (audioId: string, rTick: number) {
            if (rTick < 0 || !this.metadata || !this.metadata.audios || !audioId) {
                return "(" + this.$t("From video") + ")";
            }

            for (const aud of this.metadata.audios) {
                if (aud.id === audioId) {
                    return aud.name;
                }
            }

            return "(" + this.$t("From video") + ")";
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

        updateResolutions: function () {
            if (this.metadata && this.metadata.resolutions) {
                this.resolutions = this.metadata.resolutions.slice();
            } else {
                this.resolutions = [];
            }
        },

        updateSpeedNum: function () {
            if (typeof this.speedNum !== "number" || isNaN(this.speedNum) || this.speedNum < 0.1) {
                return;
            }

            this.speedState = this.speedNum / 100;
        },

        updateScaleNum: function () {
            if (typeof this.scaleNum !== "number" || isNaN(this.scaleNum) || this.scaleNum < 0.1) {
                return;
            }

            this.scaleState = this.scaleNum / 100;
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
});
</script>
