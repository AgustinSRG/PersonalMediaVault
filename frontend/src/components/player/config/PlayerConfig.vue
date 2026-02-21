<template>
    <div
        ref="container"
        class="player-config"
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
            <tr v-if="mediaHasDuration && !isShort">
                <td>
                    <i class="fas fa-repeat icon-config"></i>
                    <b>{{ $t("Loop") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="loop"></ToggleSwitch>
                </td>
            </tr>

            <tr v-if="isImage">
                <td>
                    <i class="fas fa-eye-slash icon-config"></i>
                    <b>{{ $t("Hide image notes") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="hideNotes" @update:val="changeNotesVisible"></ToggleSwitch>
                </td>
            </tr>

            <tr v-if="mediaHasDuration && !isShort">
                <td>
                    <i class="fas fa-forward icon-config"></i>
                    <b>{{ $t("Auto next") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="nextEnd"></ToggleSwitch>
                </td>
            </tr>

            <tr v-if="!mediaHasDuration || isShort" class="tr-button" tabindex="0" @click="goPage('auto-next')" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-forward icon-config"></i>
                    <b>{{ $t("Auto next") }}</b>
                </td>
                <td class="td-right">
                    {{ renderAutoNext(autoNext, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr v-if="mediaHasDuration && !isShort && !inAlbum">
                <td>
                    <i class="fas fa-clock icon-config"></i>
                    <b v-if="isVideo">{{ $t("Wait after video ends") }}</b>
                    <b v-else>{{ $t("Wait after audio ends") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="autoNextPageDelay"></ToggleSwitch>
                </td>
            </tr>

            <tr v-if="mediaHasDuration" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('speed')">
                <td>
                    <i class="fas fa-gauge icon-config"></i>
                    <b>{{ $t("Playback speed") }}</b>
                </td>
                <td class="td-right">
                    {{ renderSpeed(speed, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr v-if="isVideo" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('scale')">
                <td>
                    <i class="fas fa-magnifying-glass icon-config"></i>
                    <b>{{ $t("Scale") }}</b>
                </td>
                <td class="td-right">
                    {{ renderScale(scale, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr v-if="isAudio" class="tr-button" tabindex="0" @click="goPage('anim')" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-chart-column icon-config"></i>
                    <b>{{ $t("Animation style") }}</b>
                </td>
                <td class="td-right">
                    {{ renderAnimStyle(animColors, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr v-if="isAudio">
                <td>
                    <i class="fas fa-eye icon-config"></i>
                    <b>{{ $t("Show title") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="showTitle"></ToggleSwitch>
                </td>
            </tr>

            <tr v-if="isAudio">
                <td>
                    <i class="fas fa-eye icon-config"></i>
                    <b>{{ $t("Show thumbnail") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="showThumbnail"></ToggleSwitch>
                </td>
            </tr>

            <tr v-if="mediaHasResolutions" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('resolution')">
                <td>
                    <i class="fas fa-photo-film icon-config"></i>
                    <b>{{ $t("Quality") }}</b>
                </td>
                <td class="td-right">
                    {{ renderResolution(metadata, resolution, rTick, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr v-if="isImage" class="tr-button" tabindex="0" @click="goPage('background')" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-palette icon-config"></i>
                    <b>{{ $t("Background") }}</b>
                </td>
                <td class="td-right">
                    {{ renderBackground(background, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr
                v-if="isVideo && metadata.audios && metadata.audios.length > 0"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="goPage('audio')"
            >
                <td>
                    <i class="fas fa-headphones icon-config"></i>
                    <b>{{ $t("Audio") }}</b>
                </td>
                <td class="td-right">
                    {{ renderAudio(metadata, audioTrack, rTick, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr
                v-if="mediaHasDuration && metadata.subtitles && metadata.subtitles.length > 0"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="goPage('subtitles')"
            >
                <td>
                    <i class="fas fa-closed-captioning icon-config"></i>
                    <b>{{ $t("Subtitles") }}</b>
                </td>
                <td class="td-right">
                    {{ renderSubtitle(metadata, subtitles, rTick, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr v-if="isVideo" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('time-delays')">
                <td>
                    <i class="fas fa-clock icon-config"></i>
                    <b>{{ $t("Toggle play delay") }}</b>
                </td>
                <td class="td-right">
                    {{ renderToggleDelay(toggleDelay, $t) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>
        </table>

        <PlayerAutoNextConfig
            v-else-if="page === 'auto-next'"
            v-model:auto-next="autoNext"
            @update:auto-next="onAutoNextChanged"
            @go-back="goBack"
        ></PlayerAutoNextConfig>

        <PlayerSpeedConfig
            v-else-if="page === 'speed'"
            v-model:speed="speed"
            @go-back="goBack"
            @go-custom-speed="goPage('speed-custom')"
        ></PlayerSpeedConfig>

        <PlayerSpeedCustomConfig
            v-else-if="page === 'speed-custom'"
            v-model:speed="speed"
            @go-back="goPage('speed')"
        ></PlayerSpeedCustomConfig>

        <PlayerScaleConfig
            v-else-if="page === 'scale'"
            v-model:scale="scale"
            @go-back="goBack"
            @go-custom-scale="goPage('scale-custom')"
        ></PlayerScaleConfig>

        <PlayerScaleCustomConfig
            v-else-if="page === 'scale-custom'"
            v-model:scale="scale"
            @go-back="goPage('scale')"
        ></PlayerScaleCustomConfig>

        <PlayerAudioAnimationConfig
            v-else-if="page === 'anim'"
            v-model:anim-colors="animColors"
            @go-back="goBack"
        ></PlayerAudioAnimationConfig>

        <PlayerResolutionConfig
            v-else-if="page === 'resolution'"
            v-model:resolution="resolution"
            :metadata="metadata"
            :r-tick="rTick"
            @go-back="goBack"
        ></PlayerResolutionConfig>

        <PlayerBackgroundConfig
            v-else-if="page === 'background'"
            v-model:background="background"
            @go-back="goBack"
        ></PlayerBackgroundConfig>

        <PlayerAudioTrackConfig
            v-else-if="page === 'audio'"
            v-model:audio-track="audioTrack"
            :metadata="metadata"
            :r-tick="rTick"
            @update:audio-track="onAudioTrackChanged"
            @go-back="goBack"
        ></PlayerAudioTrackConfig>

        <PlayerSubtitlesConfig
            v-else-if="page === 'subtitles'"
            v-model:subtitles="subtitles"
            :metadata="metadata"
            :r-tick="rTick"
            @update:subtitles="onSubtitlesChanged"
            @go-back="goBack"
            @go-subtitles-options="goPage('subtitle-options')"
        ></PlayerSubtitlesConfig>

        <PlayerSubtitlesOptions
            v-else-if="page === 'subtitle-options'"
            @page-switch="focus"
            @go-back="goPage('subtitles')"
        ></PlayerSubtitlesOptions>

        <PlayerTogglePlayDelayConfig
            v-else-if="page === 'time-delays'"
            v-model:toggle-delay="toggleDelay"
            @update:toggle-delay="onTogglePlayDelayChanged"
            @go-back="goBack"
        ></PlayerTogglePlayDelayConfig>
    </div>
</template>

<script setup lang="ts">
import type { MediaData } from "@/api/models";
import { MEDIA_TYPE_AUDIO, MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO } from "@/api/models";
import { clickOnEnter, stopPropagationEvent } from "@/utils/events";
import type { PropType } from "vue";
import { computed, defineAsyncComponent, nextTick, ref, useTemplateRef } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import {
    getAutoNextTime,
    getImageNotesVisible,
    getSelectedSubtitles,
    getTogglePlayDelay,
    setAutoNextTime,
    setImageNotesVisible,
    setSelectedAudioTrack,
    setSelectedSubtitles,
    setTogglePlayDelay,
} from "@/control/player-preferences";
import { useI18n } from "@/composables/use-i18n";
import {
    renderAnimStyle,
    renderAudio,
    renderAutoNext,
    renderBackground,
    renderResolution,
    renderScale,
    renderSpeed,
    renderSubtitle,
    renderToggleDelay,
} from "@/utils/player-config";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { SubtitlesController } from "@/control/subtitles";

const PlayerAutoNextConfig = defineAsyncComponent({
    loader: () => import("./PlayerAutoNextConfig.vue"),
});

const PlayerSpeedConfig = defineAsyncComponent({
    loader: () => import("./PlayerSpeedConfig.vue"),
});

const PlayerSpeedCustomConfig = defineAsyncComponent({
    loader: () => import("./PlayerSpeedCustomConfig.vue"),
});

const PlayerScaleConfig = defineAsyncComponent({
    loader: () => import("./PlayerScaleConfig.vue"),
});

const PlayerScaleCustomConfig = defineAsyncComponent({
    loader: () => import("./PlayerScaleCustomConfig.vue"),
});

const PlayerAudioAnimationConfig = defineAsyncComponent({
    loader: () => import("./PlayerAudioAnimationConfig.vue"),
});

const PlayerResolutionConfig = defineAsyncComponent({
    loader: () => import("./PlayerResolutionConfig.vue"),
});

const PlayerBackgroundConfig = defineAsyncComponent({
    loader: () => import("./PlayerBackgroundConfig.vue"),
});

const PlayerAudioTrackConfig = defineAsyncComponent({
    loader: () => import("./PlayerAudioTrackConfig.vue"),
});

const PlayerSubtitlesConfig = defineAsyncComponent({
    loader: () => import("./PlayerSubtitlesConfig.vue"),
});

const PlayerSubtitlesOptions = defineAsyncComponent({
    loader: () => import("./PlayerSubtitlesOptions.vue"),
});

const PlayerTogglePlayDelayConfig = defineAsyncComponent({
    loader: () => import("./PlayerTogglePlayDelayConfig.vue"),
});

// Translation
const { $t } = useI18n();

// Shown model: Indicates if the config is being displayed
const shown = defineModel<boolean>("shown");

/**
 * Closes the configuration
 */
const close = () => {
    shown.value = false;
};

// Ref to the container element
const container = useTemplateRef("container");

// Props
const props = defineProps({
    /**
     * Media metadata
     */
    metadata: {
        type: Object as PropType<MediaData>,
        required: true,
    },

    /**
     * True if the video is too short, so the loop option
     * is always enabled and treated as an image
     */
    isShort: Boolean,

    /**
     * True if the media is being played in an album
     */
    inAlbum: Boolean,

    /**
     * Reload tick.
     * When updated, the metadata should be re-checked.
     */
    rTick: Number,
});

// True if the media has duration
const mediaHasDuration = computed(() => [MEDIA_TYPE_AUDIO, MEDIA_TYPE_VIDEO].includes(props.metadata.type));

// True if the media has resolutions
const mediaHasResolutions = computed(() => [MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO].includes(props.metadata.type));

// True if the media is an image
const isImage = computed(() => props.metadata.type === MEDIA_TYPE_IMAGE);

// True if the media is a video
const isVideo = computed(() => props.metadata.type === MEDIA_TYPE_VIDEO);

// True if the media is an audio
const isAudio = computed(() => props.metadata.type === MEDIA_TYPE_AUDIO);

// Loop
const loop = defineModel<boolean>("loop");

// Hide image notes?
const hideNotes = ref(!getImageNotesVisible());

// Auto next at the end
const nextEnd = defineModel<boolean>("nextEnd");

// Auto next time
const autoNext = ref(getAutoNextTime());

// Wait after the media ends? (in page context)
const autoNextPageDelay = defineModel<boolean>("autoNextPageDelay");

// Playback speed
const speed = defineModel<number>("speed");

// Video scale
const scale = defineModel<number>("scale");

// Animation colors
const animColors = defineModel<string>("animColors");

// Display the title? (audio)
const showTitle = defineModel<boolean>("showTitle");

// Display the thumbnail? (audio)
const showThumbnail = defineModel<boolean>("showThumbnail");

// Resolution index
const resolution = defineModel<number>("resolution");

// Image background
const background = defineModel<string>("background");

// Selected audio track (for videos)
const audioTrack = defineModel<string>("audioTrack");

// Subtitles
const subtitles = ref(getSelectedSubtitles());

// Toggle play delay (for video)
const toggleDelay = ref(getTogglePlayDelay());

// Emits
const emit = defineEmits<{
    /**
     * Emitted when the user enters the config
     */
    (e: "enter"): void;

    /**
     * Emitted when the user leaves the config
     */
    (e: "leave"): void;

    /**
     * Updates the visibility of image notes
     */
    (e: "update-notes-visible", visible: boolean): void;

    /**
     * The auto-next time was updated
     */
    (e: "update-auto-next"): void;
}>();

// Player config pages
type PlayerConfigPage =
    | ""
    | "auto-next"
    | "speed"
    | "speed-custom"
    | "scale"
    | "scale-custom"
    | "anim"
    | "resolution"
    | "background"
    | "audio"
    | "subtitles"
    | "subtitle-options"
    | "time-delays";

// Current page
const page = ref<PlayerConfigPage>("");

/**
 * Changes the page
 * @param p The page
 */
const goPage = (p: PlayerConfigPage) => {
    page.value = p;
    focus();
};

/**
 * Goes to main menu
 */
const goBack = () => {
    page.value = "";
    focus();
};

/**
 * Focuses the config menu
 */
const focus = () => {
    nextTick(() => {
        container.value?.focus();
    });
};

/**
 * Called when the image notes visibility changes
 */
const changeNotesVisible = () => {
    setImageNotesVisible(!hideNotes.value);
    emit("update-notes-visible", !hideNotes.value);
};

/**
 * Called when the auto-next delay changes,
 * in order to save the preference.
 */
const onAutoNextChanged = () => {
    setAutoNextTime(autoNext.value);
    emit("update-auto-next");
};

/**
 * Called when the audio track updates,
 * in order to save the preference.
 */
const onAudioTrackChanged = () => {
    setSelectedAudioTrack(audioTrack.value);
};

/**
 * Called when the selected subtitles change,
 * to update the preference.
 */
const onSubtitlesChanged = () => {
    setSelectedSubtitles(subtitles.value);

    SubtitlesController.OnSubtitlesChanged(subtitles.value);
};

/**
 * Called when the toggle play delay change,
 * to update the preference.
 */
const onTogglePlayDelayChanged = () => {
    setTogglePlayDelay(toggleDelay.value);
};

/**
 * Mouse enters the configuration
 */
const enterConfig = () => {
    emit("enter");
};

/**
 * Mouse leaves the configuration
 */
const leaveConfig = () => {
    emit("leave");
};

/**
 * Event handler for 'keydown'
 * @param e The keyboard event
 */
const keyDownHandle = (e: KeyboardEvent) => {
    if (e.ctrlKey) {
        return;
    }
    if (e.key === "Escape") {
        close();
        e.stopPropagation();
    }
};

// Focus trap
useFocusTrap(container, shown, close, "player-settings-no-trap", true);
</script>
