<template>
    <div ref="container">
        <div class="form-group media-tags">
            <label v-if="mediaTags.length === 0">{{ $t("There are no tags yet for this media.") }}</label>
            <div v-for="tag in mediaTags" :key="tag" class="media-tag">
                <div class="media-tag-name">{{ getTagName(tag) }}</div>
                <button
                    v-if="canWrite"
                    type="button"
                    :title="$t('Remove tag')"
                    class="media-tag-btn"
                    :disabled="busy"
                    @click="removeTag(tag)"
                >
                    <i class="fas fa-times"></i>
                </button>
            </div>
        </div>
        <div v-if="canWrite">
            <div class="form-group media-tags">
                <div class="media-tags-finder">
                    <input
                        ref="tagToAddInput"
                        v-model="tagFilter"
                        type="text"
                        autocomplete="off"
                        maxlength="255"
                        :disabled="busy"
                        class="form-control tag-to-add"
                        :placeholder="$t('Add tags') + '...'"
                        @input="onTagFilterChanged"
                        @keydown="onTagAddKeyDown"
                    />
                </div>
                <div class="media-tags-adder">
                    <button type="button" :disabled="!tagFilter" class="btn btn-primary btn-xs" @click="onAddTagButtonClicked">
                        <i class="fas fa-plus"></i> {{ $t("Add tag") }}
                    </button>
                </div>
            </div>
            <div v-if="tagSuggestions.length > 0" class="form-group">
                <button
                    v-for="mt in tagSuggestions"
                    :key="mt.id"
                    type="button"
                    class="btn btn-primary btn-sm btn-tag-mini btn-add-tag"
                    :disabled="busy"
                    @click="addTag(mt.name, false)"
                    @keydown="onSuggestionKeydown"
                >
                    <i class="fas fa-plus"></i> {{ mt.name }}
                </button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { apiTagsTagMedia, apiTagsUntagMedia } from "@/api/api-tags";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { useTagSuggestions } from "@/composables/use-tag-suggestions";
import { useTagNames } from "@/composables/use-tags-names";
import { useUserPermissions } from "@/composables/use-user-permissions";
import {
    emitAppEvent,
    EVENT_NAME_GO_NEXT,
    EVENT_NAME_GO_PREV,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { setLastUsedTag } from "@/control/app-preferences";
import { MediaController } from "@/control/media";
import { PagesController } from "@/control/pages";
import { TagsController } from "@/control/tags";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { nextTick, onMounted, ref, useTemplateRef } from "vue";

// Translation function
const { $t } = useI18n();

const emit = defineEmits<{
    /**
     * Emitted when media tags are updated
     */
    (e: "tags-update"): void;
}>();

const props = defineProps({
    /**
     * True to allow navigation using arrow keys
     */
    allowNavigation: Boolean,
});

// User permissions
const { canWrite } = useUserPermissions();

// Tag names
const { getTagName } = useTagNames();

// List of media tags
const mediaTags = ref<number[]>([]);

// Tag suggestions
const { tagFilter, onTagFilterChanged, tagSuggestions, updateTagSuggestions } = useTagSuggestions(
    "edit",
    (id) => !mediaTags.value.includes(id),
);

/**
 * Loads the tag list for the current media
 */
const load = () => {
    if (!MediaController.MediaData) {
        return;
    }
    mediaTags.value = (MediaController.MediaData.tags || []).slice();
    updateTagSuggestions();
    focusInput();
};

onMounted(load);

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, load);

// Input where the tag name is typed
const tagToAddInput = useTemplateRef("tagToAddInput");

/**
 * Focuses the tag input
 * @param select To select the text of the input
 */
const focusInput = (select?: boolean) => {
    nextTick(() => {
        tagToAddInput.value?.focus();

        if (select) {
            tagToAddInput.value?.select();
        }
    });
};

// True if a tag is being added or removed
const busy = ref(false);

// ID for tag update requests
const requestId = useRequestId();

/**
 * Adds a tag to the media
 * @param tag The tag name
 * @param resetTagInput True to reset the tag input
 */
