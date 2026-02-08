// Tags suggestions composable

"use strict";

import type { MatchingTag } from "@/control/tags";
import { TagsController } from "@/control/tags";
import type { Ref } from "vue";
import { ref } from "vue";
import { useTimeout } from "./use-timeout";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_TAGS_UPDATE } from "@/control/app-events";
import { parseTagName } from "@/utils/tags";
import { getLastUsedTags } from "@/control/app-preferences";

/**
 * Composable for tag suggestions
 */
export type TagSuggestionsComposable = {
    /**
     * Tag filter
     */
    tagFilter: Ref<string>;

    /**
     * Call when tagFilter is changed by the user,
     * in order to update the suggestions with a delay,
     * to minimize CPU usage.
     */
    onTagFilterChanged: () => void;

    /**
     * Tag suggestions
     */
    tagSuggestions: Ref<MatchingTag[]>;

    /**
     * Updates tag suggestions
     */
    updateTagSuggestions: () => void;
};

/**
 * Composable for tag suggestions
 * @param suggestionFilter A filter for the tag suggestions
 */
export function useTagSuggestions(suggestionFilter?: (id: number, name: string) => boolean): TagSuggestionsComposable {
    // Max number of tag suggestions
    const TAGS_SUGGESTION_LIMIT = 10;

    // Timeout to update the found tags
    const findTagTimeout = useTimeout();

    // Name of the tag being added
    const tagFilter = ref("");

    // Matching tags based on 'tagFilter'
    const tagSuggestions = ref<MatchingTag[]>([]);

    /**
     * Updates matching tags
     */
    const findTags = () => {
        findTagTimeout.clear();

        const nameFilter = parseTagName(tagFilter.value);

        const lastUsedTagsIds = getLastUsedTags();

        if (!nameFilter) {
            const lastUsedTags: MatchingTag[] = [];
            const addedTagIds: number[] = [];

            for (const tid of lastUsedTagsIds) {
                if (!TagsController.Tags.has(tid) || addedTagIds.includes(tid)) {
                    continue;
                }

                const tagName = TagsController.Tags.get(tid);

                if (suggestionFilter && !suggestionFilter(tid, tagName)) {
                    continue;
                }

                lastUsedTags.push({
                    id: tid,
                    name: TagsController.Tags.get(tid),
                });

                addedTagIds.push(tid);

                if (lastUsedTags.length >= TAGS_SUGGESTION_LIMIT) {
                    break;
                }
            }

            if (lastUsedTags.length < TAGS_SUGGESTION_LIMIT) {
                Array.from(TagsController.Tags.entries())
                    .filter((t) => {
                        if (suggestionFilter && !suggestionFilter(t[0], t[1])) {
                            return false;
                        }

                        return !addedTagIds.includes(t[0]);
                    })
                    .sort((a, b) => {
                        if (a[1] < b[1]) {
                            return -1;
                        } else {
                            return 1;
                        }
                    })
                    .slice(0, TAGS_SUGGESTION_LIMIT - lastUsedTags.length)
                    .forEach((t) => {
                        lastUsedTags.push({
                            id: t[0],
                            name: t[1],
                        });
                    });
            }

            tagSuggestions.value = lastUsedTags;

            return;
        }

        tagSuggestions.value = Array.from(TagsController.Tags.entries())
            .map((a) => {
                const i = a[1].indexOf(nameFilter);
                const lastUsedIndex = lastUsedTagsIds.indexOf(a[0]);
                return {
                    id: a[0],
                    name: a[1],
                    equals: i === 0 && a[1].length === nameFilter.length,
                    starts: i === 0,
                    contains: i >= 0,
                    lastUsed: lastUsedIndex === -1 ? lastUsedTagsIds.length : lastUsedIndex,
                };
            })
            .filter((a) => {
                if (suggestionFilter && !suggestionFilter(a.id, a.name)) {
                    return false;
                }

                return a.starts || a.contains;
            })
            .sort((a, b) => {
                if (a.equals && !b.equals) {
                    return -1;
                } else if (!a.equals && b.equals) {
                    return 1;
                } else if (a.starts && !b.starts) {
                    return -1;
                } else if (b.starts && !a.starts) {
                    return 1;
                } else if (a.lastUsed < b.lastUsed) {
                    return -1;
                } else if (a.lastUsed > b.lastUsed) {
                    return 1;
                } else if (a.name < b.name) {
                    return -1;
                } else {
                    return 1;
                }
            })
            .slice(0, TAGS_SUGGESTION_LIMIT);
    };

    onApplicationEvent(EVENT_NAME_TAGS_UPDATE, findTags);

    findTags();

    // Refresh the tag list from the server
    TagsController.Refresh();

    // Delay to update the matching tags (milliseconds)
    const TAGS_UPDATE_DELAY = 200;

    /**
     * Called whenever the input tag changes
     * in order to schedule updating the matching tags.
     */
    const onTagFilterChanged = () => {
        if (findTagTimeout.isSet()) {
            return;
        }

        findTagTimeout.set(findTags, TAGS_UPDATE_DELAY);
    };

    return {
        tagFilter,
        onTagFilterChanged,
        tagSuggestions,
        updateTagSuggestions: findTags,
    };
}
