// Normalize string

"use strict";

/**
 * Removes accents and diacritics.
 * @param str   The input string
 * @returns     The normalized string
 */
export function normalizeString(str: string): string {
    return str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
}

/**
 * Turns filter into words list
 * @param filter The filter
 * @returns The words
 */
export function filterToWords(filter: string): string[] {
    return normalizeString(filter)
        .split(" ")
        .filter((w) => !!w);
}

/**
 * Checks for matches
 * @param str The string
 * @param filter The filter to find
 * @param filterWords The filter splitted into words
 * @returns -1 if no match, 0 if starts with it, 1 if contains it
 */
export function matchSearchFilter(str: string, filter: string, filterWords: string[]): number {
    if (filterWords.length === 0) {
        return 0;
    }

    str = normalizeString(str).trim().toLowerCase();

    if (str.startsWith(filter)) {
        return 0;
    }

    for (const word of filterWords) {
        if (!str.includes(word)) {
            return -1;
        }
    }

    return 1;
}
