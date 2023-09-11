<template>
    <div class="player-editor-sub-content">
        <!--- Audio tracks -->

        <div class="form-group" v-if="type === 2">
            <label>{{ $t("Extra audio tracks") }}:</label>
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
                        <td class="bold">{{ aud.id }}</td>
                        <td class="bold">{{ aud.name }}</td>
                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" :disabled="busy" @click="downloadAudio(aud)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite">
                            <button type="button" class="btn btn-danger btn-xs" @click="removeAudio(aud)">
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
            <button type="button" class="btn btn-primary" :disabled="busy || !audioId || !audioName || !audioFile" @click="addAudio">
                <i class="fas fa-plus"></i> {{ $t("Add audio track file") }}
            </button>
        </div>

        <AudioTrackDeleteModal ref="audioTrackDeleteModal" v-model:display="displayAudioTrackDelete"></AudioTrackDeleteModal>
    </div>
</template>

<script lang="ts">
import { MediaAudioTrack } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
import { GetAssetURL, Request } from "@/utils/request";
import { defineComponent } from "vue";

import AudioTrackDeleteModal from "@/components/modals/AudioTrackDeleteModal.vue";
import { EditMediaAPI } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        AudioTrackDeleteModal,
    },
    name: "EditorAudios",
    emits: ["changed"],
    data: function () {
        return {
            type: 0,

            audios: [],
            audioFile: null,
            audioFileName: "",
            audioId: "en",
            audioName: "English",

            busy: false,

            canWrite: AuthController.CanWrite,

            displayAudioTrackDelete: false,
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

        audioFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.audioFile = file;
                this.audioFileName = file.name;
            }
        },

        addAudio: function () {
            if (!this.audioFile) {
                AppEvents.Emit("snack", this.$t("Please, select an audio file first"));
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
                AppEvents.Emit("snack", this.$t("There is already another audio track with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy-audios", EditMediaAPI.SetAudioTrack(mediaId, id, name, this.audioFile))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added audio track") + ": " + res.name);
                    this.busy = false;
                    this.audios.push(res);
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "INVALID_AUDIO", () => {
                            AppEvents.Emit("snack", this.$t("Invalid audio file"));
                        })
                        .add(400, "INVALID_ID", () => {
                            AppEvents.Emit("snack", this.$t("Invalid audio track identifier"));
                        })
                        .add(400, "INVALID_NAME", () => {
                            AppEvents.Emit("snack", this.$t("Invalid audio track name"));
                        })
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        removeAudio: function (aud: MediaAudioTrack) {
            this.$refs.audioTrackDeleteModal.show({
                name: aud.name,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;
                    const id = aud.id;

                    Request.Pending("media-editor-busy-audios", EditMediaAPI.RemoveAudioTrack(mediaId, id))
                        .onSuccess(() => {
                            AppEvents.Emit("snack", this.$t("Removed audio track") + ": " + aud.name);
                            this.busy = false;
                            for (let i = 0; i < this.audios.length; i++) {
                                if (this.audios[i].id === id) {
                                    this.audios.splice(i, 1);
                                    break;
                                }
                            }
                            this.$emit("changed");
                        })
                        .onCancel(() => {
                            this.busy = false;
                        })
                        .onRequestError((err) => {
                            this.busy = false;
                            Request.ErrorHandler()
                                .add(400, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Bad request"));
                                })
                                .add(401, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                    AppEvents.Emit("unauthorized");
                                })
                                .add(403, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                })
                                .add(404, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Not found"));
                                })
                                .add(500, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Internal server error"));
                                })
                                .add("*", "*", () => {
                                    AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                                })
                                .handle(err);
                        })
                        .onUnexpectedError((err) => {
                            AppEvents.Emit("snack", err.message);
                            console.error(err);
                            this.busy = false;
                        });
                },
            });
        },

        downloadAudio: function (aud: MediaAudioTrack) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = GetAssetURL(aud.url);
            link.click();
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },

    mounted: function () {
        this._handles = Object.create(null);
        this.updateMediaData();

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this._handles.mediaUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        Request.Abort("media-editor-busy-audios");
    },
});
</script>
