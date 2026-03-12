<template>
    <div ref="container" class="player-editor-sub-content" @drop="onDrop">
        <!--- Subtitles -->

        <div class="form-group">
            <label>{{ $t("Subtitles") }}:</label>
        </div>

        <div class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("ID") }}</th>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-right td-shrink"></th>
                        <th v-if="canWrite" class="text-right td-shrink"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="subtitles.length === 0">
                        <td colspan="3">
                            {{ $t("There are no subtitles yet for this media.") }}
                        </td>
                        <td v-if="canWrite" class="text-right td-shrink"></td>
                    </tr>
                    <tr v-for="sub in subtitles" :key="sub.id">
                        <td v-if="subtitleRenameSelected !== sub.id" class="bold">{{ sub.id }}</td>
                        <td v-else>
                            <input
                                v-model="subtitleRenameId"
                                type="text"
                                maxlength="255"
                                :disabled="busy || busyDeleting || subtitleRenameBusy"
                                class="form-control form-control-full-width"
                                @keydown="renameInputKeyEventHandler"
                            />
                        </td>

                        <td v-if="subtitleRenameSelected !== sub.id" class="bold">{{ sub.name }}</td>
                        <td v-else>
                            <input
                                v-model="subtitleRenameName"
                                type="text"
                                maxlength="255"
                                :disabled="busy || busyDeleting || subtitleRenameBusy"
                                class="form-control form-control-full-width edit-auto-focus"
                                @keydown="renameInputKeyEventHandler"
                            />
                        </td>

                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" @click="downloadSubtitles(sub)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>

                        <td v-if="subtitleRenameSelected === sub.id && canWrite" class="text-right td-shrink one-line">
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || subtitleRenameBusy"
                                @click="saveRename"
                            >
                                <LoadingIcon icon="fas fa-check" :loading="subtitleRenameBusy"></LoadingIcon> {{ $t("Save") }}
                            </button>
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || subtitleRenameBusy"
                                @click="cancelRename"
                            >
                                <i class="fas fa-times"></i> {{ $t("Cancel") }}
                            </button>
                        </td>
                        <td v-else-if="canWrite" class="text-right td-shrink one-line">
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || subtitleRenameBusy"
                                @click="startRename(sub)"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy || busyDeleting || subtitleRenameBusy"
                                @click="removeSubtitles(sub)"
                            >
                                <LoadingIcon icon="fas fa-trash-alt" :loading="busyDeleting && busyDeletingId === sub.id"></LoadingIcon>
                                {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="canWrite">
            <div class="form-group">
                <label>{{ $t("You can upload subtitles in SubRip format (.srt)") }}:</label>
                <input
                    ref="fileInputHidden"
                    type="file"
                    class="file-hidden srt-file-hidden"
                    name="srt-upload"
                    accept=".srt"
                    @change="srtFileChanged"
                />
                <button v-if="!srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                    <i class="fas fa-upload"></i> {{ $t("Select SRT file") }}
                </button>

                <button v-if="srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                    <i class="fas fa-upload"></i> {{ $t("SRT file") }}: {{ srtFileName }}
                </button>
            </div>
            <div class="form-group">
                <label>{{ $t("Subtitles identifier") }}:</label>
                <input v-model="srtId" type="text" autocomplete="off" maxlength="255" :disabled="busy" class="form-control" />
            </div>
            <div class="form-group">
                <label>{{ $t("Subtitles name") }}:</label>
                <input v-model="srtName" type="text" autocomplete="off" maxlength="255" :disabled="busy" class="form-control" />
            </div>
            <div class="form-group">
                <button v-if="!busy" type="button" class="btn btn-primary" :disabled="!srtId || !srtName || !srtFile" @click="addSubtitles">
                    <i class="fas fa-plus"></i> {{ $t("Add subtitles file") }}
                </button>
                <button v-else-if="uploading" type="button" class="btn btn-primary" disabled>
                    <i class="fa fa-spinner fa-spin"></i>
                    {{ $t("Uploading") + "..." + (uploadProgress > 0 ? " (" + renderProgress(uploadProgress) + ")" : "") }}
                </button>
                <button v-else type="button" class="btn btn-primary" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Processing") + "..." }}
                </button>
            </div>
        </div>

        <SubtitlesDeleteModal
            v-model:display="displaySubtitlesDelete"
            :subtitle-to-delete="subtitleToDelete"
            @confirm="removeSubtitlesConfirm"
        ></SubtitlesDeleteModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="removeSubtitlesConfirmInternal"
        ></AuthConfirmationModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import type { MediaSubtitle } from "@/api/models";
import { EVENT_NAME_MEDIA_UPDATE } from "@/global-state/app-events";
import { AppStatus } from "@/global-state/app-status";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, nextTick, ref, shallowRef, useTemplateRef, watch } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { clone } from "@/utils/objects";
import { apiMediaRemoveSubtitles, apiMediaRenameSubtitles, apiMediaSetSubtitles } from "@/api/api-media-edit";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { stringMultiReplace } from "@/utils/string-multi-replace";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useRequestId } from "@/composables/use-request-id";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
import { showSnackBarRight } from "@/global-state/snack-bar";
import { getCurrentMediaData, modifyCurrentMediaData } from "@/global-state/media";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

const SubtitlesDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SubtitlesDeleteModal.vue"),
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

// List of subtitles
const subtitles = ref<MediaSubtitle[]>(
    (getCurrentMediaData()?.subtitles || []).map((a) => {
        return {
            id: a.id,
            name: a.name,
            url: a.url,
        };
    }),
);

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, (mediaData) => {
    if (!mediaData) {
        return;
    }

    subtitles.value = (mediaData.subtitles || []).map((a) => {
        return {
            id: a.id,
            name: a.name,
            url: a.url,
        };
    });
});

// SRT file
const srtFile = shallowRef<File | null>(null);

// Subtitles file name
const srtFileName = ref("");

// Subtitles ID
const srtId = ref("");

// Subtitles name
const srtName = ref("");

// Hidden file input
const fileInputHidden = useTemplateRef("fileInputHidden");

/**
 * Opens the file selector
 */
const selectSRTFile = () => {
    const fileElem = fileInputHidden.value;
    if (fileElem) {
        fileElem.value = null;
        fileElem.click();
    }
};

/**
 * Sets the subtitles file to be uploaded
 * @param file The file
 */
const setFile = (file: File) => {
    srtFile.value = file;
    srtFileName.value = file.name;
    srtId.value = (file.name.split(".")[0] || "").toLowerCase();
    srtName.value = srtId.value.toUpperCase();
};

/**
 * Event handler for 'change' on the file input
 * @param e The event
 */
const srtFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        setFile(data[0]);
    }
};

/**
 * Event handler for 'drop' of a file
 * @param e The event
 */
const onDrop = (e: DragEvent) => {
    e.preventDefault();
    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        setFile(data[0]);
    }
};

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

// Busy uploading a new file?
const busy = ref(false);

// Request ID for adding
const requestIdAdd = useRequestId();

// True if uploading
const uploading = ref(false);

// Upload progress
const uploadProgress = ref(0);

/**
 * Adds a new subtitles file
 */
