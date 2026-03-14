<template>
    <div
        ref="container"
        class="resizable-widget"
        :class="{ hidden: !display }"
        tabindex="-1"
        :style="{
            top: y + 'px',
            left: x + 'px',
            width: width + 'px',
            height: height + 'px',
        }"
        @keydown="onKeyDown"
        @dblclick="stopPropagationEvent"
        @mousedown="propagateClick"
        @touchstart="propagateTouch"
        @contextmenu="stopPropagationEvent"
    >
        <div class="resizable-widget-header" @mousedown="startMovingMouse" @touchstart.passive="startMovingTouch">
            <div class="resizable-widget-title" :title="title">{{ title }}</div>
            <div v-if="actionButtons && actionButtons.length > 0" class="resizable-widget-action-buttons">
                <button
                    v-for="btn in actionButtons"
                    :key="btn.id"
                    type="button"
                    :disabled="busy"
                    class="action-button"
                    :title="btn.name"
                    @click="doActionButton(btn.id)"
                >
                    <i :class="btn.icon"></i>
                </button>
            </div>
            <div class="resizable-widget-close-btn">
                <button type="button" :disabled="busy" class="close-button" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        </div>

        <div class="resizable-widget-body" :class="{ resizing: resizing }">
            <slot></slot>
        </div>

        <div
            class="resize resize-left"
            @mousedown="startResizingMouse($event, 'l')"
            @touchstart.passive="startResizingTouch($event, 'l')"
        ></div>
        <div
            class="resize resize-top"
            @mousedown="startResizingMouse($event, 't')"
            @touchstart.passive="startResizingTouch($event, 't')"
        ></div>
        <div
            class="resize resize-right"
            @mousedown="startResizingMouse($event, 'r')"
            @touchstart.passive="startResizingTouch($event, 'r')"
        ></div>
        <div
            class="resize resize-bottom"
            @mousedown="startResizingMouse($event, 'b')"
            @touchstart.passive="startResizingTouch($event, 'b')"
        ></div>
        <div
            class="resize resize-corner-top-left"
            @mousedown="startResizingMouse($event, 'tl')"
            @touchstart.passive="startResizingTouch($event, 'tl')"
        ></div>
        <div
            class="resize resize-corner-top-right"
            @mousedown="startResizingMouse($event, 'tr')"
            @touchstart.passive="startResizingTouch($event, 'tr')"
        ></div>
        <div
            class="resize resize-corner-bottom-left"
            @mousedown="startResizingMouse($event, 'bl')"
            @touchstart.passive="startResizingTouch($event, 'bl')"
        ></div>
        <div
            class="resize resize-corner-bottom-right"
            @mousedown="startResizingMouse($event, 'br')"
            @touchstart.passive="startResizingTouch($event, 'br')"
        ></div>
    </div>
</template>

<script setup lang="ts">
import { onDocumentEvent } from "@/composables/on-document-event";
import { useI18n } from "@/composables/use-i18n";
import { useInterval } from "@/composables/use-interval";
import { fetchFromLocalStorage, saveIntoLocalStorage } from "@/local-storage/local-storage";
import type { PositionEvent } from "@/utils/position-event";
import { positionEventFromMouseEvent, positionEventFromTouchEvent } from "@/utils/position-event";
import type { WidgetActionButton } from "@/utils/widgets";
import type { PropType } from "vue";
import { nextTick, onMounted, ref, useTemplateRef, watch } from "vue";
import { stopPropagationEvent } from "@/utils/events";

// Initial widget width
const INITIAL_WIDTH = 480;

// Initial widget height
const INITIAL_HEIGHT = 360;

// Min widget width
const MIN_WIDTH = 250;

// Min widget width
const MIN_HEIGHT = 250;

// Translation
const { $t } = useI18n();

// Ref to the container element
const container = useTemplateRef("container");

// Display model
const display = defineModel<boolean>("display");

// Props
const props = defineProps({
    /**
     * Local storage key to store the position
     */
    positionKey: {
        type: String,
        required: true,
    },

    /**
     * Widget title
     */
    title: {
        type: String,
        required: true,
    },

    actionButtons: {
        type: Array as PropType<WidgetActionButton[]>,
        default: () => [],
    },

    /**
     * True if a context menu is opened
     */
    contextOpen: Boolean,

    /**
     * True if busy (disable action buttons)
     */
    busy: Boolean,
});

// Emits
const emit = defineEmits<{
    /**
     * The widget was clicked
     */
    (e: "clicked"): void;

    /**
     * An action button was clicked
     */
    (e: "action-btn", id: string): void;
}>();

// X coordinate
const x = ref(0);

// Y coordinate
const y = ref(0);

// Widget width
const width = ref(INITIAL_WIDTH);

// Widget height
const height = ref(INITIAL_HEIGHT);

// Moving widget?
const moving = ref(false);