const addTag = (tag: string, resetTagInput: boolean) => {
    if (busy.value) {
        return;
    }

    if (!MediaController.MediaData) {
        return;
    }

    busy.value = true;

    const mediaId = MediaController.MediaData.id;

    makeNamedApiRequest(requestId, apiTagsTagMedia(mediaId, tag))
        .onSuccess((res) => {
            busy.value = false;

            setLastUsedTag(res.id, "edit");

            PagesController.ShowSnackBar($t("Added tag") + ": " + res.name);

            if (resetTagInput) {
                tagFilter.value = "";
            }

            if (mediaTags.value.indexOf(res.id) === -1) {
                mediaTags.value.push(res.id);
            }

            updateTagSuggestions();

            TagsController.AddTag(res.id, res.name);

            if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                // Update cached media data
                if (!MediaController.MediaData.tags.includes(res.id)) {
                    MediaController.MediaData.tags.push(res.id);
                }
            }

            emit("tags-update");

            focusInput(!resetTagInput);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                invalidTagName: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Invalid tag name"));
                },
                badRequest: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Bad request"));
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            PagesController.ShowSnackBar(err.message);
            console.error(err);
            busy.value = false;
        });
};

/**
 * Removes a tag from the media
 * @param tag The tag ID
 */
const removeTag = (tag: number) => {
    if (busy.value) {
        return;
    }

    if (!MediaController.MediaData) {
        return;
    }

    busy.value = true;

    const mediaId = MediaController.MediaData.id;
    const tagName = getTagName(tag);

    makeNamedApiRequest(requestId, apiTagsUntagMedia(mediaId, tag))
        .onSuccess(({ removed }) => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Removed tag") + ": " + tagName);

            for (let i = 0; i < mediaTags.value.length; i++) {
                if (mediaTags.value[i] === tag) {
                    mediaTags.value.splice(i, 1);
                    break;
                }
            }
            if (removed) {
                TagsController.RemoveTag(tag);
            } else {
                updateTagSuggestions();
            }

            if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                // Updated cached media data
                if (MediaController.MediaData.tags.includes(tag)) {
                    MediaController.MediaData.tags = MediaController.MediaData.tags.filter((t) => {
                        return t !== tag;
                    });
                }
            }

            emit("tags-update");

            focusInput(true);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;
            handleErr(err, {
                unauthorized: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            PagesController.ShowSnackBar(err.message);
            console.error(err);
            busy.value = false;
        });
};

/**
 * Event handler for click on the tag add button
 * @param e The event
 */
const onAddTagButtonClicked = (e: Event) => {
    e.preventDefault();

    addTag(tagFilter.value, true);
};

/**
 * Event handler for 'keydown' on a suggestion element
 * @param e The keyboard event
 */
const onSuggestionKeydown = (e: KeyboardEvent) => {
    if (e.key === "ArrowRight" || e.key === "ArrowDown") {
        e.preventDefault();
        e.stopPropagation();

        const nextElem = (e.target as HTMLElement).nextSibling as HTMLElement;

        if (nextElem && nextElem.focus) {
            nextElem.focus();
        }
    } else if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
        e.preventDefault();
        e.stopPropagation();

        const prevElem = (e.target as HTMLElement).previousSibling as HTMLElement;

        if (prevElem && prevElem.focus) {
            prevElem.focus();
        } else {
            tagToAddInput.value?.focus();
        }
    }
};

// Container element
const container = useTemplateRef("container");

/**
 * Event handler for 'keydown' on the tag input element
 * @param e The keyboard event
 */
const onTagAddKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Enter") {
        e.preventDefault();
        addTag(tagFilter.value, true);
    } else if (e.key === "Tab" && !e.shiftKey) {
        updateTagSuggestions();

        if (tagSuggestions.value.length > 0) {
            if (tagSuggestions.value[0].name !== tagFilter.value) {
                e.preventDefault();
                tagFilter.value = tagSuggestions.value[0].name;
            }
        }
    } else if (e.key === "ArrowRight") {
        if (props.allowNavigation) {
            if (!tagFilter.value) {
                emitAppEvent(EVENT_NAME_GO_NEXT);
            }
        }
    } else if (e.key === "PageDown") {
        if (props.allowNavigation) {
            emitAppEvent(EVENT_NAME_GO_NEXT);
        }
    } else if (e.key === "ArrowLeft") {
        if (!tagFilter.value) {
            if (props.allowNavigation) {
                emitAppEvent(EVENT_NAME_GO_PREV);
            }
        }
    } else if (e.key === "PageUp") {
        if (props.allowNavigation) {
            emitAppEvent(EVENT_NAME_GO_PREV);
        }
    } else if (e.key === "ArrowDown") {
        const suggestionElem = container.value?.querySelector(".btn-add-tag") as HTMLElement;
        if (suggestionElem) {
            suggestionElem.focus();
        }
    }
};
</script>
