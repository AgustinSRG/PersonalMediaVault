<template>
    <ModalDialogContainer ref="container" v-model:display="display" @touchend="onTouchEnd" @mouseup="onMouseUp" @mouseleave="onMouseLeave">
        <div ref="dialog" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Crop thumbnail before applying it") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div
                    class="thumbnail-crop-container"
                    :style="{ height: loading ? '400px' : coordinates.imgHeight + 'px' }"
                    @mouseup="onMouseUp"
                    @touchend="onTouchEnd"
                >
                    <LoadingOverlay v-if="loading"></LoadingOverlay>
                    <img
                        v-if="!loading"
                        class="thumbnail-crop-image"
                        :style="{
                            top: coordinates.imgTop + 'px',
                            left: coordinates.imgLeft + 'px',
                            width: coordinates.imgWidth + 'px',
                            height: coordinates.imgHeight + 'px',
                        }"
                        :src="imageUrl"
                    />
                    <div
                        v-if="!loading && coordinates.bg1Display"
                        class="thumbnail-crop-background"
                        :style="{
                            top: coordinates.bg1Top + 'px',
                            left: coordinates.bg1Left + 'px',
                            width: coordinates.bg1Width + 'px',
                            height: coordinates.bg1Height + 'px',
                        }"
                    ></div>
                    <div
                        v-if="!loading && coordinates.bg2Display"
                        class="thumbnail-crop-background"
                        :style="{
                            top: coordinates.bg2Top + 'px',
                            left: coordinates.bg2Left + 'px',
                            width: coordinates.bg2Width + 'px',
                            height: coordinates.bg2Height + 'px',
                        }"
                    ></div>
                    <div
                        v-if="!loading"
                        class="thumbnail-crop-section"
                        :style="{
                            top: coordinates.cropTop + 'px',
                            left: coordinates.cropLeft + 'px',
                            width: coordinates.cropSize + 'px',
                            height: coordinates.cropSize + 'px',
                        }"
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

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { onBeforeUnmount, onMounted, reactive, ref, useTemplateRef, watch } from "vue";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { onDocumentEvent } from "@/composables/on-document-event";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * Image URL
     */
    imageUrl: {
        type: String,
        required: true,
    },
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when done. The cropped image file is passed.
     */
    (e: "done", file: File): void;

    /**
     * Emitted when an error happens.
     * This indicates the image file may be invalid.
     */
    (e: "error"): void;
}>();

// Dialog element
const dialog = useTemplateRef("dialog");

// Resize observer
let resizeObserver: ResizeObserver | null = null;

/**
 * Starts resize observer
 */
const startResizeObserver = () => {
    stopResizeObserver();
    resizeObserver = new ResizeObserver(onResize);

    const element = dialog.value?.querySelector(".thumbnail-crop-container") as HTMLElement;

    if (element) {
        resizeObserver.observe(element);
    }
};

/**
 * Stops resize observer
 */
const stopResizeObserver = () => {
    if (resizeObserver) {
        resizeObserver.disconnect();
        resizeObserver = null;
    }
};

onBeforeUnmount(stopResizeObserver);

// Temp image element used to load the URL
let tempImage: HTMLImageElement | null = null;

onBeforeUnmount(() => {
    if (tempImage) {
        delete tempImage.onload;
        delete tempImage.onerror;
    }
});

// Loading status
const loading = ref(true);

// Original image with
const originalImgWidth = ref(0);

// Original image height
const originalImgHeight = ref(0);

// Busy status
const busy = ref(false);

// Moving the crop section
const moving = ref(false);

// Coordinates
const coordinates = reactive({
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

    movingStartX: 0,
    movingStartY: 0,
    movingStartPos: 0,
});

/**
 * Called when the container element is resized
 */
