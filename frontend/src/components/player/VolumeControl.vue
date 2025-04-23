<template>
    <div
        class="player-volume-control"
        :class="{ expanded: expanded, 'player-min': min }"
        :style="{ width: computeFullWidth(width, min, expanded) }"
        @mouseenter="onEnter"
        @mouseleave="onLeave"
    >
        <button class="player-volume-btn" :title="$t('Volume')" @click="clickOnVolumeButton">
            <i v-if="!muted && volume > 0.5" class="fas fa-volume-up"></i>
            <i v-if="!muted && volume <= 0" class="fas fa-volume-off"></i>
            <i v-if="!muted && volume > 0 && volume <= 0.5" class="fas fa-volume-down"></i>
            <i v-if="muted" class="fas fa-volume-mute"></i>
        </button>
        <div
            class="player-volume-btn-expand"
            :class="{ hidden: !expanded }"
            :style="{ width: computeBarContainerWidth(width) }"
            @mousedown="grabVolumeMouse"
            @touchstart="grabVolumeTouch"
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
import { PositionEvent, positionEventFromMouseEvent, positionEventFromTouchEvent } from "@/utils/position-event";

export default defineComponent({
    name: "VolumeControl",
    props: {
        width: Number,
        volume: Number,
        min: Boolean,
        muted: Boolean,
        expanded: Boolean,
    },
    emits: ["update:volume", "update:muted", "update:expanded", "enter", "leave"],
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
    mounted: function () {
        if (isTouchDevice()) {
            this.expandedState = true;
        }

        this.$listenOnDocumentEvent("mouseup", this.dropVolumeMouse.bind(this));
        this.$listenOnDocumentEvent("touchend", this.dropVolumeTouch.bind(this));

        this.$listenOnDocumentEvent("mousemove", this.moveVolumeMouse.bind(this));
        this.$listenOnDocumentEvent("touchmove", this.moveVolumeTouch.bind(this));
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
            const margins = 40;
            const barWidth = width;
            let btnWidth = 40;

            if (min) {
                btnWidth = 24;
            }

            return btnWidth + (expanded ? barWidth + margins : margins / 2) + "px";
        },
        computeBarContainerWidth: function (width: number) {
            const margins = 32;
            return width + margins + "px";
        },
        computeBarContainerInnerWidth: function (width: number) {
            const margins = 16;
            return width + margins + "px";
        },
        clickOnVolumeButton: function () {
            this.mutedState = !this.mutedState;
        },
        getVolumeBarWidth: function (width: number) {
            return width + 16 + "px";
        },
        getVolumeBarCurrentWidth: function (width: number, volume: number, muted: boolean) {
            let actualVolume = volume;

            if (muted) {
                actualVolume = 0;
            }

            actualVolume = Math.max(0, Math.min(1, actualVolume));

            return Math.floor(actualVolume * width) + "px";
        },
        getVolumeThumbLeft: function (width: number, volume: number, muted: boolean) {
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

        grabVolumeMouse: function (e: MouseEvent) {
            this.grabVolume(positionEventFromMouseEvent(e));
        },

        grabVolumeTouch: function (e: TouchEvent) {
            this.grabVolume(positionEventFromTouchEvent(e));
        },

        grabVolume: function (e: PositionEvent) {
            this.volumeGrabbed = true;
            this.modifyVolumeByMouse(e.x, e.y);
        },

        dropVolumeMouse: function (e: MouseEvent) {
            this.dropVolume(positionEventFromMouseEvent(e));
        },

        dropVolumeTouch: function (e: TouchEvent) {
            this.dropVolume(positionEventFromTouchEvent(e));
        },

        dropVolume: function (e: PositionEvent) {
            if (!this.volumeGrabbed) {
                return;
            }
            this.volumeGrabbed = false;
            this.modifyVolumeByMouse(e.x, e.y);
        },

        moveVolumeMouse: function (e: MouseEvent) {
            this.moveVolume(positionEventFromMouseEvent(e));
        },

        moveVolumeTouch: function (e: TouchEvent) {
            this.moveVolume(positionEventFromTouchEvent(e));
        },

        moveVolume: function (e: PositionEvent) {
            if (!this.volumeGrabbed) {
                return;
            }
            this.modifyVolumeByMouse(e.x, e.y);
        },

        modifyVolumeByMouse: function (x: number, y: number) {
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
});
</script>
