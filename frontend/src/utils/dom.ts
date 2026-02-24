// DOM Utils

"use strict";

/**
 * Checks if an element is visible by checking the vertical coordinates
 * @param elementRect The element bounding rect
 * @param containerRect The container bounding rect
 * @returns True if visible
 */
export function elementIsVisibleVertical(elementRect: DOMRect, containerRect: DOMRect): boolean {
    const elementY1 = elementRect.top;
    const elementY2 = elementY1 + elementRect.height;

    const containerY1 = containerRect.top;
    const containerY2 = containerY1 + containerRect.height;

    return !(elementY2 < containerY1 || containerY2 < elementY1);
}
