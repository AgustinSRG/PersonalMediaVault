<template>
    <div>
        <table v-if="page === 'home'">
            <tbody>
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
            </tbody>
        </table>

        <table v-if="page === 'subSizes'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Subtitles") }} ({{ $t("Size") }}) </b>
                    </td>
                    <td class="td-right" @click="goPage('subSize-custom', $event)">
                        <a href="#subtitles-size-custom" @click="goPage('subSize-custom', $event)">{{ $t("Custom") }}</a>
                    </td>
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
            </tbody>
        </table>

        <table v-if="page === 'subSize-custom'">
            <tbody>
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
                            v-model.number="customSize"
                            type="range"
                            class="form-range"
                            :min="50"
                            :max="250"
                            :step="1"
                            @input="saveCustomSubtitleSize"
                        />
                    </td>
                </tr>

                <tr>
                    <td colspan="2" class="custom-size-row">
                        <input
                            v-model.number="customSize"
                            type="number"
                            class="form-control custom-size-input"
                            :min="1"
                            :step="1"
                            @input="saveCustomSubtitleSize"
                        />
                        <b class="custom-size-unit">%</b>
                    </td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'margin'">
            <tbody>
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
                            v-model.number="margin"
                            type="range"
                            class="form-range"
                            :min="0"
                            :max="250"
                            :step="1"
                            @input="updateSubtitlesMargin"
                        />
                    </td>
                </tr>

                <tr>
                    <td colspan="2" class="custom-size-row">
                        <input
                            v-model.number="margin"
                            type="number"
                            class="form-control custom-size-input"
                            :min="0"
                            :step="1"
                            @input="updateSubtitlesMargin"
                        />
                        <b class="custom-size-unit">px</b>
                    </td>
                </tr>
            </tbody>
        </table>

        <table v-if="page === 'subBackground'">
            <tbody>
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
            </tbody>
        </table>

        <table v-if="page === 'subPosition'">
            <tbody>
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
            </tbody>
        </table>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { getSubtitlesOptions, setSubtitlesOptions } from "@/control/player-preferences";
import { emitAppEvent, EVENT_NAME_SUBTITLES_OPTIONS_CHANGED } from "@/control/app-events";

export default defineComponent({
    name: "PlayerSubtitlesConfig",
    components: { ToggleSwitch },
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
    mounted: function () {
        this.page = "home";
        this.$listenOnAppEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED, this.fetchSubtitlesOptions.bind(this));
    },
    methods: {
        goBack: function () {
            this.$emit("go-back");
        },

        onOptionsUpdate: function () {
            emitAppEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED);
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
});
</script>