const onResize = () => {
    if (loading.value || !display.value) {
        return; // Still loading or hidden
    }

    const container = dialog.value?.querySelector(".thumbnail-crop-container") as HTMLElement;

    if (!container) {
        return;
    }

    const containerBounds = container.getBoundingClientRect();

    const containerSize = containerBounds.width || 1; // Container must be square

    if (!coordinates.vertical) {
        // Horizontal

        coordinates.imgLeft = 0;
        coordinates.imgWidth = containerSize;
        coordinates.imgHeight = Math.round(originalImgHeight.value * (containerSize / originalImgWidth.value));
        coordinates.imgTop = 0;

        coordinates.cropSize = coordinates.imgHeight;
        coordinates.cropTop = coordinates.imgTop;

        coordinates.cropLeft = Math.round((coordinates.imgWidth - coordinates.cropSize) / 2);

        coordinates.bg1Top = coordinates.imgTop;
        coordinates.bg1Left = 0;
        coordinates.bg1Height = coordinates.imgHeight;
        coordinates.bg1Width = coordinates.cropLeft;
        coordinates.bg1Display = coordinates.bg1Width > 0;

        coordinates.bg2Top = coordinates.imgTop;
        coordinates.bg2Left = coordinates.cropLeft + coordinates.cropSize;
        coordinates.bg2Height = coordinates.imgHeight;
        coordinates.bg2Width = coordinates.imgWidth - coordinates.bg2Left;
        coordinates.bg2Display = coordinates.bg2Width > 0;
    } else {
        // Vertical

        coordinates.imgTop = 0;
        coordinates.imgHeight = containerSize;
        coordinates.imgWidth = Math.round(originalImgWidth.value * (containerSize / originalImgHeight.value));
        coordinates.imgLeft = Math.round((containerSize - coordinates.imgWidth) / 2);

        coordinates.cropSize = coordinates.imgWidth;
        coordinates.cropLeft = coordinates.imgLeft;

        coordinates.cropTop = Math.round((coordinates.imgHeight - coordinates.cropSize) / 2);

        coordinates.bg1Top = 0;
        coordinates.bg1Left = coordinates.imgLeft;
        coordinates.bg1Height = coordinates.cropTop;
        coordinates.bg1Width = coordinates.imgWidth;
        coordinates.bg1Display = coordinates.bg1Height > 0;

        coordinates.bg2Top = coordinates.cropTop + coordinates.cropSize;
        coordinates.bg2Left = coordinates.imgLeft;
        coordinates.bg2Height = coordinates.imgHeight - coordinates.bg2Top;
        coordinates.bg2Width = coordinates.imgWidth;
        coordinates.bg2Display = coordinates.bg2Height > 0;
    }
};

/**
 * Loads the image
 */
const load = () => {
    if (tempImage) {
        delete tempImage.onload;
        delete tempImage.onerror;
        tempImage = null;
    }

    loading.value = true;

    const img = new Image();
    tempImage = img;

    img.onload = () => {
        tempImage = null;
        onImageLoad(img);
    };

    img.onerror = (err) => {
        tempImage = null;
        console.error(err);
        onImageError();
    };

    img.src = props.imageUrl;
};

onMounted(() => {
    if (display.value) {
        load();
    } else {
        stopResizeObserver();
    }
});

watch(display, () => {
    if (display.value) {
        load();
    }
});

/**
 * Called when the image is loaded successfully
 * @param img The image element
 */
const onImageLoad = (img: HTMLImageElement) => {
    loading.value = false;

    originalImgWidth.value = img.width;
    originalImgHeight.value = img.height;

    coordinates.vertical = originalImgHeight.value > originalImgWidth.value;

    onResize();
    startResizeObserver();
};

/**
 * Called when the image cannot be loaded
 */
const onImageError = () => {
    emit("error");
    close();
};

/**
 * Called when the user is done cropping the image
 */
const done = () => {
    const imgElement = dialog.value?.querySelector(".thumbnail-crop-image") as HTMLImageElement;

    if (!imgElement || !imgElement.complete) {
        return;
    }

    const imageScaleW = originalImgWidth.value / (coordinates.imgWidth || 1);
    const imageScaleH = originalImgHeight.value / (coordinates.imgHeight || 1);

    const cropSizeScaled = Math.round(coordinates.cropSize * (coordinates.vertical ? imageScaleH : imageScaleW));

    busy.value = true;

    try {
        // Create canvas
        const canvas = document.createElement("canvas") as HTMLCanvasElement;

        canvas.width = cropSizeScaled;
        canvas.height = cropSizeScaled;

        //  Draw video frame to the canvas
        const ctx = canvas.getContext("2d");

        if (coordinates.vertical) {
            ctx.drawImage(
                imgElement,
                // Image
                0,
                Math.round((coordinates.cropTop - coordinates.imgTop) * imageScaleH),
                Math.round(coordinates.cropSize * imageScaleH),
                Math.round(coordinates.cropSize * imageScaleH),
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
                Math.round((coordinates.cropLeft - coordinates.imgLeft) * imageScaleW),
                0,
                Math.round(coordinates.cropSize * imageScaleW),
                Math.round(coordinates.cropSize * imageScaleW),
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
            emit("done", file);
        }, "image/png");
    } catch (ex) {
        busy.value = false;
        console.error(ex);
        onImageError();
        return;
    }

    busy.value = false;

    close();
};

