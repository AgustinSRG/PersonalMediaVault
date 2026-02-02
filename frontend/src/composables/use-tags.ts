// Tags composable

"use strict";

import { TagsController } from "@/control/tags";
import { computed, ref } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_TAGS_UPDATE } from "@/control/app-events";

/**
 * Tags composable
 */
export type TagsComposable = {
    /**
     * Gets the tags
     * @returns The tags map
     */
    getTags: () => Map<number, string>;

    /**
     * Gets the name of a tag
     * @param tag The tag ID
     * @returns The tag name
     */
    getTagName: (tag: number) => string;

    /**
     * Listens for updates in the tags list
     * @param listener The listener
     */
    onTagsUpdated: (listener: () => void) => void;
};

/**
 * Gets the tags composable
 * @returns The tags composable
 */
export function useTags(): TagsComposable {
    const tagsVersion = ref(TagsController.TagsVersion);

    const getTags = () => {
        return TagsController.Tags;
    };

    const getTagName = computed(() => {
        return (id: number): string => {
            return TagsController.GetTagName(id, tagsVersion.value);
        };
    }).value;

    const updateListeners: (() => void)[] = [];

    const onTagsUpdated = (listener: () => void) => {
        updateListeners.push(listener);
    };

    onApplicationEvent(EVENT_NAME_TAGS_UPDATE, (v) => {
        tagsVersion.value = v;

        updateListeners.forEach((l) => l());
    });

    return {
        getTags,
        getTagName,
        onTagsUpdated,
    };
}
