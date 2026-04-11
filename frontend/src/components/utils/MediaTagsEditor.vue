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
            <div class="form-group">
                <button
                    v-if="mediaTags.length > 0"
                    type="button"
                    :disabled="busy"
                    class="btn btn-primary btn-xs btn-mr"
                    :class="{ 'btn-mr': tagsToPaste.length > 0 }"
                    @click="copyTags"
                >
                    <i class="fas fa-copy"></i> {{ $t("Copy tags") }}
                </button>

                <button
                    v-if="tagsToPaste.length > 0"
                    type="button"
                    :disabled="busy"
                    class="btn btn-primary btn-xs btn-mr"
                    :title="$t('Paste tags') + ': ' + tagsToPaste.join(', ')"
                    @click="pasteTags"
                >
                    <i class="fas fa-paste"></i> {{ $t("Paste tags") }}
                </button>

                <button
                    v-if="tagsToPaste.length > 0"
                    type="button"
                    :disabled="busy"
                    class="btn btn-primary btn-xs btn-mr"
                    @click="clearTagsClipboard"
                >
                    <i class="fas fa-broom"></i> {{ $t("Clear clipboard") }}
                </button>
            </div>
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
                    <button type="button" :disabled="!tagFilter || busy" class="btn btn-primary btn-xs" @click="onAddTagButtonClicked">
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

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { apiTagsTagMedia, apiTagsUntagMedia } from "@/api/api-tags";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
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
    EVENT_NAME_TAGS_EDITOR_CLIPBOARD_CHANGED,
} from "@/global-state/app-events";
import { setLastUsedTag } from "@/local-storage/app-preferences";
import { indicateTagCreation, indicateTagDeletion } from "@/global-state/tags";
import { makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { computed, defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import { showSnackBar } from "@/global-state/snack-bar";
import { getCurrentMediaData, modifyCurrentMediaData } from "@/global-state/media";
import {
    copyCurrentMediaTagsIntoTagsEditorClipboard,
    getTagsEditorClipboardContent,
    setTagsEditorClipboardContent,
} from "@/global-state/tags-editor-clipboard";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

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
    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        return;
    }
    mediaTags.value = (mediaData.tags || []).slice();

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

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, serverError, networkError } = useCommonRequestErrors();

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

    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        return;
    }

    busy.value = true;

    const mediaId = mediaData.id;

    makeNamedApiRequest(requestId, apiTagsTagMedia(mediaId, tag))
        .onSuccess((res) => {
            busy.value = false;

            setLastUsedTag(res.id, "edit");

            showSnackBar($t("Added tag") + ": " + res.name);

            if (resetTagInput) {
                tagFilter.value = "";
            }

            if (mediaTags.value.indexOf(res.id) === -1) {
                mediaTags.value.push(res.id);
            }

            updateTagSuggestions();

            indicateTagCreation(res.id, res.name);

            modifyCurrentMediaData(mediaId, (metadata) => {
                // Update cached media data
                if (!metadata.tags.includes(res.id)) {
                    metadata.tags.push(res.id);
                }
            });

            emit("tags-update");

            focusInput(!resetTagInput);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidTagName: () => {
                    setError($t("Invalid tag name") + ": " + tag);
                },
                badRequest,
                accessDenied,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            setError(err.message);
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

    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        return;
    }

    busy.value = true;

    const mediaId = mediaData.id;
    const tagName = getTagName(tag);

    makeNamedApiRequest(requestId, apiTagsUntagMedia(mediaId, tag))
        .onSuccess(({ removed }) => {
            busy.value = false;

            showSnackBar($t("Removed tag") + ": " + tagName);

            for (let i = 0; i < mediaTags.value.length; i++) {
                if (mediaTags.value[i] === tag) {
                    mediaTags.value.splice(i, 1);
                    break;
                }
            }
            if (removed) {
                indicateTagDeletion(tag);
            } else {
                updateTagSuggestions();
            }

            modifyCurrentMediaData(mediaId, (metadata) => {
                // Updated cached media data
                if (metadata.tags.includes(tag)) {
                    metadata.tags = metadata.tags.filter((t) => {
                        return t !== tag;
                    });
                }
            });

            emit("tags-update");

            focusInput(true);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;
            handleErr(err, {
                unauthorized,
                accessDenied,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            setError(err.message);
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
    } else if (["c", "C"].includes(e.key) && e.ctrlKey && e.altKey && mediaTags.value.length > 0) {
        e.preventDefault();
        copyTags();
    } else if (["v", "V"].includes(e.key) && e.ctrlKey && e.altKey && tagsToPaste.value.length > 0) {
        e.preventDefault();
        pasteTags();
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

// Tags in clipboard
const tagsInClipboard = ref(getTagsEditorClipboardContent());

onApplicationEvent(EVENT_NAME_TAGS_EDITOR_CLIPBOARD_CHANGED, (tags) => {
    tagsInClipboard.value = tags.slice();
});

// Tags to be pasted
const tagsToPaste = computed(() => {
    const mediaTagsNamesSet = new Set(mediaTags.value.map((t) => getTagName(t)));
    return tagsInClipboard.value.filter((t) => !mediaTagsNamesSet.has(t));
});

/**
 * Copies the tags
 */
const copyTags = () => {
    copyCurrentMediaTagsIntoTagsEditorClipboard();
    showSnackBar($t("Tags list copied into internal clipboard"));
};

/**
 * Adds a tag for bulk tag adding
 * (no busy checking, promise-based)
 * @param tag The tag to add
 */
const addTagInBulk = (tag: string) => {
    return new Promise<boolean>((resolve) => {
        const mediaData = getCurrentMediaData();

        if (!mediaData) {
            return resolve(false);
        }

        const mediaId = mediaData.id;

        makeApiRequest(apiTagsTagMedia(mediaId, tag))
            .onSuccess((res) => {
                if (mediaTags.value.indexOf(res.id) === -1) {
                    mediaTags.value.push(res.id);
                }

                indicateTagCreation(res.id, res.name);

                modifyCurrentMediaData(mediaId, (metadata) => {
                    // Update cached media data
                    if (!metadata.tags.includes(res.id)) {
                        metadata.tags.push(res.id);
                    }
                });

                resolve(true);
            })
            .onCancel(() => {
                busy.value = false;

                resolve(false);
            })
            .onRequestError((err, handleErr) => {
                busy.value = false;

                handleErr(err, {
                    unauthorized,
                    invalidTagName: () => {
                        setError($t("Invalid tag name") + ": " + tag);
                    },
                    badRequest,
                    accessDenied,
                    serverError,
                    networkError,
                });

                resolve(false);
            })
            .onUnexpectedError((err) => {
                setError(err.message);
                console.error(err);
                busy.value = false;

                resolve(false);
            });
    });
};

/**
 * Inserts tags all at once
 * @param tags The list of tags
 */
const insertTagsInBulk = async (tags: string[]) => {
    const results = await Promise.all(tags.map((t) => addTagInBulk(t)));

    const success = !results.includes(false);

    if (success) {
        showSnackBar($t("Added tags") + ": " + tags.join(", "));
    }

    updateTagSuggestions();

    emit("tags-update");
};

/**
 * Adds tags from the clipboard
 */
const pasteTags = () => {
    if (tagsToPaste.value.length === 0) {
        return;
    }

    if (tagsToPaste.value.length === 1) {
        addTag(tagsToPaste.value[0], false);
        return;
    }

    if (busy.value) {
        return;
    }

    busy.value = true;

    insertTagsInBulk(tagsToPaste.value)
        .then(() => {
            busy.value = false;
        })
        .catch((err) => {
            console.error(err);
            busy.value = false;
        });
};

/**
 * Clears the tags clipboard
 */
const clearTagsClipboard = () => {
    setTagsEditorClipboardContent([]);
    showSnackBar($t("Cleared tags from internal clipboard"));
};
</script>
