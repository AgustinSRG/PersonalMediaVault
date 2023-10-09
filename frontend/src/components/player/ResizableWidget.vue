<template>
    <div
        class="resizable-widget"
        :class="{ hidden: !display }"
        tabindex="-1"
        @keydown="onKeyDown"
        @dblclick="stopPropagationEvent"
        @mousedown="propagateClick"
        @touchstart="propagateTouch"
        :style="{
            top: y + 'px',
            left: x + 'px',
            width: width + 'px',
            height: height + 'px',
        }"
    >
        <div class="resizable-widget-header" @mousedown="startMoving($event, true)" @touchstart.passive="startMoving($event, false)">
            <div class="resizable-widget-title">{{ title }}</div>
            <div class="resizable-widget-close-btn">
                <button type="button" class="close-button" @click="close" :title="$t('Close')">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        </div>

        <div class="resizable-widget-body">
            <slot></slot>
        </div>

        <div
            class="resize resize-left"
            @mousedown="startResizeWidget($event, 'l', true)"
            @touchstart.passive="startResizeWidget($event, 'l', false)"
        ></div>
        <div
            class="resize resize-top"
            @mousedown="startResizeWidget($event, 't', true)"
            @touchstart.passive="startResizeWidget($event, 't', false)"
        ></div>
        <div
            class="resize resize-right"
            @mousedown="startResizeWidget($event, 'r', true)"
            @touchstart.passive="startResizeWidget($event, 'r', false)"
        ></div>
        <div
            class="resize resize-bottom"
            @mousedown="startResizeWidget($event, 'b', true)"
            @touchstart.passive="startResizeWidget($event, 'b', false)"
        ></div>
        <div
            class="resize resize-corner-top-left"
            @mousedown="startResizeWidget($event, 'tl', true)"
            @touchstart.passive="startResizeWidget($event, 'tl', false)"
        ></div>
        <div
            class="resize resize-corner-top-right"
            @mousedown="startResizeWidget($event, 'tr', true)"
            @touchstart.passive="startResizeWidget($event, 'tr', false)"
        ></div>
        <div
            class="resize resize-corner-bottom-left"
            @mousedown="startResizeWidget($event, 'bl', true)"
            @touchstart.passive="startResizeWidget($event, 'bl', false)"
        ></div>
        <div
            class="resize resize-corner-bottom-right"
            @mousedown="startResizeWidget($event, 'br', true)"
            @touchstart.passive="startResizeWidget($event, 'br', false)"
        ></div>
    </div>
</template>

<script lang="ts">
import { LocalStorage } from "@/control/local-storage";
import { useVModel } from "@/utils/v-model";
import { nextTick } from "vue";
import { defineComponent } from "vue";

const INITIAL_WIDTH = 480;
const INITIAL_HEIGHT = 360;

const MIN_WIDTH = 250;
const MIN_HEIGHT = 250;

