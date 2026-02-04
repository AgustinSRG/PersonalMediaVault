// Tags composable

"use strict";

import { TagsController } from "@/control/tags";
import { computed, ref } from "vue";
import { addAppEventListener, EVENT_NAME_TAGS_UPDATE } from "@/control/app-events";

/**
 * Tags names composable
 */
export type TagsComposable = {
    /**
     * Gets the name of a tag
     * @param tag The tag ID
     * @returns The tag name
     */
    getTagName: (tag: number) => string;
};

const tagsVersion = ref(TagsController.TagsVersion);

addAppEventListener(EVENT_NAME_TAGS_UPDATE, (v) => {
    tagsVersion.value = v;
});

/**
 * Gets the tags names composable
 * @returns The tags names composable
 */
export function useTagNames(): TagsComposable {
    const getTagName = computed(() => {
        return (id: number): string => {
            return TagsController.GetTagName(id, tagsVersion.value);
        };
    }).value;

    return {
        getTagName,
    };
}
