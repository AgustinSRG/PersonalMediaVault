// Store for description content when login is required mid-edit

"use strict";

const GlobalDescriptionStore = {
    id: -1,
    content: "",
};

/**
 * Gets the stored description
 * @returns The stored description associated to the media ID
 */
export function getStoredDescription(): [id: number, content: string] {
    return [GlobalDescriptionStore.id, GlobalDescriptionStore.content];
}

/**
 * Sets the stored description
 * @param id The media ID
 * @param content The content
 */
export function setStoredDescription(id: number, content: string) {
    GlobalDescriptionStore.id = id;
    GlobalDescriptionStore.content = content;
}
