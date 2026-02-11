<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div ref="dialog" class="modal-dialog modal-xl" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Batch operation") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="batch-op-group-search">
                    <div class="form-group">
                        <label>{{ $t("Title must contain") }}:</label>
                        <input
                            v-model="textSearch"
                            type="text"
                            name="title-search"
                            autocomplete="off"
                            maxlength="255"
                            class="form-control form-control-full-width auto-focus"
                        />
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Media type") }}:</label>
                        <select v-model="typeSearch" class="form-control form-select form-control-full-width">
                            <option :value="0">{{ $t("Any media") }}</option>
                            <option :value="MEDIA_TYPE_IMAGE">{{ $t("Images") }}</option>
                            <option :value="MEDIA_TYPE_VIDEO">{{ $t("Videos") }}</option>
                            <option :value="MEDIA_TYPE_AUDIO">{{ $t("Audios") }}</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Album") }}:</label>
                        <AlbumSelect v-model:album="albumSearch" :disabled="displayProgress"></AlbumSelect>
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Tags") }}:</label>
                        <select v-model="tagModeSearch" class="form-control form-select form-control-full-width">
                            <option :value="'all'">
                                {{ $t("Media must contain ALL of the selected tags") }}
                            </option>
                            <option :value="'any'">
                                {{ $t("Media must contain ANY of the selected tags") }}
                            </option>
                            <option :value="'none'">
                                {{ $t("Media must contain NONE of the selected tags") }}
                            </option>
                            <option :value="'untagged'">
                                {{ $t("Media must be untagged") }}
                            </option>
                        </select>
                    </div>
                    <TagIdList
                        v-if="tagModeSearch !== 'untagged'"
                        v-model:tags="tagsSearch"
                        @tab-focus-skip="skipTagSuggestionsSearch"
                    ></TagIdList>
                </div>

                <hr />

                <div class="batch-op-group-action">
                    <div class="form-group"></div>
                    <div class="form-group">
                        <label>{{ $t("Select and action to apply") }}:</label>
                        <select
                            v-model="action"
                            class="form-control form-select form-control-full-width tags-focus-skip"
                            @keydown="onTagsSkipKeyDown"
                        >
                            <option :value="''">
                                {{ $t("--- Select an action ---") }}
                            </option>
                            <option :value="'tag-add'">
                                {{ $t("Add tags to the media assets") }}
                            </option>
                            <option :value="'tag-remove'">
                                {{ $t("Remove tags from the media assets") }}
                            </option>
                            <option :value="'album-add'">
                                {{ $t("Add media assets to album") }}
                            </option>
                            <option :value="'album-remove'">
                                {{ $t("Remove media assets from album") }}
                            </option>
                            <option :value="'delete'">
                                {{ $t("Delete media assets") }}
                            </option>
                        </select>
                    </div>

                    <div v-if="action === 'tag-add' || action === 'tag-remove'">
                        <TagNameList v-model:tags="tagsAction" @tab-focus-skip="skipTagSuggestionsActions"></TagNameList>
                    </div>

                    <div v-if="action === 'album-add' || action === 'album-remove'">
                        <div class="form-group">
                            <label>{{ $t("Album") }}:</label>
                            <AlbumSelect v-model:album="albumToAdd" :disabled="displayProgress"></AlbumSelect>
                        </div>
                    </div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button
                    type="button"
                    class="modal-footer-btn tags-focus-skip-action"
                    :disabled="displayProgress"
                    @click="start"
                    @keydown="onTagsSkipActionKeyDown"
                >
                    <i class="fas fa-check"></i> {{ $t("Apply") }}
                </button>
            </div>
        </div>

        <BatchOperationProgressModal
            v-if="displayProgress"
            v-model:display="displayProgress"
            :status="status"
            :progress="progress"
            :error="error"
            :action-count="actionCount"
            @update:display="afterSubModalClosed"
            @cancel="cancel"
            @confirm="confirm"
        ></BatchOperationProgressModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="actionDeleteInternal"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { ref, useTemplateRef, watch } from "vue";
