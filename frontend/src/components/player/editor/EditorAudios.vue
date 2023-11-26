<template>
    <div class="player-editor-sub-content">
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
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";

import AudioTrackDeleteModal from "@/components/modals/AudioTrackDeleteModal.vue";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaRemoveAudioTrack, apiMediaSetAudioTrack } from "@/api/api-media-edit";

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
                PagesController.ShowSnackBar(this.$t("Please, select an audio file first"));
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
                PagesController.ShowSnackBar(this.$t("There is already another audio track with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaSetAudioTrack(mediaId, id, name, this.audioFile))
                .onSuccess((res) => {
                    PagesController.ShowSnackBar(this.$t("Added audio track") + ": " + res.name);
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
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidAudio: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid audio file"));
                        },
                        invalidId: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid audio track identifier"));
                        },
                        invalidName: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid audio track name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBar(err.message);
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

                    makeNamedApiRequest(this.requestId, apiMediaRemoveAudioTrack(mediaId, id))
                        .onSuccess(() => {
                            PagesController.ShowSnackBar(this.$t("Removed audio track") + ": " + aud.name);
                            this.busy = false;
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
                            this.busy = false;
                        })
                        .onRequestError((err, handleErr) => {
                            this.busy = false;
                            handleErr(err, {
                                unauthorized: () => {
                                    PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                                    AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                                },
                                badRequest: () => {
                                    PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                                },
                                accessDenied: () => {
                                    PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                                },
                                notFound: () => {
                                    PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                                },
                                serverError: () => {
                                    PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                                },
                                networkError: () => {
                                    PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                                },
                            });
                        })
                        .onUnexpectedError((err) => {
                            PagesController.ShowSnackBar(err.message);
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
            link.href = getAssetURL(aud.url);
            link.click();
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
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
