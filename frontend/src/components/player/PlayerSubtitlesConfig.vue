<template>
    <div>
        <table v-if="page === 'home'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Style options") }}) </b>
                </td>
                <td class="td-right"></td>
            </tr>

            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('subSizes')">
                <td>
                    <b>{{ $t("Size") }}</b>
                </td>
                <td class="td-right">
                    {{ renderSubtitleSize(size, customSize) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('subBackground')">
                <td>
                    <b>{{ $t("Background") }}</b>
                </td>
                <td class="td-right">
                    {{ renderSubtitleBackground(bg) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('subPosition')">
                <td>
                    <b>{{ $t("Position") }}</b>
                </td>
                <td class="td-right">
                    {{ renderSubtitlePosition(pos) }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('margin')">
                <td>
                    <b>{{ $t("Margin") }}</b>
                </td>
                <td class="td-right">
                    {{ margin + " px" }}
                    <i class="fas fa-chevron-right arrow-config"></i>
                </td>
            </tr>

            <tr>
                <td>
                    <b>{{ $t("Allow colors") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="allowColors" @update:val="onAllowColorsUpdated"></ToggleSwitch>
                </td>
            </tr>

            <tr>
                <td>
                    <b>{{ $t("Allow line breaks") }}</b>
                </td>
                <td class="td-right">
                    <ToggleSwitch v-model:val="allowLineBreaks" @update:val="onAllowLineBreaksUpdated"></ToggleSwitch>
                </td>
            </tr>
        </table>

        <table v-if="page === 'subSizes'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Size") }}) </b>
                </td>
                <td class="td-right" @click="goPage('subSize-custom', $event)">
                    <a href="#subtitles-size-custom" @click="goPage('subSize-custom', $event)">{{ $t("Custom") }}</a>
                </td>
            </tr>
            <tr v-for="s in subtitlesSizes" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="updateSubtitleSize(s)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== size }"></i>
                    {{ renderSubtitleSize(s, customSize) }}
                </td>
                <td class="td-right"></td>
            </tr>
            <tr v-if="size === 'custom'" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="updateSubtitleSize('custom')">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': size !== 'custom' }"></i>
                    {{ $t("Custom") }}: {{ (customSize || 0) + "%" }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>

        <table v-if="page === 'subSize-custom'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('subSizes')">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Size") }}) ({{ $t("Custom") }}) </b>
                </td>
                <td class="td-right"></td>
            </tr>

            <tr>
                <td colspan="2">
                    <input
                        type="range"
                        class="form-range"
                        v-model.number="customSize"
                        @input="saveCustomSubtitleSize"
                        :min="50"
                        :max="250"
                        :step="1"
                    />
                </td>
            </tr>

            <tr>
                <td colspan="2" class="custom-size-row">
                    <input
                        type="number"
                        class="form-control custom-size-input"
                        v-model.number="customSize"
                        @input="saveCustomSubtitleSize"
                        :min="1"
                        :step="1"
                    />
                    <b class="custom-size-unit">%</b>
                </td>
            </tr>
        </table>

        <table v-if="page === 'margin'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Margin") }})</b>
                </td>
                <td class="td-right"></td>
            </tr>

            <tr>
                <td colspan="2">
                    <input
                        type="range"
                        class="form-range"
                        v-model.number="margin"
                        @input="updateSubtitlesMargin"
                        :min="0"
                        :max="250"
                        :step="1"
                    />
                </td>
            </tr>

            <tr>
                <td colspan="2" class="custom-size-row">
                    <input
                        type="number"
                        class="form-control custom-size-input"
                        v-model.number="margin"
                        @input="updateSubtitlesMargin"
                        :min="0"
                        :step="1"
                    />
                    <b class="custom-size-unit">px</b>
                </td>
            </tr>
        </table>

        <table v-if="page === 'subBackground'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
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
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== bg }"></i>
                    {{ renderSubtitleBackground(s) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>

        <table v-if="page === 'subPosition'">
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }} ({{ $t("Position") }}) </b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="s in subtitlesPositions"
                :key="s"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="updateSubtitlePosition(s)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== pos }"></i>
                    {{ renderSubtitlePosition(s) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </table>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { EVENT_NAME_SUBTITLES_OPTIONS_CHANGED, getSubtitlesOptions, setSubtitlesOptions } from "@/control/player-preferences";
import { AppEvents } from "@/control/app-events";

export default defineComponent({
    components: { ToggleSwitch },
    name: "PlayerSubtitlesConfig",
    emits: ["page-switch", "go-back"],
    setup: function () {
        return {
            subtitlesSizes: ["s", "m", "l", "xl", "xxl"],
            subtitlesBackgrounds: ["100", "75", "50", "25", "0"],
            subtitlesPositions: ["bottom", "top"],
        };
    },
    data: function () {
        const options = getSubtitlesOptions();

        return {
            page: "home",

            size: options.size,
            customSize: options.customSize,
            bg: options.bg,
            pos: options.pos,
            allowLineBreaks: options.allowLineBreaks,
            allowColors: options.allowColors,
            margin: options.margin,
        };
    },
    methods: {
        goBack: function () {
            this.$emit("go-back");
        },

        onOptionsUpdate: function () {
            AppEvents.Emit(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED);
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

        clickOnEnter: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                (event.target as HTMLElement).click();
            }
        },

        goPage: function (page: string, e?: Event) {
            if (e) {
                e.preventDefault();
                e.stopPropagation();
            }

            this.page = page;
            this.$emit("page-switch");
        },

        saveCustomSubtitleSize: function () {
            if (typeof this.customSize !== "number" || isNaN(this.customSize)) {
                return;
            }

            const options = getSubtitlesOptions();
            options.size = "custom";
            options.customSize = this.customSize;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
        },

        renderSubtitleSize: function (s: string, customSize: number) {
            switch (s) {
                case "s":
                    return this.$t("Small");
                case "l":
                    return this.$t("Large");
                case "xl":
                    return this.$t("Extra large");
                case "xxl":
                    return this.$t("Extra extra large");
                case "custom":
                    return this.$t("Custom") + " (" + (customSize || 0) + "%)";
                default:
                    return this.$t("Medium");
            }
        },

        updateSubtitleSize: function (s: string) {
            this.size = s;

            const options = getSubtitlesOptions();
            options.size = s;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
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
            this.bg = s;

            const options = getSubtitlesOptions();
            options.bg = s;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
        },

        renderSubtitlePosition: function (s: string) {
            switch (s) {
                case "top":
                    return this.$t("Top");
                default:
                    return this.$t("Bottom");
            }
        },

        updateSubtitlePosition: function (s: string) {
            this.pos = s;

            const options = getSubtitlesOptions();
            options.pos = s;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
        },

        updateSubtitlesMargin: function () {
            if (typeof this.margin !== "number" || isNaN(this.margin)) {
                return;
            }

            const options = getSubtitlesOptions();
            options.margin = this.margin;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
        },

        onAllowColorsUpdated: function () {
            const options = getSubtitlesOptions();
            options.allowColors = this.allowColors;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
        },

        onAllowLineBreaksUpdated: function () {
            const options = getSubtitlesOptions();
            options.allowLineBreaks = this.allowLineBreaks;
            setSubtitlesOptions(options);
            this.onOptionsUpdate();
        },
    },
    mounted: function () {
        this.page = "home";
        this.$listenOnAppEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED, this.fetchSubtitlesOptions.bind(this));
    },
});
</script>