export default defineComponent({
    name: "ResizableWidget",
    emits: ["update:display", "clicked"],
    props: {
        display: Boolean,
        positionKey: String,
        contextOpen: Boolean,
        title: String,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },

    data: function () {
        return {
            x: 0,
            y: 0,
            width: INITIAL_WIDTH,
            height: INITIAL_HEIGHT,

            moving: false,
            moveOriginalX: 0,
            moveOriginalY: 0,
            moveStartX: 0,
            moveStartY: 0,

            resizing: false,
            resizeOriginalX: 0,
            resizeOriginalY: 0,
            resizeOriginalW: 0,
            resizeOriginalH: 0,
            resizeStartX: 0,
            resizeStartY: 0,
            resizeMode: "",
        };
    },

    methods: {
        close: function () {
            this.displayStatus = false;
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        onKeyDown: function (e: KeyboardEvent) {
            e.stopPropagation();

            if (e.key === "Escape") {
                this.close();
            }
        },

        propagateClick: function (e) {
            e.stopPropagation();
            if (e.button !== 0) {
                return;
            }
            this.$emit("clicked");
        },

        propagateTouch: function (e) {
            e.stopPropagation();
            this.$emit("clicked");
        },

        loadPosition: function () {
            // Load position
            if (this.positionKey) {
                const savedPosition = LocalStorage.Get(this.positionKey, null);

                if (savedPosition && typeof savedPosition === "object") {
                    this.x = Number(savedPosition.x) || 0;
                    this.y = Number(savedPosition.y) || 0;
                    this.width = Number(savedPosition.width) || 0;
                    this.height = Number(savedPosition.height) || 0;
                } else {
                    // Center with initial size
                    const parentElem: HTMLElement = this.$el.parentElement;

                    if (parentElem) {
                        const bounds = parentElem.getBoundingClientRect();

                        if (bounds.width < INITIAL_WIDTH) {
                            this.x = 0;
                            this.width = bounds.width;
                        } else {
                            this.x = Math.floor((bounds.width - INITIAL_WIDTH) / 2);
                            this.width = INITIAL_WIDTH;
                        }

                        if (bounds.height < INITIAL_HEIGHT) {
                            this.y = 0;
                            this.height = bounds.height;
                        } else {
                            this.y = Math.floor((bounds.height - INITIAL_HEIGHT) / 2);
                            this.height = INITIAL_HEIGHT;
                        }
                    }
                }
            }

            this.fixPosition();
        },

        fixPosition: function () {
            const parentElem: HTMLElement = this.$el.parentElement;

            if (parentElem) {
                const bounds = parentElem.getBoundingClientRect();

                if (bounds.width <= 0 || bounds.height <= 0) {
                    return;
                }

                // X

                if (this.x < 0) {
                    this.x = 0;
                }

                if (this.width < MIN_WIDTH) {
                    this.width = MIN_WIDTH;
                }

                if (this.x >= bounds.width) {
                    this.x = bounds.width - 1;
                }

                if (this.x + this.width >= bounds.width) {
                    // Try move to the left
                    this.x = bounds.width - this.width - 1;
                    if (this.x < 0) {
                        // Does not fit
                        this.width = bounds.width;
                        this.x = 0;
                    }
                }

                // Y

                if (this.y < 0) {
                    this.y = 0;
                }

                if (this.height < MIN_HEIGHT) {
                    this.height = MIN_HEIGHT;
                }

                if (this.y >= bounds.height) {
                    this.y = bounds.height - 1;
                }

                if (this.y + this.height >= bounds.height) {
                    // Try move to the left
                    this.y = bounds.height - this.height - 1;
                    if (this.y < 0) {
                        // Does not fit
                        this.height = bounds.height;
                        this.y = 0;
                    }
                }
            }
        },

        savePosition: function () {
            if (this.positionKey) {
                LocalStorage.Set(this.positionKey, {
                    x: this.x,
                    y: this.y,
                    width: this.width,
                    height: this.height,
                });
            }
        },

        startMoving: function (e, isMouse: boolean) {
            if (this.contextOpen) {
                return;
            }
 
            if (this.moving || this.resizing) {
                return;
            }

            if (isMouse && (e.which || e.button) !== 1) {
                return;
            }

            e.stopPropagation();


            const parentElem: HTMLElement = this.$el.parentElement;

            if (!parentElem) {
                return;
            }

            const bounds = parentElem.getBoundingClientRect();
            
            let x: number;
            let y: number;
            if (e.touches && e.touches.length > 0) {
                x = e.touches[0].pageX;
                y = e.touches[0].pageY;
            } else {
                x = e.pageX;
                y = e.pageY;
            }
            const trueX = x - bounds.left;
            const trueY = y - bounds.top;

            this.moving = true;
            this.moveStartX = trueX;
            this.moveStartY = trueY;
            this.moveOriginalX = this.x;
            this.moveOriginalY = this.y;
        },

        startResizeWidget: function (e: any, resizeMode: string, isMouse: boolean) {
            if (this.contextOpen) {
                return;
            }

            if (this.moving || this.resizing) {
                return;
            }

            if (isMouse && (e.which || e.button) !== 1) {
                return;
            }

            e.stopPropagation();

            const parentElem: HTMLElement = this.$el.parentElement;

            if (!parentElem) {
                return;
            }

            const bounds = parentElem.getBoundingClientRect();

            let x: number;
            let y: number;

            if (e.touches && e.touches.length > 0) {
                x = e.touches[0].pageX;
                y = e.touches[0].pageY;
            } else {
                x = e.pageX;
                y = e.pageY;
            }

            this.resizing = true;
            this.resizeMode = resizeMode;
            this.resizeOriginalX = this.x;
            this.resizeOriginalY = this.y;
            this.resizeOriginalW = this.width;
            this.resizeOriginalH = this.height;

            this.resizeStartX = x - bounds.left;
            this.resizeStartY = y - bounds.top;
        },

        mouseMove: function (e) {
            if (!this.moving && !this.resizing) {
                return;
            }

            const parentElem: HTMLElement = this.$el.parentElement;

            if (!parentElem) {
                return;
            }

            const bounds = parentElem.getBoundingClientRect();

            let x: number;
            let y: number;
            if (e.touches && e.touches.length > 0) {
                x = e.touches[0].pageX;
                y = e.touches[0].pageY;
            } else {
                x = e.pageX;
                y = e.pageY;
            }

            if (this.moving) {
                const trueX = x - bounds.left;
                const trueY = y - bounds.top;

                const diffX = this.moveStartX - trueX;
                this.x = Math.max(0, this.moveOriginalX - diffX);

                const diffY = this.moveStartY - trueY;
                this.y = Math.max(0, this.moveOriginalY - diffY);

                this.fixPosition();
            }
            if (this.resizing) {
                const trueX = x - bounds.left;
                const trueY = y - bounds.top;

                const diffX = this.resizeStartX - trueX;
                const diffY = this.resizeStartY - trueY;

                let x1 = this.resizeOriginalX;
                let y1 = this.resizeOriginalY;
                let x2 = x1 + this.resizeOriginalW;
                let y2 = y1 + this.resizeOriginalH;

                switch (this.resizeMode) {
                    case "t":
                        y1 -= diffY;
                        break;
                    case "b":
                        y2 -= diffY;
                        break;
                    case "l":
                        x1 -= diffX;
                        break;
                    case "r":
                        x2 -= diffX;
                        break;
                    case "tl":
                        y1 -= diffY;
                        x1 -= diffX;
                        break;
                    case "tr":
                        y1 -= diffY;
                        x2 -= diffX;
                        break;
                    case "bl":
                        y2 -= diffY;
                        x1 -= diffX;
                        break;
                    case "br":
                        y2 -= diffY;
                        x2 -= diffX;
                        break;
                }

                x1 = Math.min(bounds.width, Math.max(0, x1));
                x2 = Math.min(bounds.width, Math.max(0, x2));

                y1 = Math.min(bounds.height, Math.max(0, y1));
                y2 = Math.min(bounds.height, Math.max(0, y2));

                this.x = Math.min(x1, x2);
                this.y = Math.min(y1, y2);

                this.width = Math.max(MIN_WIDTH, Math.abs(x1 - x2));
                this.height = Math.max(MIN_HEIGHT, Math.abs(y1 - y2));

                this.fixPosition();
            }
        },

        mouseDrop: function () {
            if (!this.moving && !this.resizing) {
                return;
            }

            if (this.moving) {
                this.moving = false;
                this.savePosition();
            }
            if (this.resizing) {
                this.resizing = false;
                this.savePosition();
            }
        },
    },

    mounted: function () {
        this._handles = Object.create(null);

        this._handles.fixPositionTimer = setInterval(this.fixPosition.bind(this), 1000);

        this._handles.mouseMoveH = this.mouseMove.bind(this);

        document.addEventListener("mousemove", this._handles.mouseMoveH);
        document.addEventListener("touchmove", this._handles.mouseMoveH);

        this._handles.mouseDropH = this.mouseDrop.bind(this);
        document.addEventListener("mouseup", this._handles.mouseDropH);
        document.addEventListener("touchend", this._handles.mouseDropH);

        nextTick(() => {
            this.loadPosition();
        });
    },

    beforeUnmount: function () {
        clearInterval(this._handles.fixPositionTimer);

        document.removeEventListener("mouseup", this._handles.mouseDropH);
        document.removeEventListener("touchend", this._handles.mouseDropH);
        document.removeEventListener("mousemove", this._handles.mouseMoveH);
        document.removeEventListener("touchmove", this._handles.mouseMoveH);
    },

    watch: {
        positionKey: function () {
            this.loadPosition();
        },
    },
});
</script>

<style>
@import "@/style/player/resizable-widget.css";
</style>
