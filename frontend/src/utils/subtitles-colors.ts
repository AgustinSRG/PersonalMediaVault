// Color utils for subtitles

"use strict";

import { escapeHTML } from "./html";

const LETTERS = [
    "a",
    "b",
    "c",
    "d",
    "e",
    "f",
    "g",
    "h",
    "i",
    "j",
    "k",
    "l",
    "m",
    "n",
    "o",
    "p",
    "q",
    "r",
    "s",
    "t",
    "u",
    "v",
    "w",
    "x",
    "y",
    "z",
];

const NUMBERS = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"];

// Status of a tag being parsed
type TagStatus = {
    // Tag name
    name: string;

    // Tag color attribute
    color: string;

    // Tag inner HTML
    innerHtml: "";

    // 0 - Open, 1 - Close, 2 - Open and close
    isClose: number;
};

const HtmlRemoverStatuses = {
    TextParse: 0,
    HtmlEntityParse: 1,
    HtmlTagNameParse: 2,
    HtmlAttributesParse: 3,
    HtmlAttributeBegin: 4,
    HtmlAttributeValue: 5,
};

/**
 * Options for HtmlRemover
 */
export type HtmlRemoverOptions = {
    // True to allow font colors
    allowColors: boolean;

    // True to allow line breaks
    allowLinebreaks: boolean;
};

class HtmlRemover {
    // Final result
    private result: string;

    // Status
    private status: number;

    // Stack of HTML tags
    private tagStack: TagStatus[];

    // Current HTML tag
    private currentTag: TagStatus | null;

    // Buffer used for HTML entities
    private entityBuffer: string;

    // Buffer for HTML
    private htmlBuffer: string;

    // Buffer for attribute name
    private attributeNameBuffer: string;

    // Buffer for attribute value
    private attributeValueBuffer: string;

    // Quotes mark character
    private quotesMark: string;

    // Allow colors of subtitles?
    private allowColors: boolean;

    // Allow line breaks
    private allowLineBreaks: boolean;

    constructor() {
        this.reset();
    }

    private reset() {
        this.result = "";
        this.status = HtmlRemoverStatuses.TextParse;
        this.tagStack = [];
        this.currentTag = null;
        this.entityBuffer = "";
        this.htmlBuffer = "";
        this.attributeNameBuffer = "";
        this.attributeValueBuffer = "";
        this.quotesMark = "";
    }

    private appendResult(html: string) {
        if (this.currentTag) {
            this.currentTag.innerHtml += html;
        } else {
            this.result += html;
        }
    }

    private appendTagToResult(tag: TagStatus) {
        if (tag.name === "font" && tag.color) {
            this.appendResult(`<font color="${tag.color}">${tag.innerHtml}</font>`);
        } else {
            this.appendResult(tag.innerHtml);
        }
    }

    private finishCurrentTag() {
        let currentTag = this.currentTag;

        if (!currentTag) {
            return;
        }

        if (currentTag.isClose === 1 && this.tagStack.length > 0) {
            const lastTag = this.tagStack[this.tagStack.length - 1];

            if (lastTag && lastTag.name === currentTag.name) {
                currentTag = this.tagStack.pop();
            }
        }

        if (this.tagStack.length > 0) {
            this.currentTag = this.tagStack.pop();
        } else {
            this.currentTag = null;
        }

        this.appendTagToResult(currentTag);
    }

    private pushNewTag() {
        if (this.currentTag) {
            this.tagStack.push(this.currentTag);
        }
        this.currentTag = {
            name: "",
            color: "",
            innerHtml: "",
            isClose: 0,
        };
    }

    private handleCharTextParse(c: string, i: number): number {
        if (c === "&") {
            // Start of HTML entity
            this.status = HtmlRemoverStatuses.HtmlEntityParse;
            this.entityBuffer = "";
        } else if (c === "<") {
            // Start of HTML tag
            this.pushNewTag();
            this.status = HtmlRemoverStatuses.HtmlTagNameParse;
            this.htmlBuffer = "";
        } else if (c === '"') {
            this.appendResult("&quot;");
        } else if (c === "'") {
            this.appendResult("&apos;");
        } else if (c === "\n") {
            if (this.allowLineBreaks) {
                this.appendResult("<br />");
            }
        } else {
            this.appendResult(c);
        }

        return i + 1;
    }

