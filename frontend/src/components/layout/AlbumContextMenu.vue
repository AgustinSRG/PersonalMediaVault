<template>
    <div
        class="album-body-item-options-menu"
        :class="{
            hidden: !shown,
        }"
        :style="{
            top: top,
            left: left,
            right: right,
            bottom: bottom,
            width: width,
            'max-width': maxWidth,
            'max-height': maxHeight,
        }"
        tabindex="-1"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
    >
        <div v-if="mediaIndex > 0" tabindex="0" class="album-body-item-options-menu-btn" @click="moveMediaUp" @keydown="clickOnEnter">
            <i class="fas fa-arrow-up"></i> {{ $t("Move up") }}
        </div>
        <div
            v-if="mediaIndex < albumLength - 1"
            tabindex="0"
            class="album-body-item-options-menu-btn"
            @keydown="clickOnEnter"
            @click="moveMediaDown"
        >
            <i class="fas fa-arrow-down"></i> {{ $t("Move down") }}
        </div>
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="changePosition">
            <i class="fas fa-arrows-up-down-left-right"></i>
            {{ $t("Change position") }}
        </div>
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="removeMedia">
            <i class="fas fa-trash-alt"></i> {{ $t("Remove from the album") }}
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "AlbumContextMenu",
    props: {
        shown: Boolean,

        mediaIndex: Number,
        albumLength: Number,

        x: Number,
        y: Number,
    },
    emits: ["update:shown", "move-up", "move-down", "change-pos", "media-remove"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            shownState: useVModel(props, "shown"),
        };
    },
    data: function () {
        return {
            top: "",
            left: "",
            right: "",
            bottom: "",

            width: "",

            maxWidth: "",
            maxHeight: "",
        };
    },
    watch: {
        x: function () {
            this.computeDimensions();
        },
        y: function () {
            this.computeDimensions();
        },
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
        this.computeDimensions();

        this.$listenOnDocumentEvent("mousedown", this.hide.bind(this));
        this.$listenOnDocumentEvent("touchstart", this.hide.bind(this));

        this.focusTrap = new FocusTrap(this.$el, this.hide.bind(this), "album-body-btn");
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        moveMediaUp: function () {
            this.$emit("move-up", this.mediaIndex);
            this.hide();
        },

        moveMediaDown: function () {
            this.$emit("move-down", this.mediaIndex);
            this.hide();
        },

        changePosition: function () {
            this.$emit("change-pos", this.mediaIndex);
            this.hide();
        },

        removeMedia: function () {
            this.$emit("media-remove", this.mediaIndex);
            this.hide();
        },

        hide: function () {
            this.shownState = false;
        },

        computeDimensions: function () {
            const pageWidth = window.innerWidth;
            const pageHeight = window.innerHeight;

            const x = this.x;
            const y = this.y;

            if (y > pageHeight / 2) {
                const bottom = pageHeight - y;
                const right = pageWidth - x;

                const maxWidth = pageWidth - right;

                const maxHeight = pageHeight - bottom;

                this.top = "auto";
                this.left = "auto";
                this.right = right + "px";
                this.bottom = bottom + "px";
                this.width = "auto";
                this.maxWidth = maxWidth + "px";
                this.maxHeight = maxHeight + "px";
            } else {
                const top = y;
                const right = pageWidth - x;

                const maxWidth = pageWidth - right;

                const maxHeight = pageHeight - top;

                this.top = top + "px";
                this.left = "auto";
                this.right = right + "px";
                this.bottom = "auto";
                this.width = "auto";
                this.maxWidth = maxWidth + "px";
                this.maxHeight = maxHeight + "px";
            }
        },
    },
});
</script>
