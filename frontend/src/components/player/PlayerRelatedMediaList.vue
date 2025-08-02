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
            @click="clickOnMedia(media.id, $event)"
            @keydown="clickOnEnter"
        >
            <div class="album-body-item-thumbnail" :title="media.title || $t('Untitled')">
                <div v-if="!media.thumbnail" class="no-thumb">
                    <i v-if="media.type === 1" class="fas fa-image"></i>
                    <i v-else-if="media.type === 2" class="fas fa-video"></i>
                    <i v-else-if="media.type === 3" class="fas fa-headphones"></i>
                    <i v-else class="fas fa-ban"></i>
                </div>
                <ThumbImage v-if="media.thumbnail" :src="getThumbnail(media.thumbnail)"></ThumbImage>
                <DurationIndicator
                    v-if="media.type === 2 || media.type === 3"
                    :type="media.type"
                    :duration="media.duration"
                    :small="true"
                ></DurationIndicator>
            </div>
            <div class="related-media-item-title">{{ media.title || $t("Untitled") }}</div>
        </a>
    </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import type { MediaListItem } from "@/api/models";
import { generateURIQuery, getAssetURL } from "@/utils/api";
import { AppStatus } from "@/control/app-status";
import ThumbImage from "../utils/ThumbImage.vue";
import DurationIndicator from "../utils/DurationIndicator.vue";

export default defineComponent({
    name: "PlayerRelatedMediaList",
    components: {
        ThumbImage,
        DurationIndicator,
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

        clickOnMedia: function (mid: number, event: Event) {
            event.preventDefault();
            event.stopPropagation();
            this.close();
            AppStatus.ClickOnMedia(mid, false);
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
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    media: mid + "",
                })
            );
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
        },
    },
});
</script>
