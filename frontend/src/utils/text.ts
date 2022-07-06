// Text utils

"use strict"

/**
 * Removes acents and diacritics.
 * @param str   The input string
 * @returns     The normalized string
 */
 export function normalizeString(str: string): string {
    return str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
}

/**
 * Escapes html reserved characters.
 * @param html      Input HTML text.
 * @returns         The escaped text.
 */
export function escapeHTML(html: string): string {
    return ("" + html).replace(/&/g, "&amp;").replace(/</g, "&lt;")
        .replace(/>/g, "&gt;").replace(/"/g, "&quot;")
        .replace(/'/g, "&apos;").replace(/\//g, "&#x2f;");
}

/**
 * Turns HTML into text (removes all tags)
 * @param html HTML
 * @returns Text
 */
export function htmlToText(html: string): string {
    return (html + "").replace(/<[^>]*>/g, '').trim();
}

/**
 * Escapes single quotes and reverse bars
 * @param raw The raw input text
 * @returns The escaped text.
 */
export function escapeSingleQuotes(raw: string): string {
    return ("" + raw).replace(/\\/g, "\\\\").replace(/'/g, "\\'");
}

/**
 * Escapes double quotes and reverse bars.
 * @param raw The raw input text
 * @returns The escaped text.
 */
export function escapeDoubleQuotes(raw: string): string {
    return ("" + raw).replace(/"/g, "\\\"").replace(/\\/g, "\\\\");
}
