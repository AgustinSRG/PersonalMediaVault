<template>
    <div ref="container" class="player-editor-sub-content content-row">
        <div class="row-general-left">
            <!--- Title -->

            <form @submit="changeTitle">
                <div class="form-group">
                    <label>{{ $t("Title") }}:</label>
                    <input
                        v-model="title"
                        type="text"
                        autocomplete="off"
                        :readonly="!canWrite"
                        maxlength="255"
                        :disabled="busyTitle"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div v-if="canWrite" class="form-group">
                    <button
                        v-if="originalTitle !== title || busyTitle || !savedTitle"
                        type="submit"
                        class="btn btn-primary"
                        :disabled="busyTitle || originalTitle === title || !title"
                    >
                        <LoadingIcon icon="fas fa-pencil-alt" :loading="busyTitle"></LoadingIcon> {{ $t("Change title") }}
                    </button>
                    <button v-else type="button" disabled class="btn btn-primary">
                        <i class="fas fa-check"></i> {{ $t("Saved title") }}
                    </button>
                    <div v-if="errorTitle" class="form-error form-error-pt">{{ $t("Error") }}: {{ errorTitle }}</div>
                </div>
            </form>

            <!--- Extra config -->

            <div v-if="mediaHasDuration">
                <div v-if="canWrite" class="form-group">
                    <label
                        >{{ $t("Extra media configuration") }}:<span v-if="busyExtra" class="extra-config-loader loader-delayed-custom">
                            <i class="fa fa-spinner fa-spin"></i> {{ $t("Saving changes") }}...</span
                        ></label
                    >
                </div>
                <div class="table-responsive">
                    <table class="table no-border no-margin">
                        <tbody>
                            <tr>
                                <td class="text-right td-shrink">
                                    <toggle-switch
                                        v-model:val="startBeginning"
                                        :disabled="busyExtra || !canWrite"
                                        @update:val="changeExtraParams"
                                    ></toggle-switch>
                                </td>
                                <td class="">
                                    {{ $t("Reset time to the beginning every time the media reloads?") }}
                                </td>
                            </tr>

                            <tr v-if="type === MEDIA_TYPE_VIDEO">
                                <td class="text-right td-shrink">
                                    <toggle-switch
                                        v-model:val="isAnimation"
                                        :disabled="busyExtra || !canWrite"
                                        @update:val="changeExtraParams"
                                    ></toggle-switch>
                                </td>
                                <td class="">
                                    {{ $t("Is animation? (Force loop and disable keyboard time skipping)") }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div v-if="errorExtraConfig" class="form-group form-error">{{ $t("Error") }}: {{ errorExtraConfig }}</div>
            </div>

            <!--- Tags -->

            <div class="form-group">
                <label>{{ $t("Tags") }}:</label>
            </div>

            <MediaTagsEditor @tags-update="onMediaTagsChanged"></MediaTagsEditor>
        </div>

        <div class="row-general-right">
            <!--- Thumbnail -->

            <div class="form-group">
                <label>{{ $t("Thumbnail") }}:</label>
            </div>
            <div class="form-group" @drop="onDrop">
                <label v-if="!thumbnail">{{ $t("No thumbnail set for this media") }}</label>
                <ThumbImage v-if="thumbnail" :src="getAssetURL(thumbnail)" class-name="form-group-thumbnail"></ThumbImage>
            </div>
            <div v-if="canWrite" class="form-group">
                <input ref="thumbnailHiddenFileInput" type="file" class="file-hidden" name="thumbnail-upload" @change="inputFileChanged" />
                <div class="text-center form-group-thumbnail-buttons">
                    <button
                        v-if="!busyThumbnail"
                        type="button"
                        class="btn btn-primary image-thumbnail-button"
                        @click="uploadThumbnail"
                        @drop="onDrop"
                    >
                        <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
                    </button>
                    <button v-if="busyThumbnail" type="button" class="btn btn-primary image-thumbnail-button" disabled>
                        <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading thumbnail") }}...
                    </button>
                </div>

                <div v-if="mediaElementReady" class="form-group-pt text-center form-group-thumbnail-buttons">
                    <button
                        v-if="type === MEDIA_TYPE_IMAGE"
                        type="button"
                        class="btn btn-primary btn-sm image-thumbnail-button"
                        :title="$t('Set current image as thumbnail')"
                        :disabled="busyThumbnail"
                        @click="setCurrentImageAsThumbnail"
                    >
                        <i class="fas fa-image"></i> {{ $t("Set current image as thumbnail") }}
                    </button>
                    <button
                        v-if="type === MEDIA_TYPE_VIDEO"
                        type="button"
                        class="btn btn-primary btn-sm image-thumbnail-button"
                        :title="$t('Set current frame as thumbnail')"
                        :disabled="busyThumbnail"
                        @click="setCurrentFrameAsThumbnail"
                    >
                        <i class="fas fa-image"></i> {{ $t("Set current frame as thumbnail") }}
                    </button>
                </div>
                <div v-if="errorThumbnail" class="form-error form-error-pt text-center">{{ $t("Error") }}: {{ errorThumbnail }}</div>
            </div>
        </div>

        <ThumbnailCropModal
            v-if="displayThumbnailModal"
            v-model:display="displayThumbnailModal"
            :image-url="thumbnailModalUrl"
            @done="changeThumbnail"
            @error="onThumbnailModalError"
        ></ThumbnailCropModal>

        <SaveChangesAskModal
            v-if="displayExitConfirmation"
            v-model:display="displayExitConfirmation"
            @yes="onExitSaveChanges"
            @no="onExitDiscardChanges"
        ></SaveChangesAskModal>
    </div>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_MEDIA_METADATA_CHANGE, EVENT_NAME_MEDIA_UPDATE } from "@/global-state/app-events";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, useTemplateRef } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { apiMediaChangeExtraParams, apiMediaChangeMediaThumbnail, apiMediaChangeMediaTitle } from "@/api/api-media-edit";
