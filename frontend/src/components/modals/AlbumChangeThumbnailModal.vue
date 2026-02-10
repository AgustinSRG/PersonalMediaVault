<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <div v-if="display" class="modal-dialog modal-md" role="document" @drop="onDrop">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Change album thumbnail") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group flex-center">
                    <label v-if="!thumbnail">{{ $t("No thumbnail set for this album") }}</label>
                    <ThumbImage v-if="thumbnail" :src="getAssetURL(thumbnail)" class-name="form-group-thumbnail"></ThumbImage>
                </div>
                <div class="form-group">
                    <input ref="fileSelector" type="file" class="file-hidden" name="thumbnail-upload" @change="inputFileChanged" />
                    <div class="text-center">
                        <button v-if="!busy" type="button" class="btn btn-primary image-thumbnail-button" @click="uploadThumbnail">
                            <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
                        </button>
                        <button v-if="busy" type="button" class="btn btn-primary image-thumbnail-button" disabled>
                            <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading thumbnail") }}...
                        </button>
                    </div>
                    <div v-if="currentMediaThumbnail" class="text-center form-group-pt">
                        <button
                            type="button"
                            class="btn btn-primary btn-sm image-thumbnail-button"
                            :title="$t('Set current media thumbnail')"
                            :disabled="busy"
                            @click="setCurrentMediaThumbnail"
                        >
                            <i class="fas fa-image"></i> {{ $t("Set current media thumbnail") }}
                        </button>
                    </div>
                    <div v-if="error" class="form-error form-error-pt text-center">{{ $t("Error") + ": " + error }}</div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="button" class="modal-footer-btn" @click="close">
                    <i class="fas fa-check"></i> {{ $t("Done") }}
                </button>
            </div>
        </div>

        <ThumbnailCropModal
            v-if="displayThumbnailModal"
            v-model:display="displayThumbnailModal"
            :image-url="thumbnailModalUrl"
            @done="changeThumbnail"
            @error="onThumbnailModalError"
        ></ThumbnailCropModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { AlbumsController } from "@/control/albums";
import { EVENT_NAME_CURRENT_ALBUM_UPDATED, EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { onBeforeUnmount, ref, useTemplateRef, watch } from "vue";
import { AuthController, SESSION_TOKEN_HEADER_NAME } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsChangeAlbumThumbnail } from "@/api/api-albums";
import { getAssetURL } from "@/utils/api";
import { MediaController } from "@/control/media";
import ThumbImage from "../utils/ThumbImage.vue";
import ThumbnailCropModal from "./ThumbnailCropModal.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Current album ID
const currentAlbum = ref(AlbumsController.CurrentAlbum);

// Current album thumbnail
const thumbnail = ref(AlbumsController.CurrentAlbumData?.thumbnail || "");

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, () => {
    currentAlbum.value = AlbumsController.CurrentAlbum;
    thumbnail.value = AlbumsController.CurrentAlbumData?.thumbnail || "";
});

// Current media thumbnail
const currentMediaThumbnail = ref(MediaController.MediaData?.thumbnail || "");

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, () => {
    currentMediaThumbnail.value = MediaController.MediaData?.thumbnail || "";
});

// Temporal image
let tempImage: HTMLImageElement | null = null;

// Abort controller for thumbnail image request
let fetchMediaThumbnailAbortController: AbortController | null = null;

onBeforeUnmount(() => {
    if (fetchMediaThumbnailAbortController) {
        fetchMediaThumbnailAbortController.abort();
    }

    if (tempImage) {
        delete tempImage.onload;
        delete tempImage.onerror;
    }
});

// Busy state
const busy = ref(false);

