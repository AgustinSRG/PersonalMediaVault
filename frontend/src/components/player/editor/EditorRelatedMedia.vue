<template>
    <div ref="container" class="player-editor-sub-content">
        <!--- Related media -->

        <div class="form-group">
            <label>{{
                $t("You can add here references to other media assets related to this one, in order to find them more easily.")
            }}</label>
        </div>
        <div class="editor-related-media-list">
            <div v-for="(item, i) in relatedMedia" :key="item.id" class="editor-related-media-item">
                <MediaItemAlbumThumbnail :item="item"></MediaItemAlbumThumbnail>
                <div class="editor-related-media-right">
                    <div class="editor-related-media-title">
                        <a :href="getMediaURL(item.id)" target="_blank" rel="noopener noreferrer">{{ item.title || $t("Untitled") }}</a>
                    </div>
                    <div v-if="canWrite" class="editor-related-media-buttons">
                        <button
                            type="button"
                            :disabled="busy || i === 0"
                            class="btn btn-xs btn-mr btn-primary"
                            @click="moveRelatedMediaUp(i)"
                        >
                            <i class="fas fa-arrow-up"></i> {{ $t("Move up") }}
                        </button>
                        <button
                            type="button"
                            :disabled="busy || i >= relatedMedia.length - 1"
                            class="btn btn-xs btn-mr btn-primary"
                            @click="moveRelatedMediaDown(i)"
                        >
                            <i class="fas fa-arrow-down"></i> {{ $t("Move down") }}
                        </button>
                        <button type="button" :disabled="busy" class="btn btn-xs btn-mr btn-danger" @click="removeRelatedMedia(i)">
                            <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
        <div v-if="canWrite && relatedMedia.length < MAX_RELATED_MEDIA_COUNT" class="form-group">
            <button type="button" :disabled="busy" class="btn btn-xs btn-primary" @click="addRelatedMedia">
                <i class="fas fa-plus"></i> {{ $t("Add related media") }}
            </button>
        </div>
        <div v-if="canWrite" class="form-group">
            <button
                v-if="!compareMediaArrays(originalRelatedMedia, relatedMedia) || busy || !saved"
                type="button"
                class="btn btn-primary"
                :disabled="busy || compareMediaArrays(originalRelatedMedia, relatedMedia)"
                @click="saveChanges"
            >
                <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Change related media") }}
            </button>
            <button v-else type="button" disabled class="btn btn-primary">
                <i class="fas fa-check"></i> {{ $t("Saved related media") }}
            </button>
        </div>

        <AddRelatedMediaModal
            v-if="displayAddRelatedMediaModal"
            v-model:display="displayAddRelatedMediaModal"
            :mid="mid"
            :related-media="relatedMedia"
            @add-media="onAddRelatedMedia"
        ></AddRelatedMediaModal>

        <SaveChangesAskModal
            v-if="displayExitConfirmation"
            v-model:display="displayExitConfirmation"
            @yes="onExitSaveChanges"
            @no="onExitDiscardChanges"
        ></SaveChangesAskModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_MEDIA_METADATA_CHANGE, EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { getFrontendUrl } from "@/utils/api";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { computed, defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { apiMediaChangeRelatedMedia } from "@/api/api-media-edit";
import MediaItemAlbumThumbnail from "@/components/utils/MediaItemAlbumThumbnail.vue";
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useRequestId } from "@/composables/use-request-id";
import { useExitPreventer } from "@/composables/use-exit-preventer";
import { showSnackBarRight } from "@/control/snack-bar";
import { getCurrentMediaData, modifyCurrentMediaData } from "@/control/media";

// Limit of related media elements
const MAX_RELATED_MEDIA_COUNT = 16;

const AddRelatedMediaModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AddRelatedMediaModal.vue"),
});

const SaveChangesAskModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SaveChangesAskModal.vue"),
});

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

/**
 * Gets media URL
 * @param media The ID of the media
 * @returns The URL
 */
const getMediaURL = (media: number): string => {
    return getFrontendUrl({
        media,
    });
};

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Emits
const emit = defineEmits<{
    /**
     * Media changed
     */
    (e: "changed"): void;
}>();

// Current media ID
const mid = ref(AppStatus.CurrentMedia);

// Original related media list
const originalRelatedMedia = ref((getCurrentMediaData()?.related || []).slice());

// Related media list
const relatedMedia = ref(originalRelatedMedia.value.slice());

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, (mediaData) => {
    if (!mediaData) {
        return;
    }

    mid.value = mediaData.id;

    originalRelatedMedia.value = (mediaData.related || []).slice();
    relatedMedia.value = originalRelatedMedia.value.slice();
});

/**
 * Compares 2 lists of media elements
 * @param m1 List 1
 * @param m2 List 2
 * @returns True if they are equal
 */
function compareMediaArrays(m1: MediaListItem[], m2: MediaListItem[]): boolean {
    return m1.map((m) => m.id).join(",") === m2.map((m) => m.id).join(",");
}

// Dirty? (unsaved changes)
const dirty = computed<boolean>(() => !compareMediaArrays(relatedMedia.value, originalRelatedMedia.value));

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    nextTick(() => {
        const elem = container.value?.querySelector(".auto-focus") as HTMLElement;
        if (elem) {
            elem.focus();
        }
    });
};

onMounted(autoFocus);

// Display modal to add related media?
const displayAddRelatedMediaModal = ref(false);

/**
 * Opens the modal to add media
 */
const addRelatedMedia = () => {
    displayAddRelatedMediaModal.value = true;
};

/**
 * Called when media is added through the modal
 * @param m The media element
 * @param callback The callback to continue adding media
 */
const onAddRelatedMedia = (m: MediaListItem, callback: () => void) => {
    relatedMedia.value.push(m);

    if (relatedMedia.value.length >= MAX_RELATED_MEDIA_COUNT) {
        displayAddRelatedMediaModal.value = false;
    }

    callback();
};

/**
 * Moves media up
 * @param i The index of the element
 */
const moveRelatedMediaUp = (i: number) => {
    relatedMedia.value.splice(i - 1, 0, relatedMedia.value.splice(i, 1)[0]);
};

/**
 * Moves media down
 * @param i The index of the element
 */
const moveRelatedMediaDown = (i: number) => {
    relatedMedia.value.splice(i + 1, 0, relatedMedia.value.splice(i, 1)[0]);
};

/**
 * Removes related media
 * @param i The index of the element
 */
const removeRelatedMedia = (i: number) => {
    relatedMedia.value.splice(i, 1);
};

// Busy status
const busy = ref(false);

// True if saved
const saved = ref(false);

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

// Save request ID
const saveRequestId = useRequestId();

const saveChanges = (e?: Event) => {
    if (e) {
        e.preventDefault();
    }

    if (busy.value) {
        return;
    }

    busy.value = true;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(
        saveRequestId,
        apiMediaChangeRelatedMedia(
            mediaId,
            relatedMedia.value.map((m) => m.id),
        ),
    )
        .onSuccess(() => {
            showSnackBarRight($t("Successfully changed related media"));

            busy.value = false;
            saved.value = true;
            originalRelatedMedia.value = relatedMedia.value.slice();

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.related = relatedMedia.value.slice();
            });

            emit("changed");

            emitAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE);

            onSave();
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                badRequest,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            setError(err.message);
            console.error(err);
        });
};

// Exit preventer
const { displayExitConfirmation, onSave, onExitSaveChanges, onExitDiscardChanges } = useExitPreventer(dirty, saveChanges);
</script>
