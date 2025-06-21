<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        @touchend="onTouchEnd"
        @mouseup="onMouseUp"
        @mouseleave="onMouseLeave"
    >
        <div v-if="display" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Crop thumbnail before applying it") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div
                    class="thumbnail-crop-container"
                    :style="{ height: loading ? '400px' : imgHeight + 'px' }"
                    @mouseup="onMouseUp"
                    @touchend="onTouchEnd"
                >
                    <LoadingOverlay v-if="loading"></LoadingOverlay>
                    <img
                        v-if="!loading"
                        class="thumbnail-crop-image"
                        :style="{ top: imgTop + 'px', left: imgLeft + 'px', width: imgWidth + 'px', height: imgHeight + 'px' }"
                        :src="imageUrl"
                    />
                    <div
                        v-if="!loading && bg1Display"
                        class="thumbnail-crop-background"
                        :style="{ top: bg1Top + 'px', left: bg1Left + 'px', width: bg1Width + 'px', height: bg1Height + 'px' }"
                    ></div>
                    <div
                        v-if="!loading && bg2Display"
                        class="thumbnail-crop-background"
                        :style="{ top: bg2Top + 'px', left: bg2Left + 'px', width: bg2Width + 'px', height: bg2Height + 'px' }"
                    ></div>
                    <div
                        v-if="!loading"
                        class="thumbnail-crop-section"
                        :style="{ top: cropTop + 'px', left: cropLeft + 'px', width: cropSize + 'px', height: cropSize + 'px' }"
                        @mousedown="onMouseDown"
                        @touchstart="onTouchStart"
                    ></div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" :disabled="busy" @click="done">
                    <i class="fas fa-check"></i> {{ $t("Done") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import LoadingOverlay from "../layout/LoadingOverlay.vue";

export default defineComponent({
    name: "ThumbnailCropModal",
    components: {
        LoadingOverlay,
    },
    props: {
        display: Boolean,
        imageUrl: String,
    },
    emits: ["update:display", "done", "error"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            resizeObserver: null as ResizeObserver,
            tempImage: null as HTMLImageElement,
        };
    },
    data: function () {
        return {
            closeSignal: 0,

            loading: true,

            originalImgWidth: 0,
            originalImgHeight: 0,

            busy: false,

            imgTop: 0,
            imgLeft: 0,
            imgWidth: 0,
            imgHeight: 0,

            bg1Display: false,
            bg1Top: 0,
            bg1Left: 0,
            bg1Width: 0,
            bg1Height: 0,

            bg2Display: false,
            bg2Top: 0,
            bg2Left: 0,
            bg2Width: 0,
            bg2Height: 0,

            cropTop: 0,
            cropLeft: 0,
            cropSize: 0,

            vertical: false,

            moving: false,
            movingStartX: 0,
            movingStartY: 0,
            movingStartPos: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
                this.load();
            } else {
                this.stopResizeObserver();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.autoFocus();
            this.load();
        }

        this.$listenOnDocumentEvent("mousemove", this.onMouseMove.bind(this));
        this.$listenOnDocumentEvent("mouseup", this.onMouseMove.bind(this));

        this.$listenOnDocumentEvent("touchmove", this.onTouchMove.bind(this));
        this.$listenOnDocumentEvent("touchend", this.onTouchEnd.bind(this));
    },
    beforeUnmount: function () {
        if (this.resizeObserver) {
            this.resizeObserver.disconnect();
        }

        if (this.tempImage) {
            delete this.tempImage.onload;
            delete this.tempImage.onerror;
        }
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        stopResizeObserver: function () {
            if (this.resizeObserver) {
                this.resizeObserver.disconnect();
                this.resizeObserver = null;
            }
        },

        startResizeObserver: function () {
            this.stopResizeObserver();
            this.resizeObserver = new ResizeObserver(this.onResize.bind(this));

            const element = this.$el.querySelector(".thumbnail-crop-container");

            if (element) {
                this.resizeObserver.observe(element);
            }
        },

        load: function () {
            if (this.tempImage) {
                delete this.tempImage.onload;
                delete this.tempImage.onerror;
                this.tempImage = null;
            }

            this.loading = true;

            const img = new Image();
            this.tempImage = img;

            img.onload = () => {
                this.tempImage = null;
                this.onImageLoad(img);
            };

            img.onerror = (err) => {
                this.tempImage = null;
                console.error(err);
                this.onImageError();
            };

            img.src = this.imageUrl;
        },

        onImageLoad: function (img: HTMLImageElement) {
            this.loading = false;

            this.originalImgWidth = img.width;
            this.originalImgHeight = img.height;

            this.vertical = this.originalImgHeight > this.originalImgWidth;

            this.onResize();
            this.startResizeObserver();
        },

        onResize: function () {
            if (this.loading || !this.display) {
                return; // Still loading or hidden
            }

            const container = this.$el.querySelector(".thumbnail-crop-container") as HTMLElement;

            if (!container) {
                return;
            }

            const containerBounds = container.getBoundingClientRect();

            const containerSize = containerBounds.width || 1; // Container must be square

            if (!this.vertical) {
                // Horizontal

                this.imgLeft = 0;
                this.imgWidth = containerSize;
                this.imgHeight = Math.round(this.originalImgHeight * (containerSize / this.originalImgWidth));
                this.imgTop = 0;

                this.cropSize = this.imgHeight;
                this.cropTop = this.imgTop;

                this.cropLeft = Math.round((this.imgWidth - this.cropSize) / 2);

                this.bg1Top = this.imgTop;
                this.bg1Left = 0;
                this.bg1Height = this.imgHeight;
                this.bg1Width = this.cropLeft;
                this.bg1Display = this.bg1Width > 0;

                this.bg2Top = this.imgTop;
                this.bg2Left = this.cropLeft + this.cropSize;
                this.bg2Height = this.imgHeight;
                this.bg2Width = this.imgWidth - this.bg2Left;
                this.bg2Display = this.bg2Width > 0;
            } else {
                // Vertical

                this.imgTop = 0;
                this.imgHeight = containerSize;
                this.imgWidth = Math.round(this.originalImgWidth * (containerSize / this.originalImgHeight));
                this.imgLeft = Math.round((containerSize - this.imgWidth) / 2);

                this.cropSize = this.imgWidth;
                this.cropLeft = this.imgLeft;

                this.cropTop = Math.round((this.imgHeight - this.cropSize) / 2);

                this.bg1Top = 0;
                this.bg1Left = this.imgLeft;
                this.bg1Height = this.cropTop;
                this.bg1Width = this.imgWidth;
                this.bg1Display = this.bg1Height > 0;

                this.bg2Top = this.cropTop + this.cropSize;
                this.bg2Left = this.imgLeft;
                this.bg2Height = this.imgHeight - this.bg2Top;
                this.bg2Width = this.imgWidth;
                this.bg2Display = this.bg2Height > 0;
            }
        },

        onImageError: function () {
            this.$emit("error");
            this.close();
        },

        done: function () {
            const imgElement = this.$el.querySelector(".thumbnail-crop-image") as HTMLImageElement;

            if (!imgElement || !imgElement.complete) {
                return;
            }

            const imageScaleW = this.originalImgWidth / (this.imgWidth || 1);
            const imageScaleH = this.originalImgHeight / (this.imgHeight || 1);

            const cropSizeScaled = Math.round(this.cropSize * (this.vertical ? imageScaleH : imageScaleW));

            this.busy = true;

            try {
                // Create canvas
                const canvas = document.createElement("canvas") as HTMLCanvasElement;

                canvas.width = cropSizeScaled;
                canvas.height = cropSizeScaled;

                //  Draw video frame to the canvas
                const ctx = canvas.getContext("2d");

                if (this.vertical) {
                    ctx.drawImage(
                        imgElement,
                        // Image
                        0,
                        Math.round((this.cropTop - this.imgTop) * imageScaleH),
                        Math.round(this.cropSize * imageScaleH),
                        Math.round(this.cropSize * imageScaleH),
                        // Canvas
                        0,
                        0,
                        cropSizeScaled,
                        cropSizeScaled,
                    );
                } else {
                    ctx.drawImage(
                        imgElement,
                        // Image
                        Math.round((this.cropLeft - this.imgLeft) * imageScaleW),
                        0,
                        Math.round(this.cropSize * imageScaleW),
                        Math.round(this.cropSize * imageScaleW),
                        // Canvas
                        0,
                        0,
                        cropSizeScaled,
                        cropSizeScaled,
                    );
                }

                // Get frame as blob
                canvas.toBlob((blob) => {
                    // Convert to file
                    const file = new File([blob], "thumbnail.png");

                    console.log(URL.createObjectURL(file));

                    // Change thumbnail
                    this.$emit("done", file);
                }, "image/png");
            } catch (ex) {
                this.busy = false;
                console.error(ex);
                this.onImageError();
                return;
            }

            this.busy = false;

            this.close();
        },

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        startMove: function (x: number, y: number) {
            this.moving = true;
            this.movingStartX = x;
            this.movingStartY = y;
            this.movingStartPos = this.vertical ? this.cropTop : this.cropLeft;
        },

        move: function (x: number, y: number) {
            if (!this.moving) {
                return;
            }

            if (this.vertical) {
                const yDiff = y - this.movingStartY;
                this.cropTop = Math.max(0, Math.min(this.imgHeight - this.cropSize, this.movingStartPos + yDiff));

                this.bg1Height = this.cropTop;
                this.bg1Display = this.bg1Height > 0;

                this.bg2Top = this.cropTop + this.cropSize;
                this.bg2Height = this.imgHeight - this.bg2Top;
                this.bg2Display = this.bg2Height > 0;
            } else {
                const xDiff = x - this.movingStartX;
                this.cropLeft = Math.max(0, Math.min(this.imgWidth - this.cropSize, this.movingStartPos + xDiff));

                this.bg1Width = this.cropLeft;
                this.bg1Display = this.bg1Width > 0;

                this.bg2Left = this.cropLeft + this.cropSize;
                this.bg2Width = this.imgWidth - this.bg2Left;
                this.bg2Display = this.bg2Width > 0;
            }
        },

        endMove: function () {
            this.moving = false;
        },

        onMouseDown: function (e: MouseEvent) {
            this.startMove(e.pageX, e.pageY);
        },

        onMouseMove: function (e: MouseEvent) {
            this.move(e.pageX, e.pageY);
        },

        onMouseUp: function (e: MouseEvent) {
            e.stopPropagation();
            this.move(e.pageX, e.pageY);
            this.endMove();
        },

        onMouseLeave: function () {
            this.endMove();
        },

        onTouchStart: function (e: TouchEvent) {
            if (e.touches.length > 0 && typeof e.touches[0].pageX === "number" && typeof e.touches[0].pageY === "number") {
                this.startMove(e.touches[0].pageX, e.touches[0].pageY);
            }
        },

        onTouchMove: function (e: TouchEvent) {
            if (e.touches.length > 0 && typeof e.touches[0].pageX === "number" && typeof e.touches[0].pageY === "number") {
                this.move(e.touches[0].pageX, e.touches[0].pageY);
            }
        },

        onTouchEnd: function (e: TouchEvent) {
            e.stopPropagation();
            if (e.touches.length > 0 && typeof e.touches[0].pageX === "number" && typeof e.touches[0].pageY === "number") {
                this.move(e.touches[0].pageX, e.touches[0].pageY);
            }
            this.endMove();
        },
    },
});
</script>
