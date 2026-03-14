// Global tag list

"use strict";

import { makeNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import type { MediaListItem } from "@/api/models";
import { apiTagsGetTags } from "@/api/api-tags";
import { addAppEventListener, emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_TAGS_UPDATE, EVENT_NAME_UNAUTHORIZED } from "./app-events";
import { getUniqueStringId } from "@/utils/unique-id";
import { LOAD_RETRY_DELAY } from "@/constants";
import { isVaultLocked } from "./auth";

/**
 * Tag data with the matching information
 */
export interface MatchingTag {
    /**
     * Tag ID
     */
    id: number;

    /**
     * Tag name
     */
    name: string;

    /**
     * True of the tag name is equal to the filter
     */
    equals?: boolean;

    /**
     * True if the tag name starts with the matching filter
     */
    starts?: boolean;

    /**
     * True if the tag name contains the matching filter
     */
    contains?: boolean;
}

/**
 * Global tags state
 */
const TagsState = {
    /**
     * Tags map.
     * ID -> Name
     */
    tags: new Map<number, string>(),

    /**
     * Version of the tags.
     * Increments every time they load.
     */
    version: 0,

    /**
     * List tag ID.
     * This is used to know if tags are created,
     * since tags IDs are incremental.
     */
    lastTagId: -1,

    /**
     * True if loading
     */
    loading: true,
};

/**
 * Gets the tags map
 * @returns The tags map
 */
export function getTagsMap(): Readonly<Map<number, string>> {
    return TagsState.tags;
}

/**
 * Gets the tags data version
 * @returns The version
 */
export function getTagsVersion(): number {
    return TagsState.version;
}

/**
 * Resolves a tag name using the currently loaded tags
 * @param id Tag ID
 * @param v The tags data version
 * @returns The name
 */
export function resolveTagName(id: number, v: number): string {
    if (import.meta.env.DEV) {
        if (TagsState.version !== v) {
            console.warn("Tag version not properly updated. Current: " + TagsState.version + " | Given: " + v);
        }
    }
    return TagsState.tags.get(id) || "???";
}

/**
 * Finds tag by name
 * @param name The tag name
 * @returns The tag ID
 */
export function findTagByName(name: string): number {
    for (const [tagId, tagName] of TagsState.tags) {
        if (tagName === name) {
            return tagId;
        }
    }

    return -1;
}

// Request ID to load tags
const REQUEST_ID = getUniqueStringId();

/**
 * Loads tags list from the server
 */
function loadTags() {
    TagsState.loading = true;

    if (isVaultLocked()) {
        return; // Vault is locked
    }

    clearNamedTimeout(REQUEST_ID);

    makeNamedApiRequest(REQUEST_ID, apiTagsGetTags())
        .onSuccess((tags) => {
            TagsState.tags.clear();

            for (const tag of tags) {
                if (tag.id > TagsState.lastTagId) {
                    TagsState.lastTagId = tag.id;
                }
                TagsState.tags.set(tag.id, tag.name);
            }

            TagsState.version++;
            TagsState.loading = false;

            emitAppEvent(EVENT_NAME_TAGS_UPDATE, TagsState.version);
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(REQUEST_ID, LOAD_RETRY_DELAY, loadTags);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_ID, LOAD_RETRY_DELAY, loadTags);
        });
}

/**
 * Refreshes tags from the server
 * @param forced True to force load
 */
export function refreshTags(forced?: boolean) {
    if (!forced && TagsState.loading) {
        return;
    }

    loadTags();
}

/**
 * Indicates the creation of a new tag, after calling the API to tag a media
 * @param id The tag ID
 * @param name The tag name
 */
export function indicateTagCreation(id: number, name: string) {
    // Remove any other tag with that name
    for (const [tagId, tagName] of TagsState.tags) {
        if (tagName === name) {
            TagsState.tags.delete(tagId);
            break;
        }
    }

    TagsState.tags.set(id, name);

    TagsState.version++;
    emitAppEvent(EVENT_NAME_TAGS_UPDATE, TagsState.version);

    if (TagsState.lastTagId < id) {
        if (TagsState.lastTagId === id - 1) {
            // Next tag added, no de-sync
            TagsState.lastTagId = id;
        } else if (!TagsState.loading) {
            // De-sync, reload
            loadTags();
        }
    }
}

/**
 * Indicates a tag deletion, after the tag media list has been emptied
 * @param id The tag ID
 */
export function indicateTagDeletion(id: number) {
    TagsState.tags.delete(id);

    TagsState.version++;

    emitAppEvent(EVENT_NAME_TAGS_UPDATE, TagsState.version);
}

/**
 * Checks a media list for new tags, to reload the list
 * @param list The received list
 */
export function checkMediaListForNewTags(list: MediaListItem[]) {
    if (TagsState.loading) {
        return; // Already loading
    }

    for (const m of list) {
        if (!m.tags) {
            continue;
        }

        for (const t of m.tags) {
            if (t > TagsState.lastTagId) {
                loadTags();
                return;
            }
        }
    }
}

/**
 * Initializes the tags loader
 */
export function initializeTags() {
    addAppEventListener(EVENT_NAME_AUTH_CHANGED, loadTags);
    loadTags();
}
