// Event utils

"use strict";

/**
 * Calls event.stopPropagation();
 * @param ev The event
 */
export function stopPropagationEvent(ev: Event) {
    ev.stopPropagation();
}

/**
 * Calls event.preventDefault()
 * @param ev The event
 */
export function preventDefaultEvent(ev: Event) {
    ev.preventDefault();
}

/**
 * Listens for a 'keydown' event.
 * When 'Enter' is pressed, triggers the click event.
 * @param ev The event
 */
export function clickOnEnter(ev: KeyboardEvent) {
    if (ev.key === "Enter") {
        ev.preventDefault();
        ev.stopPropagation();
        (ev.target as HTMLElement).click();
    }
}
