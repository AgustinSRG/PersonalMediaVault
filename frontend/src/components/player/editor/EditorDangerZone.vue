<template>
    <div class="player-editor-sub-content">
        <!--- Re-Encode -->

        <div class="form-group">
            <label>
                {{ $t("If the media resource did not encode properly, try using the button below.") }}
                {{ $t("If it still does not work, try re-uploading the media.") }}
            </label>
        </div>
        <div class="form-group">
            <button type="button" class="btn btn-primary" :disabled="busyReEncode || busyReplace" @click="encodeMedia">
                <LoadingIcon icon="fas fa-sync-alt" :loading="busyReEncode"></LoadingIcon> {{ $t("Re-Encode") }}
            </button>
            <div v-if="errorReEncode" class="form-error form-error-pt">{{ $t("Error") }}: {{ errorReEncode }}</div>
        </div>

        <!--- Replace -->

        <div class="form-group">
            <label>
                {{ $t("If you want to replace the media file, try using the button below.") }}
                {{ $t("You can use it to upgrade the media quality or fix any issues it may have.") }}
            </label>
        </div>
        <div class="form-group">
            <input
                ref="replaceFileHiddenInput"
                type="file"
                class="file-hidden replace-file-hidden"
                name="attachment-upload"
                @change="replaceFileChanged"
            />
            <button
                v-if="!replacing"
                type="button"
                class="btn btn-primary"
                :disabled="busyReEncode || busyReplace"
                @click="replaceMedia"
                @drop="replaceMediaDrop"
            >
                <i class="fas fa-upload"></i> {{ $t("Replace media") }}
            </button>
            <button v-else-if="replacing && replaceProgress < 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ replaceProgress }}%)
            </button>
            <button v-else type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
            <div v-if="errorReplace" class="form-error form-error-pt">{{ $t("Error") }}: {{ errorReplace }}</div>
        </div>

        <!--- Delete -->

        <div class="form-group">
            <label>{{ $t("If you want to delete this media resource, click the button below.") }}</label>
        </div>
        <div class="form-group">
            <button type="button" class="btn btn-danger" :disabled="busyReEncode || busyReplace" @click="deleteMedia">
                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
            </button>
        </div>

        <MediaDeleteModal v-if="displayMediaDelete" v-model:display="displayMediaDelete"></MediaDeleteModal>

        <ReEncodeConfirmationModal
            v-if="displayReEncode"
            v-model:display="displayReEncode"
            @confirm="doEncodeMedia"
        ></ReEncodeConfirmationModal>

        <ReplaceMediaConfirmationModal
            v-if="displayReplace"
            v-model:display="displayReplace"
            :file-name="replaceFileName"
            :file-size="replaceFileSize"
            @confirm="doReplaceMedia"
        ></ReplaceMediaConfirmationModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="performReplaceMediaRequest"
        ></AuthConfirmationModal>
    </div>
</template>

<script setup lang="ts">
import { AppStatus } from "@/global-state/app-status";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, ref, shallowRef, useTemplateRef } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { apiMediaEncodeMedia, apiMediaReplaceMedia } from "@/api/api-media-edit";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
import { showSnackBarRight } from "@/global-state/snack-bar";
import { loadCurrentMedia } from "@/global-state/media";

const ReEncodeConfirmationModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ReEncodeConfirmationModal.vue"),
});

const ReplaceMediaConfirmationModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ReplaceMediaConfirmationModal.vue"),
});

const MediaDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/MediaDeleteModal.vue"),
});

const AuthConfirmationModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AuthConfirmationModal.vue"),
});

// Translation
const { $t } = useI18n();

// Request ID
const requestId = useRequestId();

// True to display the re-encode confirmation
const displayReEncode = ref(false);

/**
 * Displays the modal to confirm re-encoding the media
 */
const encodeMedia = () => {
    displayReEncode.value = true;
};

// Error for media re-encode
const commonErrorsReEncode = useCommonRequestErrors();
const errorReEncode = commonErrorsReEncode.error;

// Busy re-encoding?
const busyReEncode = ref(false);

/**
 * Performs the request to re-encode the media
 */
