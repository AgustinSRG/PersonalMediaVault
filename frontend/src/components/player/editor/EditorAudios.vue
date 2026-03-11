<template>
    <div ref="container" class="player-editor-sub-content" @drop="onDrop">
        <!--- Audio tracks -->

        <div class="form-group">
            <label>{{ $t("Extra audio tracks for the video.") }} {{ $t("You can use this to add multiple audio languages.") }}</label>
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
                    <tr v-if="audios.length === 0">
                        <td colspan="3">
                            {{ $t("There are no audio tracks yet for this media.") }}
                        </td>
                        <td v-if="canWrite" class="text-right td-shrink"></td>
                    </tr>
                    <tr v-for="aud in audios" :key="aud.id">
                        <td v-if="audioRenameSelected !== aud.id" class="bold">{{ aud.id }}</td>
                        <td v-else>
                            <input
                                v-model="audioRenameId"
                                type="text"
                                maxlength="255"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                class="form-control form-control-full-width"
                                @keydown="renameInputKeyEventHandler"
                            />
                        </td>

                        <td v-if="audioRenameSelected !== aud.id" class="bold">{{ aud.name }}</td>
                        <td v-else>
                            <input
                                v-model="audioRenameName"
                                type="text"
                                maxlength="255"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                class="form-control form-control-full-width edit-auto-focus"
                                @keydown="renameInputKeyEventHandler"
                            />
                        </td>

                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" @click="downloadAudio(aud)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td v-if="audioRenameSelected === aud.id && canWrite" class="text-right td-shrink one-line">
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                @click="saveRename"
                            >
                                <LoadingIcon icon="fas fa-check" :loading="audioRenameBusy"></LoadingIcon> {{ $t("Save") }}
                            </button>
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                @click="cancelRename"
                            >
                                <i class="fas fa-times"></i> {{ $t("Cancel") }}
                            </button>
                        </td>
                        <td v-else-if="canWrite" class="text-right td-shrink one-line">
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                @click="startRename(aud)"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                @click="removeAudio(aud)"
                            >
                                <LoadingIcon icon="fas fa-trash-alt" :loading="busyDeleting && busyDeletingId === aud.id"></LoadingIcon>
                                {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="canWrite">
            <div class="form-group">
                <label>{{ $t("You can upload extra audio tracks for the video (.mp3)") }}:</label>
                <input
                    ref="fileInputHidden"
                    type="file"
                    class="file-hidden audio-file-hidden"
                    name="mp3-upload"
                    accept=".mp3"
                    @change="audioFileChanged"
                />
                <button v-if="!audioFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectAudioFile">
                    <i class="fas fa-upload"></i> {{ $t("Select audio file") }}
                </button>

                <button v-if="audioFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectAudioFile">
                    <i class="fas fa-upload"></i> {{ $t("Audio file") }}: {{ audioFileName }}
                </button>
            </div>
            <div class="form-group">
                <label>{{ $t("Audio track identifier") }}:</label>
                <input v-model="audioId" type="text" autocomplete="off" maxlength="255" :disabled="busy" class="form-control" />
            </div>
            <div class="form-group">
                <label>{{ $t("Audio track name") }}:</label>
                <input v-model="audioName" type="text" autocomplete="off" maxlength="255" :disabled="busy" class="form-control" />
            </div>
            <div class="form-group">
                <button
                    v-if="!busy"
                    type="button"
                    class="btn btn-primary"
                    :disabled="!audioId || !audioName || !audioFile"
                    @click="addAudio"
                >
                    <i class="fas fa-plus"></i> {{ $t("Add audio track file") }}
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

        <AudioTrackDeleteModal
            v-model:display="displayAudioTrackDelete"
            :track-to-delete="trackToDelete"
            @confirm="removeAudioConfirm"
        ></AudioTrackDeleteModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="removeAudioConfirmInternal"
        ></AuthConfirmationModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import type { MediaAudioTrack } from "@/api/models";
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, nextTick, ref, shallowRef, useTemplateRef, watch } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { clone } from "@/utils/objects";
import { apiMediaRemoveAudioTrack, apiMediaRenameAudioTrack, apiMediaSetAudioTrack } from "@/api/api-media-edit";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useRequestId } from "@/composables/use-request-id";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
import { showSnackBarRight } from "@/control/snack-bar";
import { getCurrentMediaData, modifyCurrentMediaData } from "@/control/media";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

const AudioTrackDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AudioTrackDeleteModal.vue"),
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