/**
 * Starts moving the clop section
 * @param x X coordinate
 * @param y Y coordinate
 */
const startMove = (x: number, y: number) => {
    moving.value = true;
    coordinates.movingStartX = x;
    coordinates.movingStartY = y;
    coordinates.movingStartPos = coordinates.vertical ? coordinates.cropTop : coordinates.cropLeft;
};

/**
 * Moves the crop section
 * @param x X coordinate
 * @param y Y coordinate
 */
const move = (x: number, y: number) => {
    if (!moving.value) {
        return;
    }

    if (coordinates.vertical) {
        const yDiff = y - coordinates.movingStartY;
        coordinates.cropTop = Math.max(0, Math.min(coordinates.imgHeight - coordinates.cropSize, coordinates.movingStartPos + yDiff));

        coordinates.bg1Height = coordinates.cropTop;
        coordinates.bg1Display = coordinates.bg1Height > 0;

        coordinates.bg2Top = coordinates.cropTop + coordinates.cropSize;
        coordinates.bg2Height = coordinates.imgHeight - coordinates.bg2Top;
        coordinates.bg2Display = coordinates.bg2Height > 0;
    } else {
        const xDiff = x - coordinates.movingStartX;
        coordinates.cropLeft = Math.max(0, Math.min(coordinates.imgWidth - coordinates.cropSize, coordinates.movingStartPos + xDiff));

        coordinates.bg1Width = coordinates.cropLeft;
        coordinates.bg1Display = coordinates.bg1Width > 0;

        coordinates.bg2Left = coordinates.cropLeft + coordinates.cropSize;
        coordinates.bg2Width = coordinates.imgWidth - coordinates.bg2Left;
        coordinates.bg2Display = coordinates.bg2Width > 0;
    }
};

/**
 * Finalizes moving the crop section
 */
const endMove = () => {
    moving.value = false;
};

/**
 * Event handler for 'mousedown'
 * @param e The mouse event
 */
const onMouseDown = (e: MouseEvent) => {
    startMove(e.pageX, e.pageY);
};

/**
 * Event handler for 'mouseup'
 * @param e The mouse event
 */
const onMouseUp = (e: MouseEvent) => {
    e.stopPropagation();
    move(e.pageX, e.pageY);
    endMove();
};

onDocumentEvent("mousemove", (e) => {
    move(e.pageX, e.pageY);
});

onDocumentEvent("mouseup", onMouseUp);

/**
 * Event handler for 'mouseleave'
 */
const onMouseLeave = () => {
    endMove();
};

/**
 * Event handler for 'touchstart'
 * @param e The touch event
 */
const onTouchStart = (e: TouchEvent) => {
    if (e.touches.length > 0 && typeof e.touches[0].pageX === "number" && typeof e.touches[0].pageY === "number") {
        startMove(e.touches[0].pageX, e.touches[0].pageY);
    }
};

onDocumentEvent("touchmove", (e) => {
    if (e.touches.length > 0 && typeof e.touches[0].pageX === "number" && typeof e.touches[0].pageY === "number") {
        move(e.touches[0].pageX, e.touches[0].pageY);
    }
});

const onTouchEnd = (e: TouchEvent) => {
    e.stopPropagation();
    if (e.touches.length > 0 && typeof e.touches[0].pageX === "number" && typeof e.touches[0].pageY === "number") {
        move(e.touches[0].pageX, e.touches[0].pageY);
    }
    endMove();
};

onDocumentEvent("touchend", onTouchEnd);
</script>
