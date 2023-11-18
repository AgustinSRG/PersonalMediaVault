// Named timeouts

"use strict";

/**
 * Global map of named timeouts
 */
const NamedTimeouts = new Map<string, number>();

/**
 * Sets a named timeout.
 * If there was a timeout with the same name, the old one is cleared and replaced.
 * @param name Timeout name
 * @param ms Delay in milliseconds
 * @param handler The handler function
 */
export function setNamedTimeout(name: string, ms: number, handler: () => void) {
    clearNamedTimeout(name);
    NamedTimeouts.set(
        name,
        setTimeout(() => {
            NamedTimeouts.delete(name);
            handler();
        }, ms),
    );
}

/**
 * Clears an existing named timeout
 * @param name Timeout name
 */
export function clearNamedTimeout(name: string) {
    if (NamedTimeouts.has(name)) {
        clearTimeout(NamedTimeouts.get(name));
        NamedTimeouts.delete(name);
    }
}
