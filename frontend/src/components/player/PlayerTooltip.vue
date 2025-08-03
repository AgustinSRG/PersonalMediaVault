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

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import type { MediaListItem } from "@/api/models";

const LEFT_TOOLTIPS = ["play", "pause", "prev", "next", "volume", "scale"];

export default defineComponent({
    name: "PlayerTooltip",
    components: {
        PlayerMediaChangePreview,
    },
    props: {
        next: Object as PropType<MediaListItem>,
        prev: Object as PropType<MediaListItem>,

        pageNext: Boolean,
        pagePrev: Boolean,

        helpTooltip: String,

        hasDescription: Boolean,

        muted: Boolean,
        volume: Number,

        fit: Boolean,
        scale: Number,

        scaleRangePercent: Number,

        hideRightTooltip: Boolean,
    },
    computed: {
        isLeft(): boolean {
            return LEFT_TOOLTIPS.includes(this.helpTooltip);
        },
        isHidden(): boolean {
            if (!this.helpTooltip) {
                return true;
            }

            switch (this.helpTooltip) {
                case "next":
                    return !this.next && !this.pageNext;
                case "prev":
                    return !this.prev && !this.pagePrev;
                default:
                    return !this.isLeft && this.hideRightTooltip;
            }
        },
        helpTooltipText(): string {
            switch (this.helpTooltip) {
                case "play":
                    return this.$t("Play");
                case "pause":
                    return this.$t("Pause");
                case "volume":
                    return this.$t("Volume") + " (" + (this.muted ? this.$t("Muted") : this.renderVolume(this.volume)) + ")";
                case "scale":
                    return this.$t("Scale") + " (" + (this.fit ? this.$t("Fit") : this.renderScale(this.scale)) + ")";
                case "desc":
                    return this.hasDescription ? this.$t("Description") : this.$t("Add description");
                case "attachments":
                    return this.$t("Attachments");
                case "related-media":
                    return this.$t("Related media");
                case "config":
                    return this.$t("Player Configuration");
                case "albums":
                    return this.$t("Manage albums");
                case "full-screen":
                    return this.$t("Full screen");
                case "full-screen-exit":
                    return this.$t("Exit full screen");
                default:
                    return "???";
            }
        },
    },
    methods: {
        renderVolume: function (v: number): string {
            return Math.round(v * 100) + "%";
        },

        renderScale: function (v: number): string {
            return Math.round(50 + v * this.scaleRangePercent) + "%";
        },
    },
});
</script>
