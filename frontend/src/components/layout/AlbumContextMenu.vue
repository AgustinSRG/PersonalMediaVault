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
        <div v-if="mediaIndex > 0" tabindex="0" @click="moveMediaUp" @keydown="clickOnEnter" class="album-body-item-options-menu-btn">
            <i class="fas fa-arrow-up"></i> {{ $t("Move up") }}
        </div>
        <div
            v-if="mediaIndex < albumLength - 1"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="moveMediaDown"
            class="album-body-item-options-menu-btn"
        >
            <i class="fas fa-arrow-down"></i> {{ $t("Move down") }}
        </div>
        <div tabindex="0" @keydown="clickOnEnter" @click="changePosition" class="album-body-item-options-menu-btn">
            <i class="fas fa-arrows-up-down-left-right"></i>
            {{ $t("Change position") }}
        </div>
        <div tabindex="0" @keydown="clickOnEnter" @click="removeMedia" class="album-body-item-options-menu-btn">
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
    emits: ["update:shown", "move-up", "move-down", "change-pos", "media-remove"],
    props: {
        shown: Boolean,

        mediaIndex: Number,
        albumLength: Number,

        x: Number,
        y: Number,
    },
    setup(props) {
        return {
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
    methods: {
        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

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

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this.computeDimensions();

        this._handles.hideHandler = this.hide.bind(this);

        document.addEventListener("mousedown", this._handles.hideHandler);
        document.addEventListener("touchstart", this._handles.hideHandler);

        this._handles.focusTrap = new FocusTrap(this.$el, this.hide.bind(this), "album-body-btn");
    },
    beforeUnmount: function () {
        document.removeEventListener("mousedown", this._handles.hideHandler);
        document.removeEventListener("touchstart", this._handles.hideHandler);
        this._handles.focusTrap.destroy();
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
                this._handles.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this._handles.focusTrap.deactivate();
            }
        },
    },
});
</script>
