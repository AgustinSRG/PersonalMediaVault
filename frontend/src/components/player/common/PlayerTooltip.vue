<template>
    <div
        class="player-tooltip"
        :class="{
            'player-help-tip-left': isLeft,
            'player-help-tip-right': !isLeft,
            hidden: isHidden,
        }"
    >
        <PlayerMediaChangePreview v-if="prev && helpTooltip === 'prev'" :media="prev" :next="false"></PlayerMediaChangePreview>
        <PlayerMediaChangePreview v-else-if="next && helpTooltip === 'next'" :media="next" :next="true"></PlayerMediaChangePreview>
        <span v-else-if="pagePrev && helpTooltip === 'prev'"> {{ $t("Previous") }}</span>
        <span v-else-if="pageNext && helpTooltip === 'next'"> {{ $t("Next") }}</span>
        <span v-else>{{ helpTooltipText }}</span>
    </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { computed } from "vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import type { HelpTooltipType } from "@/utils/player-tooltip";

// Tooltips to be displayed on the left
const LEFT_TOOLTIPS = ["play", "pause", "prev", "next", "volume", "scale"];

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
    /**
     * Next element
     */
    next: Object as PropType<MediaListItem>,

    /**
     * Previous element
     */
    prev: Object as PropType<MediaListItem>,

    /**
     * Has next element? (page)
     */
    pageNext: Boolean,

    /**
     * Has previous element? (page)
     */
    pagePrev: Boolean,

    /**
     * Type of help tooltip
     */
    helpTooltip: String as PropType<HelpTooltipType>,

    /**
     * Muted?
     */
    muted: Boolean,

    /**
     * Volume
     */
    volume: Number,

    /**
     * Fit image?
     */
    fit: Boolean,

    /**
     * Image scale
     */
    scale: Number,

    /**
     * Scale (percent)
     */
    scaleRangePercent: Number,

    /**
     * Hide right tooltip?
     */
    hideRightTooltip: Boolean,
});

// True if the tooltip must be displayed to the left
const isLeft = computed<boolean>(() => LEFT_TOOLTIPS.includes(props.helpTooltip));

// True if the tooltip must be hidden
const isHidden = computed<boolean>(() => {
    if (!props.helpTooltip) {
        return true;
    }

    switch (props.helpTooltip) {
        case "next":
            return !props.next && !props.pageNext;
        case "prev":
            return !props.prev && !props.pagePrev;
        default:
            return !isLeft.value && props.hideRightTooltip;
    }
});

// Text to display in the tooltip
const helpTooltipText = computed<string>(() => {
    switch (props.helpTooltip) {
        case "play":
            return $t("Play");
        case "pause":
            return $t("Pause");
        case "volume":
            return $t("Volume") + " (" + (props.muted ? $t("Muted") : renderVolume(props.volume)) + ")";
        case "scale":
            return $t("Scale") + " (" + (props.fit ? $t("Fit") : renderScale(props.scale)) + ")";
        case "desc":
            return $t("Description");
        case "attachments":
            return $t("Attachments");
        case "related-media":
            return $t("Related media");
        case "config":
            return $t("Player Configuration");
        case "albums":
            return $t("Manage albums");
        case "full-screen":
            return $t("Full screen");
        case "full-screen-exit":
            return $t("Exit full screen");
        default:
            return "???";
    }
});

/**
 * Renders volume
 * @param v The value
 * @returns The rendered value
 */
const renderVolume = (v: number): string => {
    return Math.round(v * 100) + "%";
};

/**
 * Renders scale
 * @param v The value
 * @returns The rendered value
 */
const renderScale = (v: number): string => {
    return Math.round(50 + v * props.scaleRangePercent) + "%";
};
</script>
