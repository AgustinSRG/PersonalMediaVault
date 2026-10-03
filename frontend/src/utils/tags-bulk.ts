// Tags add / remove in bulk helper functions

"use strict";

import type { RequestError } from "@asanrom/request-browser";

/**
 * Limit for the bulk tag APIs
 */
export const TAGS_BULK_API_LIMIT = 64;

/**
 * Extracts the name of the tag with invalid name from the error
 * @param err The request error
 * @returns The tag name (empty if could not extract it)
 */
export function extractTagNameFromError(err: RequestError): string {
    if (!err.body) {
        return "";
    }

    try {
        const parsedBody = JSON.parse(err.body) as unknown;

        if (!parsedBody || typeof parsedBody !== "object" || !("message" in parsedBody) || typeof parsedBody.message !== "string") {
            return "";
        }

        return parsedBody.message.split(":").slice(1).join(":").trim();
    } catch {
        return "";
    }
}

/**
 * Removes duplicated tag names
 * @param tags Rag tag names list
 * @returns List without duplicates
 */
export function removeDuplicatedTagNames(tags: string[]): string[] {
    const tagNamesSet = new Set<string>();

    const result: string[] = [];

    for (const tag of tags) {
        if (tagNamesSet.has(tag)) {
            continue;
        }

        result.push(tag);

        tagNamesSet.add(tag);
    }

    return result;
}

/**
 * Splits tag names in batches
 * @param tags The list fo tag names
 * @returns The batches of tag names
 */
export function splitTagNamesInBatches(tags: string[]): string[][] {
    let tagsCopy = tags.slice();

    const batches: string[][] = [];

    while (tagsCopy.length > 0) {
        if (tagsCopy.length <= TAGS_BULK_API_LIMIT) {
            batches.push(tagsCopy);
            tagsCopy = [];
        } else {
            batches.push(tagsCopy.slice(0, TAGS_BULK_API_LIMIT));
            tagsCopy = tagsCopy.slice(TAGS_BULK_API_LIMIT);
        }
    }

    return batches;
}
