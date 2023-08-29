<template>
    <div
        class="player-volume-control"
        :class="{ expanded: expanded, 'player-min': min }"
        :style="{ width: computeFullWidth(width, min, expanded) }"
        @mouseenter="onEnter"
        @mouseleave="onLeave"
    >
        <button class="player-volume-btn" @click="clickOnVolumeButton" :title="$t('Volume')">
            <i v-if="!muted && volume > 0.5" class="fas fa-volume-up"></i>
            <i v-if="!muted && volume <= 0" class="fas fa-volume-off"></i>
            <i v-if="!muted && volume > 0 && volume <= 0.5" class="fas fa-volume-down"></i>
            <i v-if="muted" class="fas fa-volume-mute"></i>
        </button>
        <div
            class="player-volume-btn-expand"
            :class="{ hidden: !expanded }"
            :style="{ width: computeBarContainerWidth(width) }"
            @mousedown="grabVolume"
            @touchstart="grabVolume"
        >
            <div class="player-volume-bar-container" :style="{ width: computeBarContainerInnerWidth(width) }">
                <div class="player-volume-bar" :style="{ width: getVolumeBarWidth(width) }"></div>
                <div class="player-volume-current" :style="{ width: getVolumeBarCurrentWidth(width, volume, muted) }"></div>
                <div class="player-volume-thumb" :style="{ left: getVolumeThumbLeft(width, volume, muted) }"></div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/v-model";
import { isTouchDevice } from "../../utils/touch";

export default defineComponent({
    name: "VolumeControl",
    emits: ["update:volume", "update:muted", "update:expanded", "enter", "leave"],
    props: {
        width: Number,
        volume: Number,
        min: Boolean,
        muted: Boolean,
        expanded: Boolean,
    },
    setup(props) {
        return {
            volumeState: useVModel(props, "volume"),
            mutedState: useVModel(props, "muted"),
            expandedState: useVModel(props, "expanded"),
        };
    },
    data: function () {
        return {
            volumeGrabbed: false,
        };
    },
    methods: {
        onEnter: function () {
            this.$emit("enter");
            this.showVolumeBar();
        },
        onLeave: function () {
            this.$emit("leave");
        },
        computeFullWidth: function (width: number, min: boolean, expanded: boolean) {
            let margins = 40;
            let barWidth = width;
            let btnWidth = 40;

            if (min) {
                btnWidth = 24;
            }

            return btnWidth + (expanded ? barWidth + margins : margins / 2) + "px";
        },
        computeBarContainerWidth(width: number) {
            let margins = 32;
            return width + margins + "px";
        },
        computeBarContainerInnerWidth(width: number) {
            let margins = 16;
            return width + margins + "px";
        },
        clickOnVolumeButton() {
            this.mutedState = !this.mutedState;
        },
        getVolumeBarWidth(width: number) {
            return width + 16 + "px";
        },
        getVolumeBarCurrentWidth(width: number, volume: number, muted: boolean) {
            let actualVolume = volume;

            if (muted) {
                actualVolume = 0;
            }

            actualVolume = Math.max(0, Math.min(1, actualVolume));

            return Math.floor(actualVolume * width) + "px";
        },
        getVolumeThumbLeft(width: number, volume: number, muted: boolean) {
            return this.getVolumeBarCurrentWidth(width, volume, muted);
        },
        showVolumeBar: function () {
            this.expandedState = true;
        },
        hideVolumeBar: function () {
            if (isTouchDevice()) {
                return;
            }
            this.expandedState = false;
        },
        grabVolume(e) {
            this.volumeGrabbed = true;
            if (e.touches && e.touches.length > 0) {
                this.modifyVolumeByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else {
                this.modifyVolumeByMouse(e.pageX, e.pageY);
            }
        },
        dropVolume(e) {
            if (!this.volumeGrabbed) {
                return;
            }
            this.volumeGrabbed = false;
            if (e.touches && e.touches.length > 0) {
                this.modifyVolumeByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else {
                this.modifyVolumeByMouse(e.pageX, e.pageY);
            }
        },
        moveVolume(e) {
            if (!this.volumeGrabbed) {
                return;
            }
            if (e.touches && e.touches.length > 0) {
                this.modifyVolumeByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else {
                this.modifyVolumeByMouse(e.pageX, e.pageY);
            }
        },
        modifyVolumeByMouse: function (x, y) {
            if (typeof x !== "number" || typeof y !== "number" || isNaN(x) || isNaN(y)) {
                return;
            }
            const offset = this.$el.getBoundingClientRect();

            const offsetX = offset.left + 8 + (this.min ? 24 : 40);

            if (x < offsetX) {
                this.changeVolume(0);
            } else {
                const p = x - offsetX;
                const vol = Math.min(1, p / this.width);
                this.changeVolume(vol);
            }
        },
        changeVolume: function (v: number) {
            this.mutedState = false;
            this.volumeState = v;
        },
    },
    mounted: function () {
        this._handles = Object.create(null);

        if (isTouchDevice()) {
            this.expandedState = true;
        }

        this._handles.dropVolumeHandler = this.dropVolume.bind(this);
        document.addEventListener("mouseup", this._handles.dropVolumeHandler);
        document.addEventListener("touchend", this._handles.dropVolumeHandler);

        this._handles.moveVolumeHandler = this.moveVolume.bind(this);

        document.addEventListener("mousemove", this._handles.moveVolumeHandler);
        document.addEventListener("touchmove", this._handles.moveVolumeHandler);
    },
    beforeUnmount: function () {
        document.removeEventListener("mouseup", this._handles.dropVolumeHandler);
        document.removeEventListener("touchend", this._handles.dropVolumeHandler);

        document.removeEventListener("mousemove", this._handles.moveVolumeHandler);
        document.removeEventListener("touchmove", this._handles.moveVolumeHandler);
    },
});
</script>
