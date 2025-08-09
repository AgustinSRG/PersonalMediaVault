// HTML utils

"use strict";

/**
 * Escapes html reserved characters.
 * @param html      Input HTML text.
 * @returns         The escaped text.
 */
export function escapeHTML(html: string): string {
    return ("" + html).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&apos;");
}

const domainRegex = "[a-z0-9\\-]+(?:[.][a-z0-9\\-]+)*";
const parenthesisRegex = "[(](?:[^\\s()<>&]|&amp;)*[)]";
const linkRegex = new RegExp(
    "\\b" +
        "(?:" +
        "(?:" +
        // When using www. or http://, allow any-length TLD (like .museum)
        "(?:https?://|www[.])" +
        domainRegex +
        "|" +
        domainRegex +
        "[.]" +
        // Allow a common TLD, or any 2-3 letter TLD followed by : or /
        "(?:com?|org|net|edu|info|us|jp|[a-z]{2,3}(?=[:/]))" +
        ")" +
        "(?:[:][0-9]+)?" +
        "\\b" +
        "(?:" +
        "/" +
        "(?:" +
        "(?:" +
        "[^\\s()&]|&amp;|&quot;" +
        "|" +
        parenthesisRegex +
        ")*" +
        // URLs usually don't end with punctuation, so don't allow
        // punctuation symbols that probably aren't related to URL.
        "(?:" +
        "[^\\s`()\\[\\]{}'\".,!?;:&]" +
        "|" +
        parenthesisRegex +
        ")" +
        ")?" +
        ")?" +
        "|[a-z0-9.]+\\b@" +
        domainRegex +
        "[.][a-z]{2,3}" +
        ")",
    "ig",
);

/**
 * Replaces links in a text with its corresponding HTML tags
 * @param str The text string
 * @returns The HTML string
 */
export function replaceLinks(str: string): string {
    return str.replace(linkRegex, function (uri) {
        if (/^[a-z0-9.]+\@/gi.test(uri)) {
            return '<a href="mailto:' + uri + '" target="_blank">' + uri + "</a>";
        }
        // Insert http:// before URIs without a URI scheme specified.
        const fulluri = uri.replace(/^([a-z]*[^a-z:])/g, "http://$1");
        return '<a href="' + fulluri + '" target="_blank" rel="noopener noreferrer">' + uri + "</a>";
    });
}
