<template>
    <div
        class="related-media-list"
        :class="{ hidden: !shown }"
        tabindex="-1"
        role="dialog"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @mouseenter="enterConfig"
        @mouseleave="leaveConfig"
        @keydown="keyDownHandle"
    >
        <a
            v-for="media in relatedMedia || []"
            :key="media.id"
            class="related-media-item"
            tabindex="0"
            :href="getMediaURL(media.id)"
            :title="media.title || $t('Untitled')"
            target="_blank"
            rel="noopener noreferrer"
            @click="clickOnRelatedMedia"
            @keydown="clickOnEnter"
        >
            <MediaItemAlbumThumbnail :item="media"></MediaItemAlbumThumbnail>
            <div class="related-media-item-title">{{ media.title || $t("Untitled") }}</div>
        </a>
    </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "@/utils/v-model";
import { FocusTrap } from "@/utils/focus-trap";
import type { MediaListItem } from "@/api/models";
import { getAssetURL, getFrontendUrl } from "@/utils/api";
import MediaItemAlbumThumbnail from "@/components/utils/MediaItemAlbumThumbnail.vue";

export default defineComponent({
    name: "PlayerRelatedMediaList",
    components: {
        MediaItemAlbumThumbnail,
    },
    props: {
        shown: Boolean,
        relatedMedia: Array as PropType<MediaListItem[]>,
    },
    emits: ["update:shown", "enter", "leave"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            shownState: useVModel(props, "shown"),
        };
    },
    watch: {
        shown: function () {
            if (this.shown) {
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        enterConfig: function () {
            this.$emit("enter");
        },

        leaveConfig: function () {
            this.$emit("leave");
        },

        focus: function () {
            nextTick(() => {
                this.$el.focus();
            });
        },

        clickOnRelatedMedia: function (event: Event) {
            event.stopPropagation();
            this.close();
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

        getMediaURL: function (mid: number): string {
            return getFrontendUrl({
                media: mid,
            });
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
        },
    },
});
</script>
