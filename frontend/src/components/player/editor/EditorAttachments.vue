<template>
    <div ref="container" class="player-editor-sub-content" @drop="onDrop">
        <!--- Attachments -->

        <div class="form-group">
            <label>{{ $t("Attachments") }}. {{ $t("You can attach arbitrary files for safe keeping in the vault.") }}</label>
        </div>

        <div class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-left">{{ $t("Size") }}</th>
                        <th v-if="canWrite" class="text-right td-shrink"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="attachments.length === 0">
                        <td colspan="3">
                            {{ $t("There are no attachments yet for this media.") }}
                        </td>
                        <td v-if="canWrite" class="text-right td-shrink"></td>
                    </tr>
                    <tr v-for="att in attachments" :key="att.id">
                        <td v-if="attachmentEdit != att.id" class="bold">
                            <a :href="getAssetURL(att.url)" target="_blank" rel="noopener noreferrer">{{ att.name }}</a>
                        </td>
                        <td v-if="attachmentEdit == att.id">
                            <input
                                v-model="attachmentEditName"
                                type="text"
                                maxlength="255"
                                :disabled="busyRename || busyDelete"
                                class="form-control form-control-full-width edit-auto-focus"
                                @keydown="editInputKeyEventHandler"
                            />
                        </td>
                        <td class="one-line">{{ renderSize(att.size) }}</td>
                        <td v-if="canWrite" class="text-right td-shrink one-line">
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busyRename || busyDelete"
                                @click="editAttachment(att)"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busyRename || busyDelete"
                                @click="saveEditAttachment"
                            >
                                <LoadingIcon icon="fas fa-check" :loading="busyRename"></LoadingIcon> {{ $t("Save") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busyRename || busyDelete"
                                @click="cancelEditAttachment"
                            >
                                <i class="fas fa-times"></i> {{ $t("Cancel") }}
                            </button>
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busyRename || busyDelete"
                                @click="removeAttachment(att)"
                            >
                                <LoadingIcon icon="fas fa-trash-alt" :loading="busyDelete && busyDeleteId === att.id"></LoadingIcon>
                                {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="canWrite" class="form-group">
            <input
                ref="fileInputHidden"
                type="file"
                class="file-hidden attachment-file-hidden"
                name="attachment-upload"
                @change="attachmentFileChanged"
            />
            <button
                v-if="!busyUpload || attachmentUploadProgress <= 0"
                type="button"
                class="btn btn-primary"
                :disabled="busyUpload"
                @click="selectAttachmentFile"
            >
                <i class="fas fa-upload"></i> {{ $t("Upload attachment") }}
            </button>
            <button
                v-if="busyUpload && attachmentUploadProgress > 0 && attachmentUploadProgress < 100"
                type="button"
                class="btn btn-primary"
                disabled
            >
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ attachmentUploadProgress }}%)
            </button>
            <button v-if="busyUpload && attachmentUploadProgress >= 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
        </div>

        <AttachmentDeleteModal
            v-model:display="displayAttachmentDelete"
            :attachment-to-delete="attachmentToDelete"
            @confirm="removeAttachmentConfirm"
        ></AttachmentDeleteModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="removeAttachmentConfirmInternal"
        ></AuthConfirmationModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import type { MediaAttachment } from "@/api/models";
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, nextTick, ref, useTemplateRef, watch } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { clone } from "@/utils/objects";
import { apiMediaRemoveAttachment, apiMediaRenameAttachment, apiMediaUploadAttachment } from "@/api/api-media-edit";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useRequestId } from "@/composables/use-request-id";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { renderSize } from "@/utils/size";
import { showSnackBarRight } from "@/control/snack-bar";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

const AttachmentDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AttachmentDeleteModal.vue"),
});

const AuthConfirmationModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AuthConfirmationModal.vue"),
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

// List of attachments
const attachments = ref<MediaAttachment[]>(
    (MediaController.MediaData?.attachments || []).map((a) => {
        return {
            id: a.id,
            name: a.name,
            size: a.size,
            url: a.url,
        };
    }),
);

// Progress of attachment being uploaded
const attachmentUploadProgress = ref(0);

// Selected attachment for edition
const attachmentEdit = ref(-1);

watch(canWrite, () => {
    if (!canWrite.value) {
        attachmentEdit.value = -1;
    }
});

// Name of the attachment being edited
const attachmentEditName = ref("");

/**
 * Updates media data
 */
const updateMediaData = () => {
    if (!MediaController.MediaData) {
        return;
    }

    attachments.value = (MediaController.MediaData.attachments || []).map((a) => {
        return {
            id: a.id,
            name: a.name,
            size: a.size,
            url: a.url,
        };
    });

    attachmentUploadProgress.value = 0;
    attachmentEdit.value = -1;
    attachmentEditName.value = "";
};

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, updateMediaData);

// Ref to the hidden file input element
const fileInputHidden = useTemplateRef("fileInputHidden");

/**
 * Opens the file selector
 */
const selectAttachmentFile = () => {
    const fileElem = fileInputHidden.value;
    if (fileElem) {
        fileElem.value = null;
        fileElem.click();
    }
};

/**
 * Event handler for 'changed' on the file input
 * @param e The event
 */
const attachmentFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        const file = data[0];
        addAttachment(file);
    }
};