// List of audio tracks
const audios = ref<MediaAudioTrack[]>(
    (getCurrentMediaData()?.audios || []).map((a) => {
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

    audios.value = (mediaData.audios || []).map((a) => {
        return {
            id: a.id,
            name: a.name,
            url: a.url,
        };
    });
});

// Audio file
const audioFile = shallowRef<File | null>(null);

// Audio file name
const audioFileName = ref("");

// Audio ID
const audioId = ref("");

// Audio name
const audioName = ref("");

// Hidden file input
const fileInputHidden = useTemplateRef("fileInputHidden");

/**
 * Opens the file selector
 */
const selectAudioFile = () => {
    const fileElem = fileInputHidden.value;
    if (fileElem) {
        fileElem.value = null;
        fileElem.click();
    }
};

/**
 * Sets the audio file to be uploaded
 * @param file The file
 */
const setFile = (file: File) => {
    audioFile.value = file;
    audioFileName.value = file.name;
    audioId.value = (file.name.split(".")[0] || "").toLowerCase();
    audioName.value = audioId.value.toUpperCase();
};

/**
 * Event handler for 'change' on the file input
 * @param e The event
 */
const audioFileChanged = (e: InputEvent) => {
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
 * Adds an audio file
 */
const addAudio = () => {
    if (!audioFile.value) {
        setError($t("Please, select an audio file first"));
        return;
    }

    const id = audioId.value;
    const name = audioName.value;

    let duped = false;
    for (const aud of audios.value) {
        if (aud.id === id) {
            duped = true;
            break;
        }
    }

    if (duped) {
        setError($t("There is already another audio track with the same identifier"));
        return;
    }

    if (busy.value) {
        return;
    }

    busy.value = true;
    uploading.value = true;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestIdAdd, apiMediaSetAudioTrack(mediaId, id, name, audioFile.value))
        .onSuccess((res) => {
            showSnackBarRight($t("Added audio track") + ": " + res.name);

            busy.value = false;
            audios.value.push(res);

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.audios = clone(audios.value);
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
                invalidAudio: () => {
                    setError($t("Invalid audio file"));
                },
                invalidId: () => {
                    setError($t("Invalid audio track identifier"));
                },
                invalidName: () => {
                    setError($t("Invalid audio track name"));
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
            busy.value = false;
        });
};

// Selected audio for rename
const audioRenameSelected = ref("");

// Id of the audio being renamed (new)
const audioRenameId = ref("");

// Name of the audio being renamed (new)
const audioRenameName = ref("");

// True if renaming an audio
const audioRenameBusy = ref(false);

// Request ID for renaming
const requestIdRename = useRequestId();

/**
 * Starts renaming an audio file
 * @param aud The audio file
 */
const startRename = (aud: MediaAudioTrack) => {
    audioRenameSelected.value = aud.id;
    audioRenameId.value = aud.id;
    audioRenameName.value = aud.name;

    nextTick(() => {
        const el = container.value?.querySelector(".edit-auto-focus") as HTMLInputElement;
        if (el) {
            el.focus();
            el.select();
        }
    });
};

/**
 * Cancels the rename
 */
const cancelRename = () => {
    audioRenameSelected.value = "";
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
 * Performs the request to rename an audio file
 */
const saveRename = () => {
    if (audioRenameBusy.value) {
        return;
    }

    const audioId = audioRenameSelected.value;

    const newId = audioRenameId.value;
    const newName = audioRenameName.value;

    if (!newId) {
        setError($t("Invalid audio track identifier"));
        return;
    }

    if (newId !== audioId) {
        for (const audio of audios.value) {
            if (audio.id === newId) {
                setError($t("Audio track identifier already in use"));
                return;
            }
        }
    }

    if (!newName) {
        setError($t("Invalid audio track name"));
        return;
    }

    audioRenameBusy.value = true;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(requestIdRename, apiMediaRenameAudioTrack(mediaId, audioId, newId, newName))
        .onSuccess(() => {
            showSnackBarRight($t("Renamed audio track") + ": " + newName + " (" + newId + ")");

            audioRenameBusy.value = false;
            audioRenameSelected.value = "";

            for (let i = 0; i < audios.value.length; i++) {
                if (audios.value[i].id === audioId) {
                    audios.value[i].id = newId;
                    audios.value[i].name = newName;
                    break;
                }
            }

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.audios = clone(audios.value);
            });

            emit("changed");
        })
        .onCancel(() => {
            audioRenameBusy.value = false;
        })
        .onRequestError((err, handleErr) => {
            audioRenameBusy.value = false;
            handleErr(err, {
                unauthorized,
                invalidId: () => {
                    setError($t("Invalid audio track identifier"));
                },
                invalidName: () => {
                    setError($t("Invalid audio track name"));
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
            audioRenameBusy.value = false;
        });
};

// True to display the delete confirmation
const displayAudioTrackDelete = ref(false);

// The reference to the audio tract to delete
const trackToDelete = ref<MediaAudioTrack | null>(null);

/**
 * Opens confirmation modal to delete
 * @param aud The audio file
 */
const removeAudio = (aud: MediaAudioTrack) => {
    trackToDelete.value = aud;
    displayAudioTrackDelete.value = true;
};

/**
 * User confirmed to delete an audio file
 */
const removeAudioConfirm = () => {
    removeAudioConfirmInternal({});
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
 * Performs the request to delete an audio track
 * @param confirmation The auth confirmation
 */
const removeAudioConfirmInternal = (confirmation: ProvidedAuthConfirmation) => {
    const aud = trackToDelete.value;

    if (busyDeleting.value || !aud) {
        return;
    }

    busyDeleting.value = true;
    busyDeletingId.value = aud.id;

    const mediaId = AppStatus.CurrentMedia;
    const id = aud.id;

    makeNamedApiRequest(requestIdDelete, apiMediaRemoveAudioTrack(mediaId, id, confirmation))
        .onSuccess(() => {
            showSnackBarRight($t("Removed audio track") + ": " + aud.name);

            busyDeleting.value = false;

            for (let i = 0; i < audios.value.length; i++) {
                if (audios.value[i].id === id) {
                    audios.value.splice(i, 1);
                    break;
                }
            }

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.audios = clone(audios.value);
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
 * Downloads the audio file
 * @param aud The audio file
 */
const downloadAudio = (aud: MediaAudioTrack) => {
    const link = document.createElement("a");
    link.target = "_blank";
    link.rel = "noopener noreferrer";

    const url = getAssetURL(aud.url);
    const titlePart = aud.id ? "&filename=" + encodeURIComponent(aud.id) : "";

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