// Moving coordinates
const moveOriginalX = ref(0);
const moveOriginalY = ref(0);
const moveStartX = ref(0);
const moveStartY = ref(0);

// Resizing widget
const resizing = ref(false);

// Resizing coordinates
const resizeOriginalX = ref(0);
const resizeOriginalY = ref(0);
const resizeOriginalW = ref(0);
const resizeOriginalH = ref(0);
const resizeStartX = ref(0);
const resizeStartY = ref(0);

// Resize modes. Depending in from where is being resized
type ResizeMode = "" | "t" | "b" | "l" | "r" | "tl" | "tr" | "bl" | "br";

// Resize mode
const resizeMode = ref<ResizeMode>("");

/**
 * Closes the widget
 */
const close = () => {
    display.value = false;
};

/**
 * Propagates the click event to the parent component
 * @param e The event
 */
const propagateClick = (e: MouseEvent) => {
    e.stopPropagation();
    if (e.button !== 0) {
        return;
    }
    emit("clicked");
};

/**
 * Propagates the touch event to the parent component
 * @param e The event
 */
const propagateTouch = (e: TouchEvent) => {
    e.stopPropagation();
    emit("clicked");
};

/**
 * Fixes the position based on the bounds of the parent element
 */
const fixPosition = () => {
    const parentElem: HTMLElement = container.value?.parentElement;

    if (parentElem) {
        const bounds = parentElem.getBoundingClientRect();

        if (bounds.width <= 0 || bounds.height <= 0) {
            return;
        }

        // X

        if (x.value < 0) {
            x.value = 0;
        }

        if (width.value < MIN_WIDTH) {
            width.value = MIN_WIDTH;
        }

        if (x.value >= bounds.width) {
            x.value = bounds.width - 1;
        }

        if (x.value + width.value >= bounds.width) {
            // Try move to the left
            x.value = bounds.width - width.value - 1;
            if (x.value < 0) {
                // Does not fit
                width.value = bounds.width;
                x.value = 0;
            }
        }

        // Y

        if (y.value < 0) {
            y.value = 0;
        }

        if (height.value < MIN_HEIGHT) {
            height.value = MIN_HEIGHT;
        }

        if (y.value >= bounds.height) {
            y.value = bounds.height - 1;
        }

        if (y.value + height.value >= bounds.height) {
            // Try move to the left
            y.value = bounds.height - height.value - 1;
            if (y.value < 0) {
                // Does not fit
                height.value = bounds.height;
                y.value = 0;
            }
        }
    }
};

/**
 * Loads the widget position from local storage
 */
const loadPosition = () => {
    // Load position
    if (props.positionKey) {
        const savedPosition = fetchFromLocalStorage(props.positionKey, null);

        if (savedPosition && typeof savedPosition === "object") {
            x.value = Number(savedPosition.x) || 0;
            y.value = Number(savedPosition.y) || 0;
            width.value = Number(savedPosition.width) || 0;
            height.value = Number(savedPosition.height) || 0;
        } else {
            // Center with initial size
            const parentElem: HTMLElement = container.value?.parentElement;

            if (parentElem) {
                const bounds = parentElem.getBoundingClientRect();

                if (bounds.width < INITIAL_WIDTH) {
                    x.value = 0;
                    width.value = bounds.width;
                } else {
                    x.value = Math.floor((bounds.width - INITIAL_WIDTH) / 2);
                    width.value = INITIAL_WIDTH;
                }

                if (bounds.height < INITIAL_HEIGHT) {
                    y.value = 0;
                    height.value = bounds.height;
                } else {
                    y.value = Math.floor((bounds.height - INITIAL_HEIGHT) / 2);
                    height.value = INITIAL_HEIGHT;
                }
            }
        }
    }

    fixPosition();
};

// Timer to periodically fix the widget position
const fixPositionTimer = useInterval();

// Delay for interval to periodically fix the position
const POSITION_FIX_DELAY_INTERVAL = 100;

onMounted(() => {
    fixPositionTimer.set(fixPosition, POSITION_FIX_DELAY_INTERVAL);

    nextTick(loadPosition);
});

watch(() => props.positionKey, loadPosition);

/**
 * Saves the current position of the widget in local storage
 */
const savePosition = () => {
    if (props.positionKey) {
        saveIntoLocalStorage(props.positionKey, {
            x: x.value,
            y: y.value,
            width: width.value,
            height: height.value,
        });
    }
};

/**
 * Indicates the parent component an action button was triggered
 * @param id The button ID
 */
const doActionButton = (id: string) => {
    emit("action-btn", id);
};

/**
 * Handler for 'keydown' event
 * @param e The keyboard event
 */
