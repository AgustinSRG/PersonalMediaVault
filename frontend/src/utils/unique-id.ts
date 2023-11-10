// Unique identifiers generation

"use strict";

/**
 * Status to ensure the identifiers are unique
 */
const IdentifiersStatus = {
    /**
     * Counter to make sure identifiers are unique
     */
    next: 0,
};

/**
 * Generates an unique numeric ID for the current tab
 * @returns The identifier
 */
export function getUniqueNumericId(): number {
    IdentifiersStatus.next++;
    return IdentifiersStatus.next;
}

/**
 * Generates an unique string ID for the current tab
 * @returns The identifier
 */
export function getUniqueStringId(): string {
    IdentifiersStatus.next++;
    return "uid-" + IdentifiersStatus.next;
}
