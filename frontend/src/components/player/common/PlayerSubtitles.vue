<template>
    <div
        class="player-subtitles-container"
        :class="{
            'controls-hidden': !showControls,
            'pos-top': pos === 'top',
            'pos-bottom': pos !== 'top',
        }"
        :style="{
            '--subtitles-margin': margin + 'px',
        }"
    >
        <div
            v-if="subtitles"
            class="player-subtitles-container-inner"
            :class="{
                'player-subtitles-bg-0': bg === '0',
                'player-subtitles-bg-25': bg === '25',
                'player-subtitles-bg-50': bg === '50',
                'player-subtitles-bg-75': bg === '75',
                'player-subtitles-bg-100': bg === '100',
            }"
        >
            <div
                class="player-subtitles"
                :class="{
                    'player-subtitles-s': size === 's',
                    'player-subtitles-m': size === 'm',
                    'player-subtitles-l': size === 'l',
                    'player-subtitles-xl': size === 'xl',
                    'player-subtitles-xxl': size === 'xxl',
                }"
                :style="
                    size === 'custom'
                        ? {
                              '--subtitles-size-multiplier': customSize / 100,
                          }
                        : {}
                "
                v-html="renderedSubtitles"
            ></div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { getSubtitlesOptions } from "@/global-state/player-preferences";
import { toHtmlMinimal } from "@/utils/html-min";
import { EVENT_NAME_SUBTITLES_OPTIONS_CHANGED } from "@/global-state/app-events";
import { onApplicationEvent } from "@/composables/on-app-event";

// Props
const props = defineProps({
    /**
     * Controls are shown?
     */
    showControls: Boolean,

    /**
     * Subtitles to display
     */
    subtitles: String,
});

// Initial subtitles options
const initialOptions = getSubtitlesOptions();

// Subtitles size
const size = ref(initialOptions.size);

// Custom subtitles size (px)
const customSize = ref(initialOptions.customSize);

// Subtitles background
const bg = ref(initialOptions.bg);

// Subtitles position
const pos = ref(initialOptions.pos);

// Allow line breaks?
const allowLineBreaks = ref(initialOptions.allowLineBreaks);

// Allow colors?
const allowColors = ref(initialOptions.allowColors);

// Subtitles margin
const margin = ref(initialOptions.margin);

onApplicationEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED, () => {
    const options = getSubtitlesOptions();

    size.value = options.size;
    customSize.value = options.customSize;
    bg.value = options.bg;
    pos.value = options.pos;
    allowLineBreaks.value = options.allowLineBreaks;
    allowColors.value = options.allowColors;
    margin.value = options.margin;
});

// Rendered subtitles
const renderedSubtitles = computed(() =>
    toHtmlMinimal(props.subtitles, {
        allowColors: allowColors.value,
        allowLineBreaks: allowLineBreaks.value,
    }),
);
</script>
