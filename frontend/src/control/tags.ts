// Tags data controller

"use strict";

import { makeNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AuthController } from "./auth";
import type { MediaListItem } from "@/api/models";
import { apiTagsGetTags } from "@/api/api-tags";
import { addAppEventListener, emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_TAGS_UPDATE, EVENT_NAME_UNAUTHORIZED } from "./app-events";

const REQUEST_ID = "tags-load";

/**
 * Tags data management object
 */
export class TagsController {
    /**
     * Tags data version
     */
    public static TagsVersion = 0;

    /**
     * Tags mapping ID -> Name
     */
    public static Tags: Map<number, string> = new Map();

    /**
     * Last tag ID
     */
    public static LastTagId = -1;

    /**
     * True if loading
     */
    public static Loading = true;

    /**
     * True if initially loaded
     */
    public static InitiallyLoaded = false;

    /**
     * Initialization logic
     * Runs at the app startup
     */
    public static Initialize() {
        addAppEventListener(EVENT_NAME_AUTH_CHANGED, TagsController.Load);
        TagsController.Load();
    }

    /**
     * Gets tag name by tag ID
     * @param id Tag ID
     * @param v The tags data version
     * @returns The name
     */
    public static GetTagName(id: number, v: number): string {
        if (import.meta.env.DEV) {
            if (TagsController.TagsVersion !== v) {
                console.warn("Tag version not properly updated. Current: " + TagsController.TagsVersion + " | Given: " + v);
            }
        }
        return TagsController.Tags.get(id) || "???";
    }

    /**
     * Loads tags
     */
    public static Load() {
        TagsController.Loading = true;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        clearNamedTimeout(REQUEST_ID);

        makeNamedApiRequest(REQUEST_ID, apiTagsGetTags())
            .onSuccess((tags) => {
                TagsController.Tags = new Map();

                for (const tag of tags) {
                    if (tag.id > TagsController.LastTagId) {
                        TagsController.LastTagId = tag.id;
                    }
                    TagsController.Tags.set(tag.id, tag.name);
                }

                TagsController.TagsVersion++;
                emitAppEvent(EVENT_NAME_TAGS_UPDATE, TagsController.TagsVersion);
                TagsController.Loading = false;
                TagsController.InitiallyLoaded = true;
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    temporalError: () => {
                        // Retry
                        setNamedTimeout(REQUEST_ID, 1500, TagsController.Load);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_ID, 1500, TagsController.Load);
            });
    }

    /**
     * Finds tag by name
     * @param name The tag name
     * @returns The tag ID
     */
    public static FindTag(name: string): number {
        for (const [tagId, tagName] of TagsController.Tags) {
            if (tagName === name) {
                return tagId;
            }
        }

        return -1;
    }

    /**
     * Adds tag to the local status
     * @param id The tag ID
     * @param name The tag name
     */
    public static AddTag(id: number, name: string) {
        // Remove any other tag with that name
        for (const [tagId, tagName] of TagsController.Tags) {
            if (tagName === name) {
                TagsController.Tags.delete(tagId);
                break;
            }
        }

        TagsController.Tags.set(id, name);

        TagsController.TagsVersion++;
        emitAppEvent(EVENT_NAME_TAGS_UPDATE, TagsController.TagsVersion);

        if (TagsController.LastTagId < id) {
            if (TagsController.LastTagId === id - 1) {
                // Next tag added, no de-sync
                TagsController.LastTagId = id;
            } else if (!TagsController.Loading) {
                // De-sync, reload
                TagsController.Load();
            }
        }
    }

    /**
     * Removes tag from the local status
     * @param id The tag ID
     */
    public static RemoveTag(id: number) {
        TagsController.Tags.delete(id);

        TagsController.TagsVersion++;
        emitAppEvent(EVENT_NAME_TAGS_UPDATE, TagsController.TagsVersion);
    }

    /**
     * Checks a media list for new tags, to reload the list
     * @param list The received list
     */
    public static OnMediaListReceived(list: MediaListItem[]) {
        if (TagsController.Loading) {
            return; // Already loading
        }
        for (const m of list) {
            if (!m.tags) {
                continue;
            }

            for (const t of m.tags) {
                if (t > TagsController.LastTagId) {
                    TagsController.Load();
                    return;
                }
            }
        }
    }
}

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
     * True if the tag name starts with the matching filter
     */
    starts?: boolean;

    /**
     * True if the tag name contains the matching filter
     */
    contains?: boolean;
}