import ThumbImage from "@/components/utils/ThumbImage.vue";
import MediaTagsEditor from "@/components/utils/MediaTagsEditor.vue";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useRequestId } from "@/composables/use-request-id";
import type { MediaType } from "@/api/models";
import { MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO, MEDIA_TYPE_AUDIO } from "@/constants";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useExitPreventer } from "@/composables/use-exit-preventer";
import { useTimeout } from "@/composables/use-timeout";
import { showSnackBarRight } from "@/global-state/snack-bar";
import { refreshCurrentAlbum } from "@/global-state/album";
import { getCurrentMediaData, modifyCurrentMediaData } from "@/global-state/media";
import { getNavigationStatus } from "@/global-state/navigation";

const SaveChangesAskModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SaveChangesAskModal.vue"),
});

const ThumbnailCropModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ThumbnailCropModal.vue"),
});

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

// Media type
const type = ref<MediaType>(getCurrentMediaData()?.type || 0);

// The media has duration?
const mediaHasDuration = computed(() => [MEDIA_TYPE_AUDIO, MEDIA_TYPE_VIDEO].includes(type.value));

// Original media title
const originalTitle = ref(getCurrentMediaData()?.title || "");

// Media title
const title = ref(originalTitle.value);

// Original value or "Force start beginning" option
const originalStartBeginning = ref(getCurrentMediaData()?.force_start_beginning || false);

// Start media from beginning
const startBeginning = ref(originalStartBeginning.value);

// Orinal value of "Is animation" option
const originalIsAnimation = ref(getCurrentMediaData()?.is_anim || false);

// Is animation? (only for videos)
const isAnimation = ref(originalIsAnimation.value);

// Media thumbnail
const thumbnail = ref(getCurrentMediaData()?.thumbnail || "");

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, (mediaData) => {
    if (!mediaData) {
        return;
    }

    type.value = mediaData.type;

    originalTitle.value = mediaData.title;
    title.value = originalTitle.value;

    originalStartBeginning.value = mediaData.force_start_beginning;
    startBeginning.value = originalStartBeginning.value;

    originalIsAnimation.value = mediaData.is_anim;
    isAnimation.value = originalIsAnimation.value;

    thumbnail.value = mediaData.thumbnail;
});

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

// Has title change
const dirtyTitle = computed(() => originalTitle.value !== title.value);

// Busy saving title
const busyTitle = ref(false);

// Saved title
const savedTitle = ref(false);

// Error for title
const errorsTitle = useCommonRequestErrors();
const errorTitle = errorsTitle.error;

// Request ID for the title
const requestIdTitle = useRequestId();

/**
 * Changes the title
 * @param e The submit event (optional)
 */
const changeTitle = (e?: Event) => {
    if (e) {
        e.preventDefault();
    }

    if (busyTitle.value) {
        return;
    }

    busyTitle.value = true;
    errorTitle.value = "";

    const mediaId = getNavigationStatus().media;

    makeNamedApiRequest(requestIdTitle, apiMediaChangeMediaTitle(mediaId, title.value))
        .onSuccess(() => {
            showSnackBarRight($t("Successfully changed title"));

            busyTitle.value = false;
            savedTitle.value = true;
            originalTitle.value = title.value;

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.title = title.value;
            });

            emit("changed");

            refreshCurrentAlbum();
            emitAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, mediaId);

            onSave();
        })
        .onCancel(() => {
            busyTitle.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyTitle.value = false;

            handleErr(err, {
                unauthorized: errorsTitle.unauthorized,
                invalidTitle: () => {
                    errorsTitle.setError($t("Invalid title provided"));
                },
                badRequest: errorsTitle.badRequest,
                accessDenied: errorsTitle.accessDenied,
                notFound: errorsTitle.notFound,
                serverError: errorsTitle.serverError,
                networkError: errorsTitle.serverError,
            });
        })
        .onUnexpectedError((err) => {
            busyTitle.value = false;
            errorsTitle.setError(err.message);
            console.error(err);
        });
};

