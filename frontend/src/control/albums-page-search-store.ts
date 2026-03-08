// Global store for albums page search value

"use strict";

/**
 * Global store for album search value
 */
const GlobalAlbumsPageSearch = {
    value: "",
};

/**
 * Gets the global albums page search value
 * @returns The value
 */
export function getAlbumsPageSearch(): string {
    return GlobalAlbumsPageSearch.value;
}

/**
 * Sets the albums page search value
 * @param value The value
 */
export function setAlbumsPageSearch(value: string) {
    GlobalAlbumsPageSearch.value = value;
}
