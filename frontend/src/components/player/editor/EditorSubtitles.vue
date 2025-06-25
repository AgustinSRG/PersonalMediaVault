<template>
    <div class="player-editor-sub-content" @drop="onDrop">
        <!--- Subtitles -->

        <div class="form-group">
            <label>{{ $t("Subtitles") }}:</label>
        </div>

        <div v-if="type === 2 || type === 3" class="table-responsive">
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

        <div v-if="canWrite && (type === 2 || type === 3)" class="form-group">
            <label>{{ $t("You can upload subtitles in SubRip format (.srt)") }}:</label>
            <input type="file" class="file-hidden srt-file-hidden" name="srt-upload" accept=".srt" @change="srtFileChanged" />
            <button v-if="!srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                <i class="fas fa-upload"></i> {{ $t("Select SRT file") }}
            </button>

            <button v-if="srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                <i class="fas fa-upload"></i> {{ $t("SRT file") }}: {{ srtFileName }}
            </button>
        </div>
        <div v-if="canWrite && (type === 2 || type === 3)" class="form-group">
            <label>{{ $t("Subtitles identifier") }}:</label>
            <input v-model="srtId" type="text" autocomplete="off" maxlength="255" :disabled="busy" class="form-control" />
        </div>
        <div v-if="canWrite && (type === 2 || type === 3)" class="form-group">
            <label>{{ $t("Subtitles name") }}:</label>
            <input v-model="srtName" type="text" autocomplete="off" maxlength="255" :disabled="busy" class="form-control" />
        </div>
        <div v-if="canWrite && (type === 2 || type === 3)" class="form-group">
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
    </div>
</template>

