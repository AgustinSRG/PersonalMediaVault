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

        <table v-else-if="page === 'subSizes'">
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
                    v-for="s in SUBTITLES_SIZES"
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

        <table v-else-if="page === 'subSize-custom'">
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

        <table v-else-if="page === 'margin'">
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

        <table v-else-if="page === 'subBackground'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Subtitles") }} ({{ $t("Background") }}) </b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr
                    v-for="s in SUBTITLES_BACKGROUNDS"
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

        <table v-else-if="page === 'subPosition'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goPage('home')">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Subtitles") }} ({{ $t("Position") }}) </b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr
                    v-for="s in SUBTITLES_POSITIONS"
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

<script setup lang="ts">
import { ref } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import { getSubtitlesOptions, setSubtitlesOptions } from "@/control/player-preferences";
import { emitAppEvent, EVENT_NAME_SUBTITLES_OPTIONS_CHANGED } from "@/control/app-events";
import { onApplicationEvent } from "@/composables/on-app-event";
import { clickOnEnter } from "@/utils/events";
import { useI18n } from "@/composables/use-i18n";

// Subtitles sizes
const SUBTITLES_SIZES = ["s", "m", "l", "xl", "xxl"];

// Subtitles backgrounds
const SUBTITLES_BACKGROUNDS = ["100", "75", "50", "25", "0"];

// Subtitles positions
const SUBTITLES_POSITIONS = ["bottom", "top"];

// Translation
const { $t } = useI18n();

// Emits
const emit = defineEmits<{
    /**
     * The user want to go back
     */
    (e: "go-back"): void;

    /**
     * The user switched a page internally
     */
    (e: "page-switch"): void;
}>();

// Internal page type
type SubtitlesOptionsPage = "home" | "subSizes" | "subSize-custom" | "margin" | "subBackground" | "subPosition";

// Internal page
const page = ref<SubtitlesOptionsPage>("home");

// Initial options
const initialOptions = getSubtitlesOptions();

// Subtitles size
const size = ref(initialOptions.size);

// Custom subtitles size
const customSize = ref(initialOptions.customSize);

// Subtitles background
const bg = ref(initialOptions.bg);

// Subtitles position
const pos = ref(initialOptions.pos);

// Allow line breaks?
const allowLineBreaks = ref(initialOptions.allowLineBreaks);

// True to allow colors of subtitles
const allowColors = ref(initialOptions.allowColors);

// Subtitles margin
const margin = ref(initialOptions.margin);

onApplicationEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED, () => {
    const options = getSubtitlesOptions();

    if (options.size !== size.value) {
        size.value = options.size;
    }

    if (options.customSize !== customSize.value) {
        customSize.value = options.customSize;
    }

    if (options.allowColors !== allowColors.value) {
        allowColors.value = options.allowColors;
    }
    if (options.allowLineBreaks !== allowLineBreaks.value) {
        allowLineBreaks.value = options.allowLineBreaks;
    }

    if (options.bg !== bg.value) {
        bg.value = options.bg;
    }

    if (options.pos !== pos.value) {
        pos.value = options.pos;
    }

    if (options.margin !== margin.value) {
        margin.value = options.margin;
    }
});

/**
 * Goes back to the subtitles config
 */
const goBack = () => {
    emit("go-back");
};

/**
 * Navigates to a page
 * @param newPage The new page
 * @param e The click event, to stop it from propagating
 */
const goPage = (newPage: SubtitlesOptionsPage, e?: Event) => {
    if (e) {
        e.preventDefault();
        e.stopPropagation();
    }

    page.value = newPage;
    emit("page-switch");
};

/**
 * Call when the options are updated
 */
const onOptionsUpdate = () => {
    emitAppEvent(EVENT_NAME_SUBTITLES_OPTIONS_CHANGED);
};

/**
 * Renders a subtitle size
 * @param s The size kind
 * @param customSize The custom size (pixels)
 * @returns The rendered text to be displayed
 */
const renderSubtitleSize = (s: string, customSize: number) => {
    switch (s) {
        case "s":
            return $t("Small");
        case "l":
            return $t("Large");
        case "xl":
            return $t("Extra large");
        case "xxl":
            return $t("Extra extra large");
        case "custom":
            return $t("Custom") + " (" + (customSize || 0) + "%)";
        default:
            return $t("Medium");
    }
};

/**
 * Updates the subtitles size
 * @param s The size kind
 */
const updateSubtitleSize = (s: string) => {
    size.value = s;

    const options = getSubtitlesOptions();
    options.size = s;

    setSubtitlesOptions(options);

    onOptionsUpdate();
};

/**
 * Saves the chosen custom subtitles size
 */
const saveCustomSubtitleSize = () => {
    if (typeof customSize.value !== "number" || isNaN(customSize.value)) {
        return;
    }

    const options = getSubtitlesOptions();
    options.size = "custom";
    options.customSize = customSize.value;

    setSubtitlesOptions(options);

    onOptionsUpdate();
};

/**
 * Renders the background kind
 * @param s The background kind
 * @returns The text to be displayed to the user
 */
const renderSubtitleBackground = (s: string) => {
    switch (s) {
        case "0":
            return $t("Transparent");
        case "25":
            return $t("Translucid") + " (75%)";
        case "50":
            return $t("Translucid") + " (50%)";
        case "75":
            return $t("Translucid") + " (25%)";
        default:
            return $t("Opaque");
    }
};

/**
 * Updates the subtitles background kind
 * @param s The background kind
 */
const updateSubtitleBackground = (s: string) => {
    bg.value = s;

    const options = getSubtitlesOptions();
    options.bg = s;
    setSubtitlesOptions(options);

    onOptionsUpdate();
};

/**
 * Renders the subtitles position
 * @param s The subtitles position
 * @returns The text to be displayed for the user
 */
const renderSubtitlePosition = (s: string) => {
    switch (s) {
        case "top":
            return $t("Top");
        default:
            return $t("Bottom");
    }
};

/**
 * Updates the subtitles position
 * @param s The position
 */
const updateSubtitlePosition = (s: string) => {
    pos.value = s;

    const options = getSubtitlesOptions();
    options.pos = s;
    setSubtitlesOptions(options);

    onOptionsUpdate();
};

/**
 * Updates the subtitles margin
 */
const updateSubtitlesMargin = () => {
    if (typeof margin.value !== "number" || isNaN(margin.value)) {
        return;
    }

    const options = getSubtitlesOptions();
    options.margin = margin.value;
    setSubtitlesOptions(options);

    onOptionsUpdate();
};

/**
 * Called when the "Allow colors" option changes
 */
const onAllowColorsUpdated = () => {
    const options = getSubtitlesOptions();
    options.allowColors = allowColors.value;
    setSubtitlesOptions(options);

    onOptionsUpdate();
};

/**
 * Called when the "Allow line breaks" option changes
 */
const onAllowLineBreaksUpdated = () => {
    const options = getSubtitlesOptions();
    options.allowLineBreaks = allowLineBreaks.value;
    setSubtitlesOptions(options);

    onOptionsUpdate();
};
</script>
