<template>
    <div class="player-editor-sub-content">
        <!--- Subtitles -->

        <div class="form-group" v-if="type === 2 || type === 3">
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
                            <button type="button" class="btn btn-primary btn-xs" :disabled="busy" @click="downloadSubtitles(sub)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite">
                            <button type="button" class="btn btn-danger btn-xs" @click="removeSubtitles(sub)">
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
            <button type="button" class="btn btn-primary" :disabled="busy || !srtId || !srtName || !srtFile" @click="addSubtitles">
                <i class="fas fa-plus"></i> {{ $t("Add subtitles file") }}
            </button>
        </div>

        <SubtitlesDeleteModal ref="subtitlesDeleteModal" v-model:display="displaySubtitlesDelete"></SubtitlesDeleteModal>
    </div>
</template>

<script lang="ts">
import { MediaSubtitle } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
import { GetAssetURL, Request } from "@/utils/request";
import { defineComponent } from "vue";
import SubtitlesDeleteModal from "@/components/modals/SubtitlesDeleteModal.vue";
import { EditMediaAPI } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        SubtitlesDeleteModal,
    },
    name: "EditorSubtitles",
    emits: ["changed"],
    data: function () {
        return {
            page: "general",

            type: 0,

            subtitles: [],
            srtFile: null,
            srtFileName: "",
            srtId: "en",
            srtName: "English",

            busy: false,

            canWrite: AuthController.CanWrite,

            displaySubtitlesDelete: false,
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
                AppEvents.Emit("snack", this.$t("Please, select a SubRip file first"));
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
                AppEvents.Emit("snack", this.$t("There is already another subtitles file with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy-subtitles", EditMediaAPI.SetSubtitles(mediaId, id, name, this.srtFile))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added subtitles") + ": " + res.name);
                    this.busy = false;
                    this.subtitles.push(res);
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "INVALID_SRT", () => {
                            AppEvents.Emit("snack", this.$t("Invalid SubRip file"));
                        })
                        .add(400, "INVALID_ID", () => {
                            AppEvents.Emit("snack", this.$t("Invalid subtitles identifier"));
                        })
                        .add(400, "INVALID_NAME", () => {
                            AppEvents.Emit("snack", this.$t("Invalid subtitles name"));
                        })
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(413, "*", () => {
                            AppEvents.Emit("snack", this.$t("Subtitles file too big (max is $MAX)").replace("$MAX", "10MB"));
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

        removeSubtitles: function (sub) {
            this.$refs.subtitlesDeleteModal.show({
                name: sub.name,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;
                    const id = sub.id;

                    Request.Pending("media-editor-busy-subtitles", EditMediaAPI.RemoveSubtitles(mediaId, id))
                        .onSuccess(() => {
                            AppEvents.Emit("snack", this.$t("Removed subtitles") + ": " + sub.name);
                            this.busy = false;
                            for (let i = 0; i < this.subtitles.length; i++) {
                                if (this.subtitles[i].id === id) {
                                    this.subtitles.splice(i, 1);
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

        downloadSubtitles: function (sub: MediaSubtitle) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = GetAssetURL(sub.url);
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

        Request.Abort("media-editor-busy-subtitles");
    },
});
</script>