/**
 * Event handler for 'drop', in order to catch files
 * @param e The event
 */
const onDrop = (e: DragEvent) => {
    e.preventDefault();
    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        const file = data[0];
        addAttachment(file);
    }
};

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

// Request ID
const requestId = useRequestId();

// Uploading a file?
const busyUpload = ref(false);

/**
 * Uploads an attachment
 * @param file The attachment file
 */
const addAttachment = (file: File) => {
    if (busyUpload.value) {
        return;
    }

    busyUpload.value = true;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestId, apiMediaUploadAttachment(mediaId, file))
        .onSuccess((res) => {
            showSnackBarRight($t("Added attachment") + ": " + res.name);

            busyUpload.value = false;
            attachmentUploadProgress.value = 0;
            attachments.value.push(res);

            if (MediaController.MediaData) {
                MediaController.MediaData.attachments = clone(attachments.value);
            }

            emit("changed");
        })
        .onUploadProgress((loaded, total) => {
            if (total) {
                attachmentUploadProgress.value = Math.floor(((loaded * 100) / total) * 100) / 100;
            }
        })
        .onCancel(() => {
            busyUpload.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyUpload.value = false;

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
            setError(err.message);
            console.error(err);
            busyUpload.value = false;
        });
};

/**
 * Starts editing an attachment
 * @param att The attachment
 */
const editAttachment = (att: MediaAttachment) => {
    attachmentEdit.value = att.id;
    attachmentEditName.value = att.name;

    nextTick(() => {
        const el = container.value?.querySelector(".edit-auto-focus") as HTMLInputElement;
        if (el) {
            el.focus();
            el.select();
        }
    });
};

/**
 * Cancels the edition of an attachment
 */
const cancelEditAttachment = () => {
    attachmentEdit.value = -1;
    attachmentEditName.value = "";
};

/**
 * Event handler for 'keydown'
 * on the attachment name input that is being edited
 * @param e The keyboard event
 */
const editInputKeyEventHandler = (e: KeyboardEvent) => {
    if (e.key === "Enter") {
        e.preventDefault();
        saveEditAttachment();
    }
};

// Busy renaming?
const busyRename = ref(false);

/**
 * Performs the request to edit an attachment
 */
const saveEditAttachment = () => {
    if (busyRename.value) {
        return;
    }

    busyRename.value = true;

    const mediaId = AppStatus.CurrentMedia;
    const id = attachmentEdit.value;

    makeNamedApiRequest(requestId, apiMediaRenameAttachment(mediaId, id, attachmentEditName.value))
        .onSuccess((res) => {
            showSnackBarRight($t("Renamed attachment") + ": " + res.name);

            busyRename.value = false;
            attachmentEdit.value = -1;
            attachmentEditName.value = "";

            for (let i = 0; i < attachments.value.length; i++) {
                if (attachments.value[i].id === res.id) {
                    attachments.value[i].name = res.name;
                    attachments.value[i].url = res.url;
                    break;
                }
            }

            if (MediaController.MediaData) {
                MediaController.MediaData.attachments = clone(attachments.value);
            }

            emit("changed");
        })
        .onCancel(() => {
            busyRename.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyRename.value = false;

            handleErr(err, {
                unauthorized,
                invalidName: () => {
                    setError($t("Invalid attachment name"));
                },
                badRequest,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            setError(err.message);
            console.error(err);
            busyRename.value = false;
        });
};

// Display deletion confirmation modal
const displayAttachmentDelete = ref(false);

// Reference to the attachment to delete
const attachmentToDelete = ref<MediaAttachment | null>(null);

/**
 * Removes an attachment
 * @param att The attachment
 */
const removeAttachment = (att: MediaAttachment) => {
    attachmentToDelete.value = att;
    displayAttachmentDelete.value = true;
};

/**
 * The user confirmed the deletion
 */
const removeAttachmentConfirm = () => {
    removeAttachmentConfirmInternal({});
};

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

// Deleting in progress
const busyDelete = ref(false);

// ID of the attachment being deleted
const busyDeleteId = ref(-1);

/**
 * Performs the request to delete the attachment
 * @param confirmation The auth confirmation
 */
const removeAttachmentConfirmInternal = (confirmation: ProvidedAuthConfirmation) => {
    const att = attachmentToDelete.value;

    if (busyDelete.value || !att) {
        return;
    }

    busyDelete.value = true;
    busyDeleteId.value = att.id;

    const mediaId = AppStatus.CurrentMedia;
    const id = att.id;

    makeNamedApiRequest(requestId, apiMediaRemoveAttachment(mediaId, id, confirmation))
        .onSuccess(() => {
            showSnackBarRight($t("Removed attachment") + ": " + att.name);

            busyDelete.value = false;
            for (let i = 0; i < attachments.value.length; i++) {
                if (attachments.value[i].id === id) {
                    attachments.value.splice(i, 1);
                    break;
                }
            }
            if (MediaController.MediaData) {
                MediaController.MediaData.attachments = clone(attachments.value);
            }

            emit("changed");
        })
        .onCancel(() => {
            busyDelete.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyDelete.value = false;

            handleErr(err, {
                unauthorized,
                badRequest,
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            setError(err.message);
            console.error(err);
            busyDelete.value = false;
        });
};
</script>
