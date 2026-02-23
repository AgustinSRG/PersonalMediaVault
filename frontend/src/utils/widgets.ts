// Utils for resizable widgets

"use strict";

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
