<template>
    <div class="player-media-change-preview">
        <div class="player-media-change-preview-thumbnail">
            <img v-if="thumbnail" :src="getThumbnail(thumbnail)" :alt="$t('Thumbnail')" loading="lazy" />
            <div v-if="!thumbnail" class="player-media-change-preview-no-thumbnail">
                <i v-if="type === 0" class="fas fa-ban"></i>
                <i v-if="type === 1" class="fas fa-image"></i>
                <i v-if="type === 2" class="fas fa-video"></i>
                <i v-if="type === 3" class="fas fa-headphones"></i>
            </div>
            <div class="preview-thumb-tag" v-if="type === 2 || type === 3">
                {{ renderDuration(duration) }}
            </div>
        </div>
        <div class="player-media-change-preview-details">
            <div class="player-media-change-preview-title">
                {{ next ? $t("Next") : $t("Previous") }}:<br />
                {{ title || $t("Untitled") }}
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { GetAssetURL } from "@/utils/request";
import { defineComponent } from "vue";
import { renderTimeSeconds } from "../../utils/time";

export default defineComponent({
    name: "PlayerMediaChangePreview",
    emits: [],
    props: {
        media: Object,
        next: Boolean,
    },
    data: function () {
        return {
            type: 0,
            thumbnail: "",
            title: "",
            duration: 0,
            width: 0,
            height: 0,
            fps: 0,
        };
    },
    methods: {
        getThumbnail(thumb: string) {
            return GetAssetURL(thumb);
        },
        updateData: function () {
            if (this.media) {
                this.type = this.media.type;
                this.thumbnail = this.media.thumbnail;
                this.title = this.media.title;
                this.width = this.media.width;
                this.height = this.media.height;
                this.fps = this.media.fps;
                this.duration = this.media.duration;
            }
        },
        renderDuration: function (s) {
            return renderTimeSeconds(s);
        },
    },
    mounted: function () {
        this.updateData();
    },
    beforeUnmount: function () {},
    watch: {
        media: function () {
            this.updateData();
        },
    },
});
</script>
