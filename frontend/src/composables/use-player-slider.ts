// Player slider composable

"use strict";

import type { PositionEvent } from "@/utils/position-event";
import { positionEventFromMouseEvent, positionEventFromTouchEvent } from "@/utils/position-event";
import { isTouchDevice } from "@/utils/touch";
import type { Ref, ShallowRef } from "vue";
import { computed, onMounted, ref, watch } from "vue";
import { onDocumentEvent } from "./on-document-event";

/**
 * Player slider composable
 */
export type PlayerSliderComposable = {
    /**
     * Full width for the CSS style
     */
    fullWidth: Ref<string>;

    /**
     * Bar container width (CSS)
     */
    barContainerWidth: Ref<string>;

    /**
     * Bar container inner width (CSS)
     */
    barContainerInnerWidth: Ref<string>;

    /**
     * Bar width (CSS)
     */
    barWidth: Ref<string>;

    /**
     * Current bar width (CSS)
     */
    barCurrentWidth: Ref<string>;

    /**
     * Slider thumb 'left' style (CSS)
     */
    thumbLeft: Ref<string>;

    /**
     * Expands the slider
     */
    expand: () => void;

    /**
     * Event handler for 'mousedown' on the slider
     * @param e The mouse event
     */
    grabMouse: (e: MouseEvent) => void;

    /**
     * Event handler for 'touchstart' on the slider
     * @param e The touch event
     */
    grabTouch: (e: TouchEvent) => void;
};

/**
 * Required properties for player slider
 */
export type PlayerSliderProps = {
    /**
     * Width (pixels)
     */
    width: number;

    /**
     * Miniature mode
     */
    min?: boolean;
};

// Width of the button
const BTN_WIDTH = 40;

// Width of the button in miniature mode
const BTN_WIDTH_MIN = 24;

// Margin for full width
const FULL_WIDTH_MARGIN = 40;

// Margin for the bar container
const BAR_CONTAINER_MARGIN = 32;

// Margin for the bar inner container
const BAR_CONTAINER_INNER_MARGIN = 16;

// Margin for thumb
const THUMB_MARGIN = 8;

/**
 * Gets the player slider composable
 * @param container The container
 * @param props The properties
 * @param value The value ref
 * @param auto The auto ref
 * @param expanded The expanded state ref
 */
export function usePlayerSlider(
    container: ShallowRef<HTMLElement>,
    props: PlayerSliderProps,
    value: Ref<number>,
    auto: Ref<boolean>,
    expanded: Ref<boolean>,
): PlayerSliderComposable {
    // Is touch device?
    const IS_TOUCH_DEVICE = isTouchDevice();

    onMounted(() => {
        // For touch devices, keep it expanded
        if (IS_TOUCH_DEVICE) {
            expanded.value = true;
        }
    });

    /**
     * Expands the control
     */
    const expand = () => {
        expanded.value = true;
    };

    // Full width for the CSS style
    const fullWidth = computed<string>(() => {
        const margins = FULL_WIDTH_MARGIN;

        const barWidth = props.width;

        let btnWidth = BTN_WIDTH;

        if (props.min) {
            btnWidth = BTN_WIDTH_MIN;
        }

        const result = btnWidth + (expanded.value ? barWidth + margins : margins / 2);

        return result + "px";
    });

    // Bar container width
    const barContainerWidth = computed(() => `${props.width + BAR_CONTAINER_MARGIN}px`);

    // Bar container inner width
    const barContainerInnerWidth = computed(() => `${props.width + BAR_CONTAINER_INNER_MARGIN}px`);

    // Bar width
    const barWidth = barContainerInnerWidth;

    // Current bar width
    const barCurrentWidth = computed<string>(() => {
        let actualValue = value.value;

        if (auto.value) {
            actualValue = 0;
        }

        actualValue = Math.max(0, Math.min(1, actualValue));

        return Math.floor(actualValue * props.width) + "px";
    });

    // Thumb left style
    const thumbLeft = barCurrentWidth;

    /**
     * Sets the slider value
     * @param v The value
     */
    const setValue = (v: number) => {
        value.value = v;
        auto.value = false;
    };

    /**
     * Modifies the value by the position
     * @param x The X coordinate
     * @param y The Y coordinate
     */
    const modifyValueByPosition = (x: number, y: number) => {
        if (!container.value) {
            return;
        }

        if (typeof x !== "number" || typeof y !== "number" || isNaN(x) || isNaN(y)) {
            return;
        }

        const offset = container.value.getBoundingClientRect();

        const offsetX = offset.left + THUMB_MARGIN + (props.min ? BTN_WIDTH_MIN : BTN_WIDTH);

        if (x < offsetX) {
            setValue(0);
        } else {
            const p = x - offsetX;
            const vol = Math.min(1, p / props.width);
            setValue(vol);
        }
    };

    // Grabbed?
    const grabbed = ref(false);

    /**
     * Grabs the slider
     * @param e The position event
     */
    const grab = (e: PositionEvent) => {
        grabbed.value = true;
        modifyValueByPosition(e.x, e.y);
    };

    /**
     * Grabs the slider with the mouse
     * @param e The mouse event
     */
    const grabMouse = (e: MouseEvent) => {
        e.stopPropagation();
        if (isTouchDevice()) {
            return;
        }
        grab(positionEventFromMouseEvent(e));
    };

    /**
     * Grabs the slider with the touch screen
     * @param e The touch event
     */
    const grabTouch = (e: TouchEvent) => {
        grab(positionEventFromTouchEvent(e));
    };

    /**
     * Moves the slider
     * @param e The position event
     */
    const move = (e: PositionEvent) => {
        if (!grabbed.value) {
            return;
        }

        modifyValueByPosition(e.x, e.y);
    };

    /**
     * Drops the slider
     * @param e The position event
     */
    const drop = (e?: PositionEvent) => {
        if (!grabbed.value) {
            return;
        }

        grabbed.value = false;

        if (e) {
            modifyValueByPosition(e.x, e.y);
        }
    };

    // Ensure the control cannot be grabbed if it it collapsed
    watch(expanded, () => {
        if (!expanded.value) {
            grabbed.value = false;
        }
    });

    // Document event listeners

    if (IS_TOUCH_DEVICE) {
        onDocumentEvent("touchend", (e: TouchEvent) => {
            e.stopPropagation();
            drop(null);
        });

        onDocumentEvent("touchmove", (e: TouchEvent) => {
            move(positionEventFromTouchEvent(e));
        });
    } else {
        onDocumentEvent("mouseup", (e: MouseEvent) => {
            drop(positionEventFromMouseEvent(e));
        });

        onDocumentEvent("mousemove", (e: MouseEvent) => {
            move(positionEventFromMouseEvent(e));
        });
    }

    // Return composable

    return {
        fullWidth: fullWidth,
        barContainerWidth,
        barContainerInnerWidth,
        barWidth,
        barCurrentWidth,
        thumbLeft,
        expand,
        grabMouse,
        grabTouch,
    };
}
