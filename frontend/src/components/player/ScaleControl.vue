<template>
    <div
        class="player-scale-control"
        :class="{ expanded: expanded, 'player-min': min }"
        :style="{ width: computeFullWidth(width, min, expanded) }"
        @mouseenter="onEnter"
        @mouseleave="onLeave"
    >
        <button class="player-scale-btn" @click="clickOnScaleButton">
            <i class="fas fa-magnifying-glass"></i>
        </button>
        <div
            class="player-scale-btn-expand"
            :class="{ hidden: !expanded }"
            :style="{ width: computeBarContainerWidth(width) }"
            @mousedown="grabScaleMouse"
            @touchstart.passive="grabScaleTouch"
        >
            <div class="player-scale-bar-container" :style="{ width: computeBarContainerInnerWidth(width) }">
                <div class="player-scale-bar" :style="{ width: getScaleBarWidth(width) }"></div>
                <div class="player-scale-current" :style="{ width: getScaleBarCurrentWidth(width, scale, fit) }"></div>
                <div class="player-scale-thumb" :style="{ left: getScaleThumbLeft(width, scale, fit) }"></div>
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
    name: "ScaleControl",
    props: {
        width: Number,
        scale: Number,
        min: Boolean,
        fit: Boolean,
        expanded: Boolean,
    },
    emits: ["update:scale", "update:fit", "update:expanded", "enter", "leave"],
    setup(props) {
        return {
            scaleState: useVModel(props, "scale"),
            fitState: useVModel(props, "fit"),
            expandedState: useVModel(props, "expanded"),
        };
    },
    data: function () {
        return {
            scaleGrabbed: false,
        };
    },
    mounted: function () {
        if (isTouchDevice()) {
            this.expandedState = true;
        }

        this.$listenOnDocumentEvent("mouseup", this.dropScaleMouse.bind(this));
        this.$listenOnDocumentEvent("touchend", this.dropScaleTouch.bind(this));

        this.$listenOnDocumentEvent("mousemove", this.moveScaleMouse.bind(this));
        this.$listenOnDocumentEvent("touchmove", this.moveScaleTouch.bind(this));
    },
    methods: {
        onEnter: function () {
            this.$emit("enter");
            this.showScaleBar();
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
        clickOnScaleButton: function () {
            this.fitState = !this.fitState;
        },
        getScaleBarWidth: function (width: number) {
            return width + 16 + "px";
        },
        getScaleBarCurrentWidth: function (width: number, scale: number, fit: boolean) {
            let actualScale = scale;

            if (fit) {
                actualScale = 0;
            }

            actualScale = Math.max(0, Math.min(1, actualScale));

            return Math.floor(actualScale * width) + "px";
        },
        getScaleThumbLeft: function (width: number, scale: number, fit: boolean) {
            return this.getScaleBarCurrentWidth(width, scale, fit);
        },
        showScaleBar: function () {
            this.expandedState = true;
        },
        hideScaleBar: function () {
            if (isTouchDevice()) {
                return;
            }
            this.expandedState = false;
        },

        grabScaleMouse: function (e: MouseEvent) {
            this.grabScale(positionEventFromMouseEvent(e));
        },

        grabScaleTouch: function (e: TouchEvent) {
            this.grabScale(positionEventFromTouchEvent(e));
        },

        grabScale: function (e: PositionEvent) {
            this.scaleGrabbed = true;
            this.modifyScaleByMouse(e.x, e.y);
        },

        dropScaleMouse: function (e: MouseEvent) {
            this.dropScale(positionEventFromMouseEvent(e));
        },

        dropScaleTouch: function (e: TouchEvent) {
            this.dropScale(positionEventFromTouchEvent(e));
        },

        dropScale(e: PositionEvent) {
            if (!this.scaleGrabbed) {
                return;
            }
            this.scaleGrabbed = false;
            this.modifyScaleByMouse(e.x, e.y);
        },

        moveScaleMouse: function (e: MouseEvent) {
            this.moveScale(positionEventFromMouseEvent(e));
        },

        moveScaleTouch: function (e: TouchEvent) {
            this.moveScale(positionEventFromTouchEvent(e));
        },

        moveScale: function (e: PositionEvent) {
            if (!this.scaleGrabbed) {
                return;
            }
            this.modifyScaleByMouse(e.x, e.y);
        },

        modifyScaleByMouse: function (x: number, y: number) {
            if (typeof x !== "number" || typeof y !== "number" || isNaN(x) || isNaN(y)) {
                return;
            }
            const offset = this.$el.getBoundingClientRect();

            const offsetX = offset.left + 8 + (this.min ? 24 : 40);

            if (x < offsetX) {
                this.changeScale(0);
            } else {
                const p = x - offsetX;
                const vol = Math.min(1, p / this.width);
                this.changeScale(vol);
            }
        },

        changeScale: function (z: number) {
            this.fitState = false;
            this.scaleState = z;
        },
    },
});
</script>