<script lang="ts">
import type { MediaSubtitle } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import SubtitlesDeleteModal from "@/components/modals/SubtitlesDeleteModal.vue";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaRemoveSubtitles, apiMediaRenameSubtitles, apiMediaSetSubtitles } from "@/api/api-media-edit";
import AuthConfirmationModal from "@/components/modals/AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "EditorSubtitles",
    components: {
        LoadingIcon,
        SubtitlesDeleteModal,
        AuthConfirmationModal,
    },
    emits: ["changed"],
    setup() {
        return {
            requestIdAdd: getUniqueStringId(),
            requestIdRename: getUniqueStringId(),
            requestIdDelete: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            page: "general",

            type: 0,

            subtitles: [] as MediaSubtitle[],
            srtFile: null as File | null,
            srtFileName: "",
            srtId: "",
            srtName: "",

            busy: false,
            busyDeleting: false,
            busyDeletingId: "",

            uploading: false,
            uploadProgress: 0,

            subtitleRenameSelected: "",
            subtitleRenameId: "",
            subtitleRenameName: "",
            subtitleRenameBusy: false,

            canWrite: AuthController.CanWrite,

            displaySubtitlesDelete: false,
            subtitleToDelete: null as MediaSubtitle,

            displayAuthConfirmation: false,
            authConfirmationCooldown: 0,
            authConfirmationTfa: false,
            authConfirmationError: "",
        };
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestIdAdd);
        abortNamedApiRequest(this.requestIdRename);
        abortNamedApiRequest(this.requestIdDelete);
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.subtitles = (MediaController.MediaData.subtitles || []).map((a) => {
                return {
                    id: a.id,
                    name: a.name,
                    url: a.url,
                };
            });
        },

        // Subtitles

        selectSRTFile: function () {
            this.$el.querySelector(".srt-file-hidden").click();
        },

        srtFileChanged: function (e: InputEvent) {
            const data = (e.target as HTMLInputElement).files;
            if (data && data.length > 0) {
                this.setFile(data[0]);
            }
        },

        onDrop: function (e: DragEvent) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                this.setFile(data[0]);
            }
        },

        setFile: function (file: File) {
            this.srtFile = file;
            this.srtFileName = file.name;
            this.srtId = (file.name.split(".")[0] || "").toLowerCase();
            this.srtName = this.srtId.toUpperCase();
        },

        addSubtitles: function () {
            if (!this.srtFile) {
                PagesController.ShowSnackBarRight(this.$t("Please, select a SubRip file first"));
                return;
            }

            const id = this.srtId;
            const name = this.srtName;

            let duped = false;
            for (const sub of this.subtitles) {
                if (sub.id === id) {
                    duped = true;
                    break;
                }
            }

            if (duped) {
                PagesController.ShowSnackBarRight(this.$t("There is already another subtitles file with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;
            this.uploading = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestIdAdd, apiMediaSetSubtitles(mediaId, id, name, this.srtFile))
                .onSuccess((res) => {
                    PagesController.ShowSnackBarRight(this.$t("Added subtitles") + ": " + res.name);
                    this.busy = false;
                    this.subtitles.push(res);
                    if (MediaController.MediaData) {
                        MediaController.MediaData.subtitles = clone(this.subtitles);
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
                        invalidSRT: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid SubRip file"));
                        },
                        invalidId: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid subtitles identifier"));
                        },
                        invalidName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid subtitles name"));
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
                        fileTooLarge: () => {
                            PagesController.ShowSnackBarRight(
                                this.$t("Error") + ": " + this.$t("Subtitles file too big (max is $MAX)").replace("$MAX", "10MB"),
                            );
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

        removeSubtitles: function (sub: MediaSubtitle) {
            this.subtitleToDelete = sub;
            this.displaySubtitlesDelete = true;
        },

        removeSubtitlesConfirm: function () {
            this.removeSubtitlesConfirmInternal({});
        },

        removeSubtitlesConfirmInternal: function (confirmation: ProvidedAuthConfirmation) {
            const sub = this.subtitleToDelete;

            if (this.busyDeleting || !sub) {
                return;
            }

            this.busyDeleting = true;
            this.busyDeletingId = sub.id;

            const mediaId = AppStatus.CurrentMedia;
            const id = sub.id;

            makeNamedApiRequest(this.requestIdDelete, apiMediaRemoveSubtitles(mediaId, id, confirmation))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Removed subtitles") + ": " + sub.name);
                    this.busyDeleting = false;
                    for (let i = 0; i < this.subtitles.length; i++) {
                        if (this.subtitles[i].id === id) {
                            this.subtitles.splice(i, 1);
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.subtitles = clone(this.subtitles);
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
                        requiredAuthConfirmationPassword: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = "";
                            this.authConfirmationTfa = false;
                        },
                        invalidPassword: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("Invalid password");
                            this.authConfirmationTfa = false;
                            this.authConfirmationCooldown = Date.now() + 5000;
                        },
                        requiredAuthConfirmationTfa: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = "";
                            this.authConfirmationTfa = true;
                        },
                        invalidTfaCode: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("Invalid one-time code");
                            this.authConfirmationTfa = true;
                            this.authConfirmationCooldown = Date.now() + 5000;
                        },
                        cooldown: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("You must wait 5 seconds to try again");
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

        downloadSubtitles: function (sub: MediaSubtitle) {
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
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        renderProgress: function (p: number): string {
            return Math.max(0, Math.min(100, Math.floor(p * 100))) + "%";
        },

        startRename: function (sub: MediaSubtitle) {
            this.subtitleRenameSelected = sub.id;
            this.subtitleRenameId = sub.id;
            this.subtitleRenameName = sub.name;

            nextTick(() => {
                const el = this.$el.querySelector(".edit-auto-focus");
                if (el) {
                    el.focus();
                    el.select();
                }
            });
        },

        cancelRename: function () {
            this.subtitleRenameSelected = "";
        },

        renameInputKeyEventHandler: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();
                this.saveRename();
            }
        },

        saveRename: function () {
            if (this.subtitleRenameBusy) {
                return;
            }

            const subtitleId = this.subtitleRenameSelected;

            const newId = this.subtitleRenameId;
            const newName = this.subtitleRenameName;

            if (!newId) {
                PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid subtitles identifier"));
                return;
            }

            if (newId !== subtitleId) {
                for (const subtitle of this.subtitles) {
                    if (subtitle.id === newId) {
                        PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Subtitles identifier already in use"));
                        return;
                    }
                }
            }

            if (!newName) {
                PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid subtitles name"));
                return;
            }

            this.subtitleRenameBusy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestIdRename, apiMediaRenameSubtitles(mediaId, subtitleId, newId, newName))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Renamed subtitles") + ": " + newName + " (" + newId + ")");
                    this.subtitleRenameBusy = false;
                    this.subtitleRenameSelected = "";
                    for (let i = 0; i < this.subtitles.length; i++) {
                        if (this.subtitles[i].id === subtitleId) {
                            this.subtitles[i].id = newId;
                            this.subtitles[i].name = newName;
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.subtitles = clone(this.subtitles);
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.subtitleRenameBusy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.subtitleRenameBusy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidId: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid subtitles identifier"));
                        },
                        invalidName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid subtitles name"));
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
                    this.subtitleRenameBusy = false;
                });
        },
    },
});
</script>