// Request error
const { error, unauthorized, badRequest, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Resets the error messages
const resetErrors = () => {
    error.value = "";
};

watch(display, () => {
    if (display.value) {
        resetErrors();
    }
});

/**
 * Called when the thumbnail crop modal results in
 * an error. This means the thumbnail image is invalid.
 */
const onThumbnailModalError = () => {
    error.value = $t("Invalid thumbnail provided");
};

// Request ID
const changeThumbnailRequestId = useRequestId();

/**
 * Changes the album thumbnail
 * @param file The thumbnail file
 */
const changeThumbnail = (file: File) => {
    if (busy.value) {
        return;
    }

    resetErrors();

    busy.value = true;

    const albumId = currentAlbum.value;

    makeNamedApiRequest(changeThumbnailRequestId, apiAlbumsChangeAlbumThumbnail(albumId, file))
        .onSuccess((res) => {
            PagesController.ShowSnackBarRight($t("Successfully changed thumbnail"));

            busy.value = false;

            thumbnail.value = res.url;

            AlbumsController.OnChangedAlbum(albumId);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidThumbnail: () => {
                    error.value = $t("Invalid thumbnail provided");
                },
                badRequest,
                accessDenied,
                notFound: () => {
                    error.value = $t("Not found");
                },
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            error.value = err.message;

            console.error(err);
        });
};

// True to display the thumbnail crop modal
const displayThumbnailModal = ref(false);

// Thumbnail URL to be loaded in the crop modal
const thumbnailModalUrl = ref("");

// Reference to the file selector element
const fileSelector = useTemplateRef("fileSelector");

/**
 * Opens the file selector in order to select a thumbnail
 * to be uploaded
 */
const uploadThumbnail = () => {
    if (fileSelector.value) {
        fileSelector.value.value = null;
        fileSelector.value.click();
    }
};

/**
 * Event handler for 'change' on the file selector element
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
 * Event handler for 'drop' on the component
 * @param e The event
 */
const onDrop = (e: DragEvent) => {
    e.preventDefault();

    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        const file = data[0];
        prepareChangeThumbnail(file);
    }
};

/**
 * Prepares a file to be uploaded as a thumbnail
 * @param file The file
 */
const prepareChangeThumbnail = (file: File) => {
    if (tempImage) {
        delete tempImage.onload;
        delete tempImage.onerror;
        tempImage = null;
    }

    const url = URL.createObjectURL(file);

    const img = new Image();
    tempImage = img;

    img.onload = () => {
        tempImage = null;

        if (img.width === img.height) {
            changeThumbnail(file);
            return;
        }

        thumbnailModalUrl.value = url;
        error.value = "";
        displayThumbnailModal.value = true;
    };

    img.onerror = (err) => {
        tempImage = null;
        console.error(err);
        onThumbnailModalError();
    };

    img.src = url;
};

/**
 * Uses the thumbnail of the current media
 * as the thumbnail for the album
 */
const setCurrentMediaThumbnail = () => {
    if (busy.value) {
        return;
    }

    if (!currentMediaThumbnail.value) {
        return;
    }

    resetErrors();

    const thumbnailUrl = getAssetURL(currentMediaThumbnail.value);

    busy.value = true;

    const abortController = new AbortController();
    fetchMediaThumbnailAbortController = abortController;

    fetch(thumbnailUrl, {
        signal: abortController.signal,
        headers: {
            [SESSION_TOKEN_HEADER_NAME]: AuthController.Session,
        },
    })
        .then((response) => {
            fetchMediaThumbnailAbortController = null;

            if (response.status === 200) {
                response
                    .blob()
                    .then((blob) => {
                        busy.value = false;
                        changeThumbnail(new File([blob], "thumbnail.jpg"));
                    })
                    .catch((err) => {
                        console.error(err);
                        error.value = err.message;
                        busy.value = false;
                    });
            } else {
                error.value = $t("Could not fetch media thumbnail");
                busy.value = false;
            }
        })
        .catch((err) => {
            fetchMediaThumbnailAbortController = null;
            console.error(err);
            error.value = err.message;
            busy.value = false;
        });
};
</script>