import BatchOperationProgressModal from "./BatchOperationProgressModal.vue";
import { TagsController } from "@/control/tags";
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { MediaController } from "@/control/media";
import { normalizeString, filterToWords, matchSearchFilter } from "@/utils/normalize";
import type { MediaType } from "@/api/models";
import { MEDIA_TYPE_AUDIO, MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO, type MediaListItem } from "@/api/models";
import { apiAlbumsAddMediaToAlbum, apiAlbumsGetAlbum, apiAlbumsRemoveMediaFromAlbum } from "@/api/api-albums";
import { apiMediaDeleteMedia } from "@/api/api-media-edit";
import { apiTagsTagMedia, apiTagsUntagMedia } from "@/api/api-tags";
import { apiAdvancedSearch } from "@/api/api-search";
import AlbumSelect from "../utils/AlbumSelect.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import TagIdList from "../utils/TagIdList.vue";
import TagNameList from "../utils/TagNameList.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";

// Page size for requests
const PAGE_SIZE = 50;

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Text search filter
const textSearch = ref("");

// Media type filter
const typeSearch = ref<MediaType>(0);

// Album filter
const albumSearch = ref<number>(-1);

// List of tags for filtering
const tagsSearch = ref<number[]>([]);

// Tag modes
type TagSearchMode = "all" | "any" | "none" | "untagged";

// Tag filter mode
const tagModeSearch = ref<TagSearchMode>("all");

// Batch operation actions
type BatchOperationAction = "" | "tag-add" | "tag-remove" | "album-add" | "album-remove" | "delete";

// Action to perform
const action = ref<BatchOperationAction>("");

// List of tags for the action
const tagsAction = ref<string[]>([]);

// Ifd of the album for the action
const albumToAdd = ref(-1);

// True to display the progress modal
const displayProgress = ref(false);

// Progress of the batch operation (%)
const progress = ref(0);

// Statuses of the batch operation
type BatchOperationStatus = "" | "search" | "confirmation" | "confirmation-delete" | "action" | "success" | "error";

// Batch operation status
const status = ref<BatchOperationStatus>("");

// Number of media items to apply the action
const actionCount = ref(0);

// List of IDs of media elements to apply the action
const actionItems = ref<number[]>([]);

// Batch request ID
const batchRequestId = useRequestId();

// Auth confirmation
const {
    displayAuthConfirmation,
    authConfirmationCooldown,
    authConfirmationTfa,
    authConfirmationError,
    requiredAuthConfirmationPassword,
    invalidPassword,
    requiredAuthConfirmationTfa,
    invalidTfaCode,
    cooldown,
} = useAuthConfirmation();

// Media Id being deleted
const authConfirmationDeleteId = ref(-1);

// Media Id to delete next
const authConfirmationDeleteNext = ref(-1);

// Request error
const error = ref("");

watch(display, () => {
    if (display.value) {
        error.value = "";
    }
});

/**
 * Starts the batch operation
 */
const start = () => {
    if (!action.value) {
        return;
    }

    switch (action.value) {
        case "tag-add":
        case "tag-remove":
            if (tagsAction.value.length === 0) {
                return;
            }
            break;
        case "album-add":
        case "album-remove":
            if (albumToAdd.value < 0) {
                return;
            }
            break;
    }

    displayProgress.value = true;
    status.value = "search";
    progress.value = 0;
    actionItems.value = [];

    if (albumSearch.value >= 0) {
        loadAlbumSearch();
    } else {
        searchNext(null);
    }
};

/**
 * Cancels the batch operation
 */
const cancel = () => {
    abortNamedApiRequest(batchRequestId);
};

const loadAlbumSearch = () => {
    abortNamedApiRequest(batchRequestId);

    makeNamedApiRequest(batchRequestId, apiAlbumsGetAlbum(albumSearch.value))
        .onSuccess((result) => {
            filterElements(result.list);
            finishSearch();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    status.value = "error";
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    status.value = "error";
                    error.value = $t("The selected album was not found");
                },
                serverError: () => {
                    status.value = "error";
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    status.value = "error";
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            status.value = "error";
            error.value = err.message;
        });
};

