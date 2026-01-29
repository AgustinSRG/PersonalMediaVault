// Util to replace multiple values in string

"use strict";

/**
 * Replaces multiple sub-strings within an input string.
 *
 * For example:
 *
 * assert(stringMultiReplace("Hello $USER, I'm $NAME.", {"$USER": "Bob", "$NAME": "Me"})  === "Hello Bob, I'm Me.")
 *
 * @param inputString The input string that may contain sub-strings to be replaced
 * @param replacements A record mapping the original sub-strings to the replacement sub-strings
 * @returns The input string after the sub-string have been replaced
 */
export function stringMultiReplace(inputString: string, replacements: Record<string, string>): string {
    const keys = Object.keys(replacements).map((key) =>
        // Escape Regexp characters
        key.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"),
    );

    if (keys.length === 0) {
        return inputString;
    }

    // Create a regex that matches every single ley
    const pattern = new RegExp(keys.map((k) => `(${k})`).join("|"), "g");

    // Replace
    return inputString.replace(pattern, (matched) => replacements[matched]);
}
