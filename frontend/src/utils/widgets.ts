// Utils for resizable widgets

"use strict";

/**
 * Widget initial position
 */
export type WidgetInitialPosition = "center" | "top-left" | "top-right" | "bottom-left" | "bottom-right";

/**
 * Widget action button
 */
export interface WidgetActionButton {
    /**
     * Button ID
     */
    id: string;

    /**
     * Button name
     */
    name: string;

    /**
     * Button icon (FA)
     */
    icon: string;

    /**
     * List of keyboard keys that trigger the action button
     */
    key?: string | string[];
}

/**
 * Widget rectangle
 * (position + size)
 */
export type WidgetRectangle = {
    x: number;
    y: number;
    width: number;
    height: number;
};

/**
 * Computes the initial position of a widget
 * @param parentElem The parent element (container of the widget)
 * @param initialWidth The initial width
 * @param initialHeight The initial height
 * @param initialPosition The initial desired position
 * @returns The position and size for the widget
 */
export function computeWidgetInitialPositionAndSize(
    parentElem: HTMLElement,
    initialWidth: number,
    initialHeight: number,
    initialPosition: WidgetInitialPosition,
): WidgetRectangle {
    const bounds = parentElem.getBoundingClientRect();

    const result: WidgetRectangle = {
        x: 0,
        y: 0,
        width: initialWidth,
        height: initialHeight,
    };

    if (bounds.width < initialWidth) {
        result.x = 0;
        result.width = bounds.width;
    } else {
        switch (initialPosition) {
            case "top-left":
            case "bottom-left":
                result.x = 0;
                break;
            case "top-right":
            case "bottom-right":
                result.x = Math.floor(bounds.width - initialWidth);
                break;
            default:
                // Center
                result.x = Math.floor((bounds.width - initialWidth) / 2);
        }

        result.width = initialWidth;
    }

    if (bounds.height < initialHeight) {
        result.y = 0;
        result.height = bounds.height;
    } else {
        switch (initialPosition) {
            case "top-left":
            case "top-right":
                result.y = 0;
                break;
            case "bottom-left":
            case "bottom-right":
                result.y = Math.floor(bounds.height - initialHeight);
                break;
            default:
                // Center
                result.y = Math.floor((bounds.height - initialHeight) / 2);
        }

        result.height = initialHeight;
    }

    return result;
}
