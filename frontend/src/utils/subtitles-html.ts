// HTML subtitles sanitizer

"use strict";

const HtmlSubtitlesTag = {
    counter: 0,
};

export function getUniqueSubtitlesLoadTag(): string {
    HtmlSubtitlesTag.counter++;
    return "tag-" + HtmlSubtitlesTag.counter;
}

export async function sanitizeSubtitlesHTML(html: string): Promise<string> {
    const SanitizeHTML = await import("sanitize-html");
    return SanitizeHTML.default(html, {
        allowedTags: ["b", "i", "u", "em", "strong", "font", "span"],
        allowedAttributes: {
            span: ["style"],
            font: ["face", "size", "color"],
        },
        allowedStyles: {
            "*": {
                color: [/^#(0x)?[0-9a-f]+$/i, /^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$/],
                background: [/^#(0x)?[0-9a-f]+$/i, /^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$/],
                "background-color": [/^#(0x)?[0-9a-f]+$/i, /^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$/],
                "font-size": [/^\d+(?:px|em|%)$/],
            },
        },
    });
}
