<template>
    <div class="player-editor-sub-content">
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
                        <th class="text-right td-shrink" v-if="canWrite"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="subtitles.length === 0">
                        <td colspan="3">
                            {{ $t("There are no subtitles yet for this media.") }}
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite"></td>
                    </tr>
                    <tr v-for="sub in subtitles" :key="sub.id">
                        <td class="bold">{{ sub.id }}</td>
                        <td class="bold">{{ sub.name }}</td>
                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" @click="downloadSubtitles(sub)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite">
                            <button v-if="busyDeleting && busyDeletingId === sub.id" type="button" class="btn btn-danger btn-xs" disabled>
                                <i class="fa fa-spinner fa-spin"></i> {{ $t("Deleting") }}...
                            </button>
                            <button
                                v-else
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busyDeleting"
                                @click="removeSubtitles(sub)"
                            >
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("You can upload subtitles in SubRip format (.srt)") }}:</label>
            <input type="file" class="file-hidden srt-file-hidden" @change="srtFileChanged" name="srt-upload" accept=".srt" />
            <button v-if="!srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                <i class="fas fa-upload"></i> {{ $t("Select SRT file") }}
            </button>

            <button v-if="srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                <i class="fas fa-upload"></i> {{ $t("SRT file") }}: {{ srtFileName }}
            </button>
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("Subtitles identifier") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="srtId" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("Subtitles name") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="srtName" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
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
    </div>
</template>

<script lang="ts">
import { MediaSubtitle } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";
import SubtitlesDeleteModal from "@/components/modals/SubtitlesDeleteModal.vue";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaRemoveSubtitles, apiMediaSetSubtitles } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        SubtitlesDeleteModal,
    },
    name: "EditorSubtitles",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            page: "general",

            type: 0,

            subtitles: [] as MediaSubtitle[],
            srtFile: null,
            srtFileName: "",
            srtId: "en",
            srtName: "English",

            busy: false,
            busyDeleting: false,
            busyDeletingId: "",

            uploading: false,
            uploadProgress: 0,

            canWrite: AuthController.CanWrite,

            displaySubtitlesDelete: false,
            subtitleToDelete: null as MediaSubtitle,
        };
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

        srtFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.srtFile = file;
                this.srtFileName = file.name;
            }
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

            makeNamedApiRequest(this.requestId, apiMediaSetSubtitles(mediaId, id, name, this.srtFile))
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
            const sub = this.subtitleToDelete;

            if (this.busyDeleting || !sub) {
                return;
            }

            this.busyDeleting = true;
            this.busyDeletingId = sub.id;

            const mediaId = AppStatus.CurrentMedia;
            const id = sub.id;

            makeNamedApiRequest(this.requestId, apiMediaRemoveSubtitles(mediaId, id))
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
            link.href = getAssetURL(sub.url);
            link.click();
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        renderProgress: function (p: number): string {
            return Math.max(0, Math.min(100, Math.floor(p * 100))) + "%";
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