// Exit preventer
const { displayExitConfirmation, onSave, onExitSaveChanges, onExitDiscardChanges } = useExitPreventer(dirtyTitle, changeTitle);

// Saving extra config?
const busyExtra = ref(false);

// Error for extra config
const errorsExtraConfig = useCommonRequestErrors();
const errorExtraConfig = errorsExtraConfig.error;

// Request ID for extra config
const requestIdExtra = useRequestId();

/**
 * Changes the extra configuration params of the media
 */
const changeExtraParams = () => {
    if (busyExtra.value) {
        return;
    }

    busyExtra.value = true;
    errorExtraConfig.value = "";

    const mediaId = getNavigationStatus().media;

    makeNamedApiRequest(requestIdExtra, apiMediaChangeExtraParams(mediaId, startBeginning.value, isAnimation.value))
        .onSuccess(() => {
            showSnackBarRight($t("Successfully changed media extra params"));

            busyExtra.value = false;
            originalStartBeginning.value = startBeginning.value;
            originalIsAnimation.value = isAnimation.value;

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.force_start_beginning = startBeginning.value;
                metadata.is_anim = isAnimation.value;
            });

            emit("changed");
        })
        .onCancel(() => {
            busyExtra.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyExtra.value = false;
            handleErr(err, errorsExtraConfig);
        })
        .onUnexpectedError((err) => {
            busyExtra.value = false;
            errorsExtraConfig.setError(err.message);
            console.error(err);
        });
};

/**
 * Called when the media tags are changed
 */
const onMediaTagsChanged = () => {
    emit("changed");
};

// True if the media element is ready
// (this means it is possible to get a thumbnail from the current image / frame)
const mediaElementReady = ref(false);

// Timeout to check the media element after a delay
const mediaElementCheckTimer = useTimeout();

// Delay to retry checking the media element (milliseconds)
const CHECK_MEDIA_ELEMENT_RETRY_DELAY = 1500;

/**
 * Checks media element
 */
const checkMediaElement = () => {
    mediaElementCheckTimer.clear();

    if (type.value === MEDIA_TYPE_IMAGE) {
        const imageElement = document.querySelector(".player-container .player-img-element") as HTMLImageElement;
        mediaElementReady.value = imageElement && imageElement.complete;
    } else if (type.value === MEDIA_TYPE_VIDEO) {
        const videoElement = document.querySelector(".player-container .player-video-element") as HTMLVideoElement;
        mediaElementReady.value = videoElement && videoElement.readyState === 4;
    } else {
        return;
    }

    if (!mediaElementReady.value) {
        mediaElementCheckTimer.set(checkMediaElement, CHECK_MEDIA_ELEMENT_RETRY_DELAY);
    }
};

onMounted(() => {
    checkMediaElement();
});

// A temporal image element to load thumbnails
const tempImage = shallowRef<HTMLImageElement | null>(null);

// Remove temp image event listeners if the component unmounts
onBeforeUnmount(() => {
    if (tempImage.value) {
        delete tempImage.value.onload;
        delete tempImage.value.onerror;
    }
});

// Display the modal to crop the thumbnail image into an square?
const displayThumbnailModal = ref(false);

// Url of the image for the crop modal
const thumbnailModalUrl = ref("");

// Busy updating the thumbnail
const busyThumbnail = ref(false);

// Error for changing the thumbnail
const errorsThumbnail = useCommonRequestErrors();
const errorThumbnail = errorsThumbnail.error;

/**
 * Prepares the thumbnail to be changed
 * @param file The new thumbnail image file
 */
const prepareChangeThumbnail = (file: File) => {
    if (tempImage.value) {
        delete tempImage.value.onload;
        delete tempImage.value.onerror;
        tempImage.value = null;
    }

    const url = URL.createObjectURL(file);

    const img = new Image();
    tempImage.value = img;

    img.onload = () => {
        tempImage.value = null;

        if (img.width === img.height) {
            changeThumbnail(file);
            return;
        }

        thumbnailModalUrl.value = url;
        errorThumbnail.value = "";
        displayThumbnailModal.value = true;
    };

    img.onerror = (err) => {
        tempImage.value = null;
        console.error(err);
        onThumbnailModalError();
    };

    img.src = url;
};

/**
 * Called when an error happens loading the image
 */