const addSubtitles = () => {
    if (!srtFile.value) {
        setError($t("Please, select a SubRip file first"));
        return;
    }

    const id = srtId.value;
    const name = srtName.value;

    let duped = false;
    for (const sub of subtitles.value) {
        if (sub.id === id) {
            duped = true;
            break;
        }
    }

    if (duped) {
        setError($t("There is already another subtitles file with the same identifier"));
        return;
    }

    if (busy.value) {
        return;
    }

    busy.value = true;
    uploading.value = true;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestIdAdd, apiMediaSetSubtitles(mediaId, id, name, srtFile.value))
        .onSuccess((res) => {
            showSnackBarRight($t("Added subtitles") + ": " + res.name);

            busy.value = false;
            subtitles.value.push(res);

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.subtitles = clone(subtitles.value);
            });

            emit("changed");
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onUploadProgress((loaded, total) => {
            uploadProgress.value = loaded / Math.max(1, total);
            uploading.value = loaded < total;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidSRT: () => {
                    setError($t("Invalid SubRip file"));
                },
                invalidId: () => {
                    setError($t("Invalid subtitles identifier"));
                },
                invalidName: () => {
                    setError($t("Invalid subtitles name"));
                },
                badRequest,
                accessDenied,
                notFound,
                fileTooLarge: () => {
                    setError(stringMultiReplace($t("Subtitles file too big (max is $MAX)"), { $MAX: "10MB" }));
                },
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

// Selected subtitles for rename
const subtitleRenameSelected = ref("");

// Id of the subtitles being renamed (new)
const subtitleRenameId = ref("");

// Name of the subtitles being renamed (new)
const subtitleRenameName = ref("");

// True if renaming an subtitles
const subtitleRenameBusy = ref(false);

// Request ID for renaming
const requestIdRename = useRequestId();

/**
 * Starts renaming a subtitle
 * @param sub The subtitle
 */
const startRename = (sub: MediaSubtitle) => {
    subtitleRenameSelected.value = sub.id;
    subtitleRenameId.value = sub.id;
    subtitleRenameName.value = sub.name;

    nextTick(() => {
        const el = container.value?.querySelector(".edit-auto-focus") as HTMLInputElement;
        if (el) {
            el.focus();
            el.select();
        }
    });
};

/**
 * Cancels renaming
 */
const cancelRename = () => {
    subtitleRenameSelected.value = "";
};

watch(canWrite, () => {
    if (!canWrite.value) {
        cancelRename();
    }
});

/**
 * Event handler for 'keydown' on the rename inputs
 * @param e The keyboard event
 */
const renameInputKeyEventHandler = (e: KeyboardEvent) => {
    if (e.key === "Enter") {
        e.preventDefault();
        saveRename();
    }
};

/**
 * Performs the request to rename a subtitles file
 */
const saveRename = () => {
    if (subtitleRenameBusy.value) {
        return;
    }

    const subtitleId = subtitleRenameSelected.value;

    const newId = subtitleRenameId.value;
    const newName = subtitleRenameName.value;

    if (!newId) {
        setError($t("Invalid subtitles identifier"));
        return;
    }

    if (newId !== subtitleId) {
        for (const subtitle of subtitles.value) {
            if (subtitle.id === newId) {
                setError($t("Subtitles identifier already in use"));
                return;
            }
        }
    }

    if (!newName) {
        setError($t("Invalid subtitles name"));
        return;
    }

    subtitleRenameBusy.value = true;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestIdRename, apiMediaRenameSubtitles(mediaId, subtitleId, newId, newName))
        .onSuccess(() => {
            showSnackBarRight($t("Renamed subtitles") + ": " + newName + " (" + newId + ")");

            subtitleRenameBusy.value = false;
            subtitleRenameSelected.value = "";

            for (let i = 0; i < subtitles.value.length; i++) {
                if (subtitles.value[i].id === subtitleId) {
                    subtitles.value[i].id = newId;
                    subtitles.value[i].name = newName;
                    break;
                }
            }

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.subtitles = clone(subtitles.value);
            });

            emit("changed");
        })
        .onCancel(() => {
            subtitleRenameBusy.value = false;
        })
        .onRequestError((err, handleErr) => {
            subtitleRenameBusy.value = false;

            handleErr(err, {
                unauthorized,
                invalidId: () => {
                    setError($t("Invalid subtitles identifier"));
                },
                invalidName: () => {
                    setError($t("Invalid subtitles name"));
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
            subtitleRenameBusy.value = false;
        });
};

// Display deletion confirmation?
const displaySubtitlesDelete = ref(false);

// Reference to the subtitle to delete
const subtitleToDelete = ref<MediaSubtitle | null>(null);

/**
 * Opens the modal to confirm the deletion of a file
 * @param sub The subtitles file
 */
const removeSubtitles = (sub: MediaSubtitle) => {
    subtitleToDelete.value = sub;
    displaySubtitlesDelete.value = true;
};

/**
 * The user confirmed to delete a subtitles file
 */
const removeSubtitlesConfirm = () => {
    removeSubtitlesConfirmInternal({});
};

// Request ID to delete an audio
const requestIdDelete = useRequestId();

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
const busyDeleting = ref(false);

// ID of the attachment being deleted
const busyDeletingId = ref("");

/**
 * Performs the request to delete a subtitles file
 * @param confirmation The auth confirmation
 */
const removeSubtitlesConfirmInternal = (confirmation: ProvidedAuthConfirmation) => {
    const sub = subtitleToDelete.value;

    if (busyDeleting.value || !sub) {
        return;
    }

    busyDeleting.value = true;
    busyDeletingId.value = sub.id;

    const mediaId = AppStatus.CurrentMedia;
    const id = sub.id;

    makeNamedApiRequest(requestIdDelete, apiMediaRemoveSubtitles(mediaId, id, confirmation))
        .onSuccess(() => {
            showSnackBarRight($t("Removed subtitles") + ": " + sub.name);

            busyDeleting.value = false;

            for (let i = 0; i < subtitles.value.length; i++) {
                if (subtitles.value[i].id === id) {
                    subtitles.value.splice(i, 1);
                    break;
                }
            }

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.subtitles = clone(subtitles.value);
            });

            emit("changed");
        })
        .onCancel(() => {
            busyDeleting.value = false;
        })
        .onRequestError((err, handleErr) => {
            busyDeleting.value = false;

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
            busyDeleting.value = false;
        });
};

/**
 * Downloads a subtitles file
 * @param sub The subtitles file
 */
const downloadSubtitles = (sub: MediaSubtitle) => {
    const link = document.createElement("a");
    link.target = "_blank";
    link.rel = "noopener noreferrer";

    const url = getAssetURL(sub.url);
    const titlePart = sub.id ? "&filename=" + encodeURIComponent(sub.id) : "";

    if (url.includes("?")) {
        link.href = url + "&download=force" + titlePart;
    } else {
        link.href = url + "?download=force" + titlePart;
    }

    link.click();
};

/**
 * Renders progress as percentage
 * @param p The progress (0-1)
 * @returns The progress as percentage
 */
const renderProgress = (p: number): string => {
    return Math.max(0, Math.min(100, Math.floor(p * 100))) + "%";
};
</script>