    private handleCharEntityParse(c: string, i: number): number {
        if (c === ";") {
            // End of HTML entity

            let toAdd: string;

            if (this.entityBuffer) {
                toAdd = "&" + toAdd + ";";
            } else {
                toAdd = "&amp;;";
            }

            this.appendResult(toAdd);

            this.status = HtmlRemoverStatuses.TextParse;
        } else if (LETTERS.includes(c.toLowerCase()) || NUMBERS.includes(c.toLowerCase()) || (!this.entityBuffer && c === "#")) {
            // Allowed character
            this.entityBuffer += c;
        } else {
            // Forbidden character, escape this and add to the content
            const escapedPart = escapeHTML("&" + this.entityBuffer);

            this.appendResult(escapedPart);

            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        return i + 1;
    }

    private handleCharTagNameParse(c: string, i: number): number {
        const currentTag = this.currentTag;
        if (!currentTag) {
            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        this.htmlBuffer += c;

        if (c.trim() === "") {
            // Is a space: End of name
            this.status = HtmlRemoverStatuses.HtmlAttributesParse;
            this.attributeNameBuffer = "";
        } else if (LETTERS.includes(c.toLowerCase()) || NUMBERS.includes(c.toLowerCase()) || c === "-" || c === "_" || c === ".") {
            currentTag.name += c;
        } else if (c === "/") {
            currentTag.isClose = 1;
        } else if (c === ">") {
            // End of tag
            if (currentTag.isClose) {
                this.finishCurrentTag();
            }

            this.status = HtmlRemoverStatuses.TextParse;
        } else {
            // Wrong HTML tag
            this.appendResult(escapeHTML("<" + this.htmlBuffer));

            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        return i + 1;
    }

    private handleCharTagAttributesParse(c: string, i: number): number {
        const currentTag = this.currentTag;

        if (!currentTag) {
            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        this.htmlBuffer += c;

        if (c.trim() === "") {
            // Space, attribute reset
            this.attributeNameBuffer = "";
        } else if (LETTERS.includes(c.toLowerCase()) || NUMBERS.includes(c.toLowerCase()) || c === "-" || c === "_" || c === ".") {
            this.attributeNameBuffer += c;
        } else if (c === "/" && !this.attributeNameBuffer) {
            if (!currentTag.isClose) {
                currentTag.isClose = 2;
            }
        } else if (c === "=") {
            this.status = HtmlRemoverStatuses.HtmlAttributeBegin;
        } else if (c === ">") {
            // End of tag
            if (currentTag.isClose) {
                this.finishCurrentTag();
            }

            this.status = HtmlRemoverStatuses.TextParse;
        } else {
            // Wrong HTML tag
            this.appendResult(escapeHTML("<" + this.htmlBuffer));

            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        return i + 1;
    }

    private handleCharTagAttributeBegin(c: string, i: number): number {
        const currentTag = this.currentTag;

        if (!currentTag) {
            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        this.htmlBuffer += c;

        if (c === '"' || c === "'") {
            this.quotesMark = c;
            this.status = HtmlRemoverStatuses.HtmlAttributeValue;
            this.attributeValueBuffer = "";
        } else {
            // Wrong HTML tag
            this.appendResult(escapeHTML("<" + this.htmlBuffer));

            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        return i + 1;
    }

    private handleCharTagAttributeValue(c: string, i: number): number {
        const currentTag = this.currentTag;

        if (!currentTag) {
            this.status = HtmlRemoverStatuses.TextParse;
            return i;
        }

        this.htmlBuffer += c;

        if (c === this.quotesMark) {
            if (this.attributeNameBuffer.toLowerCase() === "color") {
                const color = escapeHTML(this.attributeValueBuffer);
                if (this.allowColors && window.CSS && CSS.supports && CSS.supports("color", color)) {
                    // Valid color
                    currentTag.color = color;
                }
            }

            this.status = HtmlRemoverStatuses.HtmlAttributesParse;
            this.attributeNameBuffer = "";
        } else {
            this.attributeValueBuffer += c;
        }

        return i + 1;
    }

    private handleChar(c: string, i: number): number {
        switch (this.status) {
            case HtmlRemoverStatuses.TextParse:
                return this.handleCharTextParse(c, i);
            case HtmlRemoverStatuses.HtmlEntityParse:
                return this.handleCharEntityParse(c, i);
            case HtmlRemoverStatuses.HtmlTagNameParse:
                return this.handleCharTagNameParse(c, i);
            case HtmlRemoverStatuses.HtmlAttributesParse:
                return this.handleCharTagAttributesParse(c, i);
            case HtmlRemoverStatuses.HtmlAttributeBegin:
                return this.handleCharTagAttributeBegin(c, i);
            case HtmlRemoverStatuses.HtmlAttributeValue:
                return this.handleCharTagAttributeValue(c, i);
            default:
                return i + 1;
        }
    }

    public parse(html: string, options: HtmlRemoverOptions): string {
        this.reset();
        this.allowColors = options.allowColors;
        this.allowLineBreaks = options.allowLinebreaks;

        for (let i = 0; i < html.length; i = this.handleChar(html.charAt(i), i));

        // Append the content of all the tags

        while (this.currentTag) {
            this.finishCurrentTag();
        }

        return this.result;
    }
}

const HtmlGlobalRemover = new HtmlRemover();

/**
 * Removes all HTML tags except some minimal features
 * @param html The HTML
 * @param options The options
 * @returns The HTML stripped of everything, except minimal features
 */
export function toHtmlMinimal(html: string, options: HtmlRemoverOptions): string {
    return HtmlGlobalRemover.parse(html, options);
}
