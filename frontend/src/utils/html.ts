// HTML utils

"use strict";

/**
 * Escapes html reserved characters.
 * @param html      Input HTML text.
 * @returns         The escaped text.
 */
export function escapeHTML(html: string): string {
    return ("" + html)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&apos;")
        .replace(/\//g, "&#x2f;");
}

/**
 * Turns HTML into text (removes all tags)
 * @param html HTML
 * @returns Text
 */
export function htmlToText(html: string): string {
    return (html + "").replace(/<[^>]*>/g, "").trim();
}