const onThumbnailModalError = () => {
    errorsThumbnail.setError($t("Invalid thumbnail provided"));
};

// Hidden file input to upload thumbnails
const thumbnailHiddenFileInput = useTemplateRef("thumbnailHiddenFileInput");

/**
 * Opens the file selector to upload a thumbnail
 */
const uploadThumbnail = () => {
    const fileElem = thumbnailHiddenFileInput.value;
    if (fileElem) {
        fileElem.value = null;
        fileElem.click();
    }
};

/**
 * Event handler for 'change' on the hidden thumbnail file input
 * @param e The event
 */
const inputFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        const file = data[0];
        prepareChangeThumbnail(file);
    }
};

/**
 * Event handler for 'drop' on the thumbnail area
 * @param e The drop event
 */
const onDrop = (e: DragEvent) => {
    e.preventDefault();
    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        const file = data[0];
        prepareChangeThumbnail(file);
    }
};

// Request ID for updating the thumbnail
const requestIdThumbnail = useRequestId();

/**
 * Changes the thumbnail
 * @param file The new thumbnail image file
 */
const changeThumbnail = (file: File) => {
    if (busyThumbnail.value) {
        return;
    }

    busyThumbnail.value = true;
    errorThumbnail.value = "";

    const mediaId = getNavigationStatus().media;

    makeNamedApiRequest(requestIdThumbnail, apiMediaChangeMediaThumbnail(mediaId, file))
        .onSuccess((res) => {
            showSnackBarRight($t("Successfully changed thumbnail"));

            busyThumbnail.value = false;
            thumbnail.value = res.url;

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.thumbnail = res.url;
            });

            emit("changed");

            refreshCurrentAlbum();
            emitAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, mediaId);
        })
        .onCancel(() => {
            busyThumbnail.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyThumbnail.value = false;
            handleErr(err, {
                unauthorized: errorsThumbnail.unauthorized,
                invalidThumbnail: () => {
                    errorsThumbnail.setError($t("Invalid thumbnail provided"));
                },
                badRequest: errorsThumbnail.badRequest,
                accessDenied: errorsThumbnail.accessDenied,
                notFound: errorsThumbnail.notFound,
                serverError: errorsThumbnail.notFound,
                networkError: errorsThumbnail.networkError,
            });
        })
        .onUnexpectedError((err) => {
            busyThumbnail.value = false;
            errorsThumbnail.setError(err.message);
            console.error(err);
        });
};

/**
 * Sets the current video frame as the thumbnail
 */
const setCurrentFrameAsThumbnail = () => {
    if (busyThumbnail.value) {
        return;
    }

    const videoElement = document.querySelector(".player-container .player-video-element") as HTMLVideoElement;

    if (!videoElement || videoElement.readyState !== 4) {
        errorsThumbnail.setError($t("Could not take the current frame") + ". " + $t("Maybe the media is not yet loaded?"));
    }

    try {
        // Create canvas
        const canvas = document.createElement("canvas") as HTMLCanvasElement;

        canvas.height = videoElement.videoHeight;
        canvas.width = videoElement.videoWidth;

        //  Draw video frame to the canvas
        const ctx = canvas.getContext("2d");
        ctx.drawImage(videoElement, 0, 0, canvas.width, canvas.height);

        // Get frame as blob
        canvas.toBlob((blob) => {
            // Convert to file
            const file = new File([blob], "thumbnail.png");

            // Change thumbnail
            prepareChangeThumbnail(file);
        }, "image/png");
    } catch (ex) {
        console.error(ex);
        errorsThumbnail.setError($t("Could not take the current frame") + ": " + ex.message);
    }
};

/**
 * Sets the current image as the thumbnail
 */
const setCurrentImageAsThumbnail = () => {
    if (busyThumbnail.value) {
        return;
    }

    const imageElement = document.querySelector(".player-container .player-img-element") as HTMLImageElement;

    if (!imageElement || !imageElement.complete) {
        errorsThumbnail.setError($t("Could not find the current image") + ". " + $t("Maybe the media is not yet loaded?"));
    }

    try {
        // Create canvas
        const canvas = document.createElement("canvas") as HTMLCanvasElement;

        canvas.width = imageElement.width;
        canvas.height = imageElement.height;

        //  Draw image to the canvas
        const ctx = canvas.getContext("2d");
        ctx.drawImage(imageElement, 0, 0, canvas.width, canvas.height);

        // Get image as blob
        canvas.toBlob((blob) => {
            // Convert to file
            const file = new File([blob], "thumbnail.png");

            // Change thumbnail
            prepareChangeThumbnail(file);
        }, "image/png");
    } catch (ex) {
        console.error(ex);
        errorsThumbnail.setError($t("Could not find the current image") + ": " + ex.message);
    }
};
</script>
