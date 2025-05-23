// Mix of TouchEvent and MouseEvent

"use strict";

/**
 * Position event
 */
export interface PositionEvent {
    /**
     * Event target
     */
    target: HTMLElement;

    /**
     * Function to call to stop the propagation of the event
     */
    stopPropagation: () => void;

    /**
     * X position
     */
    x: number;

    /**
     * Y position
     */
    y: number;
}

/**
 * Creates PositionEvent from MouseEvent
 * @param e The MouseEvent
 * @returns The PositionEvent
 */
export function positionEventFromMouseEvent(e: MouseEvent): PositionEvent {
    return {
        target: e.target as HTMLElement,
        stopPropagation: e.stopPropagation,
        x: e.pageX,
        y: e.pageY,
    };
}

/**
 * Creates PositionEvent from TouchEvent
 * @param e The TouchEvent
 * @returns The PositionEvent
 */
export function positionEventFromTouchEvent(e: TouchEvent): PositionEvent {
    return {
        target: e.target as HTMLElement,
        stopPropagation: e.stopPropagation,
        x: e.touches.length > 0 ? e.touches[0].pageX : 0,
        y: e.touches.length > 0 ? e.touches[0].pageY : 0,
    };
}