const doEncodeMedia = () => {
    if (busyReEncode.value) {
        return;
    }

    busyReEncode.value = true;
    errorReEncode.value = "";

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestId, apiMediaEncodeMedia(mediaId))
        .onSuccess(() => {
            showSnackBarRight($t("Successfully requested pending encoding tasks"));

            busyReEncode.value = false;

            loadCurrentMedia();
        })
        .onCancel(() => {
            busyReEncode.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyReEncode.value = false;

            handleErr(err, commonErrorsReEncode);
        })
        .onUnexpectedError((err) => {
            busyReEncode.value = false;
            commonErrorsReEncode.setError(err.message);
            console.error(err);
        });
};

// Reference to the file for media replacement
const fileRef = shallowRef<File | null>(null);

// True to display the modal to replace the media
const displayReplace = ref(false);

// Name of the file to replace the media
const replaceFileName = ref("");

// Size of the file to replace the media
const replaceFileSize = ref(0);

// Reference to the hidden file input to replace media
const replaceFileHiddenInput = useTemplateRef("replaceFileHiddenInput");

/**
 * The user clicked in the "Replace media" button.
 * Open the file selector.
 */
const replaceMedia = () => {
    const fileElem = replaceFileHiddenInput.value;
    if (fileElem) {
        fileElem.value = null;
        fileElem.click();
    }
};

/**
 * Event handler for 'change' on the file input
 * @param e The event
 */
const replaceFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        const file = data[0] as File;
        fileRef.value = file;
        replaceFileName.value = file.name;
        replaceFileSize.value = file.size;
        displayReplace.value = true;
    }
};

/**
 * Event handler for 'drop' on the 'replace media' option
 * @param e The event
 */
const replaceMediaDrop = (e: DragEvent) => {
    e.preventDefault();

    if (busyReplace.value) {
        return;
    }

    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        const file = data[0] as File;
        fileRef.value = file;
        replaceFileName.value = file.name;
        replaceFileSize.value = file.size;
        displayReplace.value = true;
    }
};

// True if replacing
const replacing = ref(false);

// Progress uploading the file to replace
const replaceProgress = ref(0);

// Error for media replace
const commonErrorsReplace = useCommonRequestErrors();
const errorReplace = commonErrorsReplace.error;

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

// Busy replacing media
const busyReplace = ref(false);

/**
 * Replaces the media with the selected file
 */
const doReplaceMedia = () => {
    performReplaceMediaRequest({});
};

/**
 * Performs the request to replace the media
 * @param confirmation The auth confirmation
 */
const performReplaceMediaRequest = (confirmation: ProvidedAuthConfirmation) => {
    if (busyReplace.value) {
        return;
    }

    const file = fileRef.value;

    if (!file) {
        return;
    }

    busyReplace.value = true;
    replacing.value = true;
    replaceProgress.value = 0;
    errorReplace.value = "";

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestId, apiMediaReplaceMedia(mediaId, file, confirmation))
        .onSuccess(() => {
            showSnackBarRight($t("Successfully uploaded") + ": " + file.name);

            busyReplace.value = false;
            replacing.value = false;
            replaceProgress.value = 0;
            fileRef.value = null;

            loadCurrentMedia();
        })
        .onUploadProgress((loaded, total) => {
            if (total) {
                replaceProgress.value = Math.floor(((loaded * 100) / total) * 100) / 100;
            }
        })
        .onCancel(() => {
            busyReplace.value = false;
            replacing.value = false;
            replaceProgress.value = 0;
            fileRef.value = null;
        })
        .onRequestError((err, handleErr) => {
            busyReplace.value = false;
            replacing.value = false;
            replaceProgress.value = 0;

            handleErr(err, {
                unauthorized: commonErrorsReplace.unauthorized,
                invalidMedia: () => {
                    fileRef.value = null;
                    commonErrorsReplace.setError($t("Invalid media file provided"));
                },
                invalidMediaType: () => {
                    fileRef.value = null;
                    commonErrorsReplace.setError($t("You must upload a file of the same type in order to replace the media"));
                },
                badRequest: commonErrorsReplace.badRequest,
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied: commonErrorsReplace.accessDenied,
                notFound: commonErrorsReplace.notFound,
                serverError: commonErrorsReplace.serverError,
                networkError: commonErrorsReplace.networkError,
            });
        })
        .onUnexpectedError((err) => {
            busyReplace.value = false;
            replacing.value = false;
            replaceProgress.value = 0;

            commonErrorsReplace.setError(err.message);
            console.error(err);
        });
};

// True to display the modal to delete the media
const displayMediaDelete = ref(false);

/**
 * Displays the modal to delete the media
 */
const deleteMedia = () => {
    displayMediaDelete.value = true;
};
</script>