/**
 * Searches media elements
 * @param continueRef The continuation token reference
 */
const searchNext = (continueRef: number | null) => {
    abortNamedApiRequest(batchRequestId);

    makeNamedApiRequest(batchRequestId, apiAdvancedSearch(getTagMode(), getTagList(), "asc", continueRef, PAGE_SIZE))
        .onSuccess((result) => {
            filterElements(result.items);

            progress.value = (Math.max(0, result.scanned) / Math.max(1, result.total_count)) * 100;

            if (result.scanned >= result.total_count) {
                // Finished
                finishSearch();
            } else {
                searchNext(result.continue);
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    status.value = "error";
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                serverError: () => {
                    status.value = "error";
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    status.value = "error";
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            status.value = "error";
            error.value = err.message;
        });
};

/**
 * Gets the tag mode for the API
 */
const getTagMode = (): "allof" | "anyof" | "noneof" => {
    switch (tagModeSearch.value) {
        case "any":
            if (tagsSearch.value.length > 16) {
                return "allof";
            }
            return "anyof";
        case "none":
            return "noneof";
        default:
            return "allof";
    }
};

/**
 * Gets the tag list for the search API
 */
const getTagList = (): string[] => {
    if (tagModeSearch.value === "untagged") {
        return [];
    }
    if (tagModeSearch.value === "any" && tagsSearch.value.length > 16) {
        return [];
    }
    return tagsSearch.value
        .map((tag) => {
            return TagsController.GetTagName(tag, TagsController.TagsVersion);
        })
        .slice(0, 16);
};

/**
 * Filters media elements and adds them to the list
 * @param elements The elements to add
 */
const filterElements = (elements: MediaListItem[]) => {
    const filterText = normalizeString(textSearch.value).trim().toLowerCase();
    const filterTextWords = filterToWords(filterText);
    const filterType = typeSearch.value;
    const filterTags = tagsSearch.value.slice();
    const filterTagMode = tagModeSearch.value;

    for (const e of elements) {
        if (filterText) {
            if (matchSearchFilter(e.title, filterText, filterTextWords) < 0) {
                continue;
            }
        }

        if (filterType) {
            if (e.type !== filterType) {
                continue;
            }
        }

        if (filterTagMode === "all") {
            if (filterTags.length > 0) {
                let passesTags = true;
                for (const tag of filterTags) {
                    if (!e.tags || !e.tags.includes(tag)) {
                        passesTags = false;
                        break;
                    }
                }

                if (!passesTags) {
                    continue;
                }
            }
        } else if (filterTagMode === "none") {
            if (filterTags.length > 0) {
                let passesTags = true;
                for (const tag of filterTags) {
                    if (e.tags && e.tags.includes(tag)) {
                        passesTags = false;
                        break;
                    }
                }

                if (!passesTags) {
                    continue;
                }
            }
        } else if (filterTagMode === "any") {
            if (filterTags.length > 0) {
                let passesTags = false;
                for (const tag of filterTags) {
                    if (!e.tags || e.tags.includes(tag)) {
                        passesTags = true;
                        break;
                    }
                }

                if (!passesTags) {
                    continue;
                }
            }
        } else if (filterTagMode === "untagged") {
            if (e.tags && e.tags.length > 0) {
                continue;
            }
        }

        actionItems.value.push(e.id);
    }
};

/**
 * Indicates the media element search finished
 */
const finishSearch = () => {
    if (actionItems.value.length > 0) {
        status.value = action.value === "delete" ? "confirmation-delete" : "confirmation";
        actionCount.value = actionItems.value.length;
    } else {
        status.value = "error";
        error.value = $t("No items found matching the specified criteria");
    }
};

/**
 * Call when the user confirms the action
 */
const confirm = () => {
    status.value = "action";
    progress.value = 0;
    actionNext(0);
};

/**
 * Performs the next action
 * @param i Index of the element to perform the action
 */
const actionNext = (i: number) => {
    abortNamedApiRequest(batchRequestId);

    if (i >= actionItems.value.length) {
        // Finish
        status.value = "success";

        AlbumsController.LoadCurrentAlbum();
        MediaController.Load();
        TagsController.Load();
        return;
    }

    const mediaId = actionItems.value[i];

    switch (action.value) {
        case "tag-add":
            actionAddTag(mediaId, tagsAction.value.slice(), i + 1);
            break;
        case "tag-remove":
            actionRemoveTag(mediaId, tagsAction.value.slice(), i + 1);
            break;
        case "album-add":
            actionAddAlbum(mediaId, i + 1);
            break;
        case "album-remove":
            actionRemoveAlbum(mediaId, i + 1);
            break;
        case "delete":
            actionDelete(mediaId, i + 1);
            break;
    }

    progress.value = ((i + 1) * 100) / (actionItems.value.length || 1);
};

/**
 * Performs the action to add tags to a media element
 * @param mid The media ID
 * @param tags The tags to add
 * @param next The next element for when the action is done
 */
const actionAddTag = (mid: number, tags: string[], next: number) => {
    if (tags.length === 0) {
        actionNext(next);
        return;
    }

    makeNamedApiRequest(batchRequestId, apiTagsTagMedia(mid, tags[0]))
        .onSuccess(() => {
            actionAddTag(mid, tags.slice(1), next);
        })
        .onRequestError((err, handleErr) => {
            status.value = "error";

            handleErr(err, {
                unauthorized: () => {
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                invalidTagName: () => {
                    error.value = $t("Invalid tag name");
                },
                badRequest: () => {
                    error.value = $t("Bad request");
                },
                accessDenied: () => {
                    error.value = $t("Access denied");
                },
                serverError: () => {
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            error.value = err.message;
            console.error(err);
            status.value = "error";
        });
};

/**
 * Performs the action to remove tags to a media element
 * @param mid The media ID
 * @param tags The tags to remove
 * @param next The next element for when the action is done
 */
const actionRemoveTag = (mid: number, tags: string[], next: number) => {
    if (tags.length === 0) {
        actionNext(next);
        return;
    }

    const tagId = TagsController.FindTag(tags[0]);

    if (tagId < 0) {
        // Tag not found
        actionRemoveTag(mid, tags.slice(1), next);
        return;
    }

    makeNamedApiRequest(batchRequestId, apiTagsUntagMedia(mid, tagId))
        .onSuccess(() => {
            actionRemoveTag(mid, tags.slice(1), next);
        })
        .onRequestError((err, handleErr) => {
            status.value = "error";
            handleErr(err, {
                unauthorized: () => {
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    error.value = $t("Access denied");
                },
                serverError: () => {
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            error.value = err.message;
            console.error(err);
            status.value = "error";
        });
};

/**
 * Performs an action to add the media into an album
 * @param mid The media ID
 * @param next The next element for when the action is done
 */
const actionAddAlbum = (mid: number, next: number) => {
    makeNamedApiRequest(batchRequestId, apiAlbumsAddMediaToAlbum(albumToAdd.value, mid))
        .onSuccess(() => {
            actionNext(next);
        })
        .onRequestError((err, handleErr) => {
            status.value = "error";
            handleErr(err, {
                unauthorized: () => {
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                maxSizeReached: () => {
                    error.value = $t("The album reached the limit of 1024 elements. Please, consider creating another album.");
                },
                badRequest: () => {
                    error.value = $t("Bad request");
                },
                accessDenied: () => {
                    error.value = $t("Access denied");
                },
                notFound: () => {
                    error.value = $t("Not found");
                },
                serverError: () => {
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            error.value = err.message;
            console.error(err);
            status.value = "error";
        });
};

/**
 * Performs an action to remove the media from an album
 * @param mid The media ID
 * @param next The next element for when the action is done
 */
const actionRemoveAlbum = (mid: number, next: number) => {
    makeNamedApiRequest(batchRequestId, apiAlbumsRemoveMediaFromAlbum(albumToAdd.value, mid))
        .onSuccess(() => {
            actionNext(next);
        })
        .onRequestError((err, handleErr) => {
            status.value = "error";
            handleErr(err, {
                unauthorized: () => {
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    error.value = $t("Access denied");
                },
                notFound: () => {
                    error.value = $t("Not found");
                },
                serverError: () => {
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            error.value = err.message;
            console.error(err);
            status.value = "error";
        });
};

/**
 * Performs the deletion of a media element
 * (with auth confirmation)
 * @param mid The media ID
 * @param next The next element for when the action is done
 */
const actionDelete = (mid: number, next: number) => {
    authConfirmationDeleteId.value = mid;
    authConfirmationDeleteNext.value = next;
    actionDeleteInternal({});
};

/**
 * Performs the deletion of a media element
 * @param confirmation Auth confirmation parameters
 */
const actionDeleteInternal = (confirmation: ProvidedAuthConfirmation) => {
    status.value = "action";
    displayProgress.value = true;

    makeNamedApiRequest(batchRequestId, apiMediaDeleteMedia(authConfirmationDeleteId.value, confirmation))
        .onSuccess(() => {
            actionNext(authConfirmationDeleteNext.value);
        })
        .onRequestError((err, handleErr) => {
            status.value = "error";

            handleErr(err, {
                unauthorized: () => {
                    error.value = $t("Access denied");
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                requiredAuthConfirmationPassword: () => {
                    status.value = "";
                    displayProgress.value = false;

                    requiredAuthConfirmationPassword();
                },
                invalidPassword: () => {
                    status.value = "";
                    displayProgress.value = false;

                    invalidPassword();
                },
                requiredAuthConfirmationTfa: () => {
                    status.value = "";
                    displayProgress.value = false;

                    requiredAuthConfirmationTfa();
                },
                invalidTfaCode: () => {
                    status.value = "";
                    displayProgress.value = false;

                    invalidTfaCode();
                },
                cooldown: () => {
                    status.value = "";
                    displayProgress.value = false;

                    cooldown();
                },
                accessDenied: () => {
                    error.value = $t("Access denied");
                },
                notFound: () => {
                    error.value = $t("Not found");
                },
                serverError: () => {
                    error.value = $t("Internal server error");
                },
                networkError: () => {
                    error.value = $t("Could not connect to the server");
                },
            });
        })
        .onUnexpectedError((err) => {
            error.value = err.message;
            console.error(err);
            status.value = "error";
        });
};

/**
 * Called when a sub-modal display status changes
 * @param subModalDisplay Sub-modal display status
 */
const afterSubModalClosed = function (subModalDisplay: boolean) {
    if (!subModalDisplay && display.value) {
        focus();
    }
};

// Reference to the dialog element
const dialog = useTemplateRef("dialog");

/**
 * Event handler for 'keydown' on the element to skip from the tags suggestions
 * @param event The keyboard event
 */
const onTagsSkipKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && event.shiftKey) {
        const inputElem = dialog.value?.querySelector(".tags-input-search") as HTMLElement;
        if (inputElem) {
            event.preventDefault();
            inputElem.focus();
        }
    }
};

/**
 * Event handler for 'keydown' on the element to skip from the tags suggestions
 * for the action part of the dialog
 * @param event The keyboard event
 */
const onTagsSkipActionKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && event.shiftKey) {
        const inputElem = dialog.value?.querySelector(".tags-input-search-action") as HTMLElement;
        if (inputElem) {
            event.preventDefault();
            inputElem.focus();
        }
    }
};

/**
 * Skips to the next element from tag suggestions
 */
const skipTagSuggestionsSearch = () => {
    const elem = dialog.value?.querySelector(".tags-focus-skip") as HTMLElement;
    if (elem) {
        elem.focus();
    }
};

/**
 * Skips to the next element from tag suggestions (action)
 */
const skipTagSuggestionsActions = () => {
    const toFocus = dialog.value?.querySelector(".tags-focus-skip-action") as HTMLElement;
    if (toFocus) {
        toFocus.focus();
    }
};
</script>