const onKeyDown = (e: KeyboardEvent) => {
    e.stopPropagation();

    if (e.key === "Escape") {
        close();
        return;
    }

    if (props.actionButtons) {
        for (const btn of props.actionButtons) {
            if (!btn.key) {
                continue;
            }

            if ((typeof btn.key === "string" && btn.key === e.key) || (Array.isArray(btn.key) && btn.key.includes(e.key))) {
                e.preventDefault();
                doActionButton(btn.id);
                return;
            }
        }
    }
};

/**
 * Starts moving the widget
 * @param e The position event
 */
const startMoving = (e: PositionEvent) => {
    if (props.contextOpen) {
        return;
    }

    if (moving.value || resizing.value) {
        return;
    }

    e.e.stopPropagation();

    const parentElem: HTMLElement = container.value?.parentElement;

    if (!parentElem) {
        return;
    }

    const bounds = parentElem.getBoundingClientRect();

    const trueX = e.x - bounds.left;
    const trueY = e.y - bounds.top;

    moving.value = true;
    moveStartX.value = trueX;
    moveStartY.value = trueY;
    moveOriginalX.value = x.value;
    moveOriginalY.value = y.value;
};

/**
 * Starts moving widget using the mouse
 * @param e The mouse event
 */
const startMovingMouse = (e: MouseEvent) => {
    if (e.button !== 0) {
        return;
    }

    startMoving(positionEventFromMouseEvent(e));
};

/**
 * Starts moving the widget using the touch screen
 * @param e The touch event
 */
const startMovingTouch = (e: TouchEvent) => {
    startMoving(positionEventFromTouchEvent(e));
};

/**
 * Starts resizing the widget
 * @param e The position event
 * @param rm The resize mode
 */
const startResizing = (e: PositionEvent, rm: ResizeMode) => {
    if (props.contextOpen) {
        return;
    }

    if (moving.value || resizing.value) {
        return;
    }

    e.e.stopPropagation();

    const parentElem: HTMLElement = container.value?.parentElement;

    if (!parentElem) {
        return;
    }

    const bounds = parentElem.getBoundingClientRect();

    resizing.value = true;
    resizeMode.value = rm;
    resizeOriginalX.value = x.value;
    resizeOriginalY.value = y.value;
    resizeOriginalW.value = width.value;
    resizeOriginalH.value = height.value;

    resizeStartX.value = e.x - bounds.left;
    resizeStartY.value = e.y - bounds.top;
};

/**
 * Starts moving widget using the mouse
 * @param e The mouse event
 * @param resizeMode The resize mode
 */
const startResizingMouse = (e: MouseEvent, resizeMode: ResizeMode) => {
    if (e.button !== 0) {
        return;
    }

    startResizing(positionEventFromMouseEvent(e), resizeMode);
};

/**
 * Starts moving the widget using the touch screen
 * @param e The touch event
 * @param resizeMode The resize mode
 */
const startResizingTouch = (e: TouchEvent, resizeMode: ResizeMode) => {
    startResizing(positionEventFromTouchEvent(e), resizeMode);
};

/**
 * Called when the position moves
 * @param e The position event
 */
const move = (e: PositionEvent) => {
    if (!moving.value && !resizing.value) {
        return;
    }

    const parentElem: HTMLElement = container.value?.parentElement;

    if (!parentElem) {
        return;
    }

    const bounds = parentElem.getBoundingClientRect();

    const trueX = e.x - bounds.left;
    const trueY = e.y - bounds.top;

    if (moving.value) {
        const diffX = moveStartX.value - trueX;
        x.value = Math.max(0, moveOriginalX.value - diffX);

        const diffY = moveStartY.value - trueY;
        y.value = Math.max(0, moveOriginalY.value - diffY);

        fixPosition();
    }

    if (resizing.value) {
        const diffX = resizeStartX.value - trueX;
        const diffY = resizeStartY.value - trueY;

        let x1 = resizeOriginalX.value;
        let y1 = resizeOriginalY.value;
        let x2 = x1 + resizeOriginalW.value;
        let y2 = y1 + resizeOriginalH.value;

        switch (resizeMode.value) {
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

        x.value = Math.min(x1, x2);
        y.value = Math.min(y1, y2);

        width.value = Math.max(MIN_WIDTH, Math.abs(x1 - x2));
        height.value = Math.max(MIN_HEIGHT, Math.abs(y1 - y2));

        fixPosition();
    }
};

onDocumentEvent("mousemove", (e) => {
    move(positionEventFromMouseEvent(e));
});

onDocumentEvent("touchmove", (e) => {
    move(positionEventFromTouchEvent(e));
});

/**
 * Drops the widget after grabbing it for moving or resizing
 */
const drop = () => {
    if (!moving.value && !resizing.value) {
        return;
    }

    if (moving.value) {
        moving.value = false;
        savePosition();
    }

    if (resizing.value) {
        resizing.value = false;
        savePosition();
    }
};

onDocumentEvent("mouseup", drop);
onDocumentEvent("touchend", drop);
</script>
