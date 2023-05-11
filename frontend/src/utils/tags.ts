// Tags utils

"use strict";

/**
 * Parses a tag name
 * @param tag The tag name
 * @returns The tag ID
 */
export function parseTagName(tag: string): string {
    return tag.replace(/[\n]/g, " ").replace(/[\r]/g, "").trim().replace(/[\s]/g, "_").toLowerCase();
}
