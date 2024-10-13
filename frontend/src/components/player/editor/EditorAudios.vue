<template>
    <div class="player-editor-sub-content" @drop="onDrop">
        <!--- Audio tracks -->

        <div class="form-group">
            <label>{{ $t("Extra audio tracks for the video.") }} {{ $t("You can use this to add multiple audio languages.") }}</label>
        </div>

        <div v-if="type === 2" class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("ID") }}</th>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-right td-shrink"></th>
                        <th class="text-right td-shrink" v-if="canWrite"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="audios.length === 0">
                        <td colspan="3">
                            {{ $t("There are no audio tracks yet for this media.") }}
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite"></td>
                    </tr>
                    <tr v-for="aud in audios" :key="aud.id">
                        <td class="bold" v-if="audioRenameSelected !== aud.id">{{ aud.id }}</td>
                        <td v-else>
                            <input
                                type="text"
                                maxlength="255"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                class="form-control form-control-full-width"
                                v-model="audioRenameId"
                                @keydown="renameInputKeyEventHandler"
                            />
                        </td>

                        <td class="bold" v-if="audioRenameSelected !== aud.id">{{ aud.name }}</td>
                        <td v-else>
                            <input
                                type="text"
                                maxlength="255"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                class="form-control form-control-full-width edit-auto-focus"
                                v-model="audioRenameName"
                                @keydown="renameInputKeyEventHandler"
                            />
                        </td>

                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" @click="downloadAudio(aud)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink one-line" v-if="audioRenameSelected === aud.id && canWrite">
                            <button
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                                @click="saveRename"
                            >
                                <i class="fas fa-check"></i> {{ $t("Save") }}
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
                        <td class="text-right td-shrink one-line" v-else-if="canWrite">
                            <button
                                v-if="!audioRenameSelected"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                @click="startRename(aud)"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button v-if="busyDeleting && busyDeletingId === aud.id" type="button" class="btn btn-danger btn-xs" disabled>
                                <i class="fa fa-spinner fa-spin"></i> {{ $t("Deleting") }}...
                            </button>
                            <button
                                v-else
                                type="button"
                                class="btn btn-danger btn-xs"
                                @click="removeAudio(aud)"
                                :disabled="busy || busyDeleting || audioRenameBusy"
                            >
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group" v-if="canWrite && type === 2">
            <label>{{ $t("You can upload extra audio tracks for the video (.mp3)") }}:</label>
            <input type="file" class="file-hidden audio-file-hidden" @change="audioFileChanged" name="mp3-upload" accept=".mp3" />
            <button v-if="!audioFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectAudioFile">
                <i class="fas fa-upload"></i> {{ $t("Select audio file") }}
            </button>

            <button v-if="audioFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectAudioFile">
                <i class="fas fa-upload"></i> {{ $t("Audio file") }}: {{ audioFileName }}
            </button>
        </div>
        <div class="form-group" v-if="canWrite && type === 2">
            <label>{{ $t("Audio track identifier") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="audioId" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && type === 2">
            <label>{{ $t("Audio track name") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="audioName" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && type === 2">
            <button v-if="!busy" type="button" class="btn btn-primary" :disabled="!audioId || !audioName || !audioFile" @click="addAudio">
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

        <AudioTrackDeleteModal
            v-model:display="displayAudioTrackDelete"
            :trackToDelete="trackToDelete"
            @confirm="removeAudioConfirm"
        ></AudioTrackDeleteModal>
    </div>
</template>

<script lang="ts">
import { MediaAudioTrack } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";

import AudioTrackDeleteModal from "@/components/modals/AudioTrackDeleteModal.vue";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaRemoveAudioTrack, apiMediaRenameAudioTrack, apiMediaSetAudioTrack } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        AudioTrackDeleteModal,
    },
    name: "EditorAudios",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            audios: [] as MediaAudioTrack[],
            audioFile: null,
            audioFileName: "",
            audioId: "en",
            audioName: "English",

            busy: false,
            busyDeleting: false,
            busyDeletingId: "",

            uploading: false,
            uploadProgress: 0,

            audioRenameSelected: "",
            audioRenameId: "",
            audioRenameName: "",
            audioRenameBusy: false,

            canWrite: AuthController.CanWrite,

            displayAudioTrackDelete: false,
            trackToDelete: null as MediaAudioTrack,
        };
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.audios = (MediaController.MediaData.audios || []).map((a) => {
                return {
                    id: a.id,
                    name: a.name,
                    url: a.url,
                };
            });
        },

        // Audios

        selectAudioFile: function () {
            this.$el.querySelector(".audio-file-hidden").click();
        },

        audioFileChanged: function (e: InputEvent) {
            const data = (e.target as HTMLInputElement).files;
            if (data && data.length > 0) {
                const file = data[0];
                this.audioFile = file;
                this.audioFileName = file.name;
            }
        },

        onDrop: function (e: DragEvent) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.audioFile = file;
                this.audioFileName = file.name;
            }
        },

        addAudio: function () {
            if (!this.audioFile) {
                PagesController.ShowSnackBarRight(this.$t("Please, select an audio file first"));
                return;
            }

            const id = this.audioId;
            const name = this.audioName;

            let duped = false;
            for (const aud of this.audios) {
                if (aud.id === id) {
                    duped = true;
                    break;
                }
            }

            if (duped) {
                PagesController.ShowSnackBarRight(this.$t("There is already another audio track with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;
            this.uploading = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaSetAudioTrack(mediaId, id, name, this.audioFile))
                .onSuccess((res) => {
                    PagesController.ShowSnackBarRight(this.$t("Added audio track") + ": " + res.name);
                    this.busy = false;
                    this.audios.push(res);
                    if (MediaController.MediaData) {
                        MediaController.MediaData.audios = clone(this.audios);
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onUploadProgress((loaded, total) => {
                    this.uploadProgress = loaded / Math.max(1, total);
                    this.uploading = loaded < total;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidAudio: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio file"));
                        },
                        invalidId: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio track identifier"));
                        },
                        invalidName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio track name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        removeAudio: function (aud: MediaAudioTrack) {
            this.trackToDelete = aud;
            this.displayAudioTrackDelete = true;
        },

        removeAudioConfirm: function () {
            const aud = this.trackToDelete;

            if (this.busyDeleting || !aud) {
                return;
            }

            this.busyDeleting = true;
            this.busyDeletingId = aud.id;

            const mediaId = AppStatus.CurrentMedia;
            const id = aud.id;

            makeNamedApiRequest(this.requestId, apiMediaRemoveAudioTrack(mediaId, id))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Removed audio track") + ": " + aud.name);
                    this.busyDeleting = false;
                    for (let i = 0; i < this.audios.length; i++) {
                        if (this.audios[i].id === id) {
                            this.audios.splice(i, 1);
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.audios = clone(this.audios);
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busyDeleting = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busyDeleting = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.busyDeleting = false;
                });
        },

        downloadAudio: function (aud: MediaAudioTrack) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = getAssetURL(aud.url);
            link.click();
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        renderProgress: function (p: number): string {
            return Math.max(0, Math.min(100, Math.floor(p * 100))) + "%";
        },

        startRename: function (aud: MediaAudioTrack) {
            this.audioRenameSelected = aud.id;
            this.audioRenameId = aud.id;
            this.audioRenameName = aud.name;

            nextTick(() => {
                const el = this.$el.querySelector(".edit-auto-focus");
                if (el) {
                    el.focus();
                    el.select();
                }
            });
        },

        cancelRename: function () {
            this.audioRenameSelected = "";
        },

        renameInputKeyEventHandler: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();
                this.saveRename();
            }
        },

        saveRename: function () {
            if (this.audioRenameBusy) {
                return;
            }

            const audioId = this.audioRenameSelected;

            const newId = this.audioRenameId;
            const newName = this.audioRenameName;

            if (!newId) {
                PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio track identifier"));
                return;
            }

            if (newId !== audioId) {
                for (const audio of this.audios) {
                    if (audio.id === newId) {
                        PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Audio track identifier already in use"));
                        return;
                    }
                }
            }

            if (!newName) {
                PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio track name"));
                return;
            }

            this.audioRenameBusy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaRenameAudioTrack(mediaId, audioId, newId, newName))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Renamed audio track") + ": " + newName + " (" + newId + ")");
                    this.audioRenameBusy = false;
                    this.audioRenameSelected = "";
                    for (let i = 0; i < this.audios.length; i++) {
                        if (this.audios[i].id === audioId) {
                            this.audios[i].id = newId;
                            this.audios[i].name = newName;
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.audios = clone(this.audios);
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.audioRenameBusy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.audioRenameBusy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidId: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio track identifier"));
                        },
                        invalidName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid audio track name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.audioRenameBusy = false;
                });
        },
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);
    },
});
</script>
