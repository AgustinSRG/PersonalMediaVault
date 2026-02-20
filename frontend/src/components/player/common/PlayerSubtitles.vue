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
                v-html="renderSubtitles(subtitles, allowLineBreaks, allowColors)"
            ></div>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { getSubtitlesOptions } from "@/control/player-preferences";
import { toHtmlMinimal } from "@/utils/html-min";
import { EVENT_NAME_SUBTITLES_OPTIONS_CHANGED } from "@/control/app-events";

export default defineComponent({
    name: "PlayerSubtitles",
    props: {
        showControls: Boolean,
        subtitles: String,
    },
    data: function () {
        const options = getSubtitlesOptions();

        return {
            size: options.size,
            customSize: options.customSize,
            bg: options.bg,
            pos: options.pos,
            allowLineBreaks: options.allowLineBreaks,
            allowColors: options.allowColors,
            margin: options.margin,
        };
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED, this.fetchSubtitlesOptions.bind(this));
    },
    methods: {
        renderSubtitles: function (text: string, allowLineBreaks: boolean, allowColors: boolean): string {
            return toHtmlMinimal(text, {
                allowColors,
                allowLineBreaks,
            });
        },

        fetchSubtitlesOptions: function () {
            const options = getSubtitlesOptions();

            if (options.size !== this.size) {
                this.size = options.size;
            }

            if (options.customSize !== this.customSize) {
                this.customSize = options.customSize;
            }

            if (options.allowColors !== this.allowColors) {
                this.allowColors = options.allowColors;
            }
            if (options.allowLineBreaks !== this.allowLineBreaks) {
                this.allowLineBreaks = options.allowLineBreaks;
            }

            if (options.bg !== this.bg) {
                this.bg = options.bg;
            }

            if (options.pos !== this.pos) {
                this.pos = options.pos;
            }

            if (options.margin !== this.margin) {
                this.margin = options.margin;
            }
        },
    },
});
</script>
