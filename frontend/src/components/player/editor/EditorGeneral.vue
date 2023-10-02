<template>
    <div class="player-editor-sub-content content-row">
        <div class="row-general-left">
            <!--- Title -->

            <form @submit="changeTitle">
                <div class="form-group">
                    <label>{{ $t("Title") }}:</label>
                    <input
                        type="text"
                        autocomplete="off"
                        :readonly="!canWrite"
                        maxlength="255"
                        :disabled="busyTitle"
                        v-model="title"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-group" v-if="canWrite">
                    <button type="submit" class="btn btn-primary" :disabled="busyTitle || !title || originalTitle === title">
                        <i class="fas fa-pencil-alt"></i> {{ $t("Change title") }}
                    </button>
                </div>
            </form>

            <!--- Description -->

            <div class="form-group">
                <label>{{ $t("Description") }}:</label>
                <textarea
                    v-model="desc"
                    :readonly="!canWrite"
                    maxlength="1024"
                    class="form-control form-control-full-width form-textarea"
                    rows="3"
                    :disabled="busyDescription"
                ></textarea>
            </div>
            <div class="form-group" v-if="canWrite">
                <button
                    type="button"
                    class="btn btn-primary"
                    :disabled="busyDescription || originalDesc === desc"
                    @click="changeDescription"
                >
                    <i class="fas fa-pencil-alt"></i> {{ $t("Change description") }}
                </button>
            </div>

            <!--- Extra config -->

            <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
                <label>{{ $t("Extra media configuration") }}:</label>
            </div>
            <div class="table-responsive" v-if="canWrite && (type === 2 || type === 3)">
                <table class="table no-border">
                    <tr v-if="type === 2 || type === 3">
                        <td class="text-right td-shrink">
                            <toggle-switch v-model:val="startBeginning" :disabled="busyExtra"></toggle-switch>
                        </td>
                        <td class="">
                            {{ $t("Reset time to the beginning every time the media reloads?") }}
                        </td>
                    </tr>
                </table>
            </div>
            <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
                <button
                    type="button"
                    class="btn btn-primary"
                    :disabled="busyExtra || originalStartBeginning === startBeginning"
                    @click="changeExtraParams"
                >
                    <i class="fas fa-pencil-alt"></i> {{ $t("Change extra configuration") }}
                </button>
            </div>
        </div>

        <div class="row-general-right">
            <!--- Thumbnail -->

            <div class="form-group">
                <label>{{ $t("Thumbnail") }}:</label>
            </div>
            <div class="form-group" @drop="onDrop">
                <label v-if="!thumbnail">{{ $t("No thumbnail set for this media") }}</label>
                <img v-if="thumbnail" :src="getThumbnail(thumbnail)" :alt="originalTitle" class="form-group-thumbnail" loading="lazy" />
            </div>
            <div class="form-group" v-if="canWrite">
                <input type="file" class="file-hidden" @change="inputFileChanged" name="thumbnail-upload" />
                <button v-if="!busyThumbnail" type="button" class="btn btn-primary" @click="uploadThumbnail">
                    <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
                </button>
                <button v-if="busyThumbnail" type="button" class="btn btn-primary" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading thumbnail") }}...
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
import { GetAssetURL, Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import { EditMediaAPI } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        ToggleSwitch,
    },
    name: "EditorGeneral",
    emits: ["changed"],
    data: function () {
        return {
            type: 0,

            title: "",
            originalTitle: "",

            desc: "",
            originalDesc: "",

            thumbnail: "",

            busyTitle: false,
            busyDescription: false,
            busyThumbnail: false,
            busyExtra: false,

            canWrite: AuthController.CanWrite,

            originalStartBeginning: false,
            startBeginning: false,
        };
    },

    methods: {
        autoFocus: function () {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.originalTitle = MediaController.MediaData.title;
            this.title = this.originalTitle;

            this.originalDesc = MediaController.MediaData.description;
            this.desc = this.originalDesc;

            this.originalStartBeginning = MediaController.MediaData.force_start_beginning;
            this.startBeginning = this.originalStartBeginning;

            this.thumbnail = MediaController.MediaData.thumbnail;
        },

        getThumbnail(thumb: string) {
            return GetAssetURL(thumb);
        },

        uploadThumbnail: function () {
            this.$el.querySelector(".file-hidden").click();
        },

        inputFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.changeThumbnail(file);
            }
        },

        onDrop: function (e) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.changeThumbnail(file);
            }
        },

        changeThumbnail: function (file) {
            if (this.busyThumbnail) {
                return;
            }

            this.busyThumbnail = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy-thumbnail", EditMediaAPI.ChangeMediaThumbnail(mediaId, file))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Successfully changed thumbnail"));
                    this.busyThumbnail = false;
                    this.thumbnail = res.url;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.thumbnail = res.url;
                    }
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit("media-meta-change");
                })
                .onCancel(() => {
                    this.busyThumbnail = false;
                })
                .onRequestError((err) => {
                    this.busyThumbnail = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid thumbnail provided"));
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
                    this.busyThumbnail = false;
                });
        },

        changeTitle: function (e) {
            if (e) {
                e.preventDefault();
            }

            if (this.busyTitle) {
                return;
            }

            this.busyTitle = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy-title", EditMediaAPI.ChangeMediaTitle(mediaId, this.title))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed title"));
                    this.busyTitle = false;
                    this.originalTitle = this.title;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.title = this.title;
                    }
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit("media-meta-change");
                })
                .onCancel(() => {
                    this.busyTitle = false;
                })
                .onRequestError((err) => {
                    this.busyTitle = false;
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
                    this.busyTitle = false;
                });
        },

        changeDescription: function () {
            if (this.busyDescription) {
                return;
            }

            this.busyDescription = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy-description", EditMediaAPI.ChangeMediaDescription(mediaId, this.desc))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed description"));
                    this.busyDescription = false;
                    this.originalDesc = this.desc;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.description = this.desc;
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busyDescription = false;
                })
                .onRequestError((err) => {
                    this.busyDescription = false;
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
                    this.busyDescription = false;
                });
        },

        changeExtraParams: function () {
            if (this.busyExtra) {
                return;
            }

            this.busyExtra = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy-extra", EditMediaAPI.ChangeExtraParams(mediaId, this.startBeginning))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed media extra params"));
                    this.busyExtra = false;
                    this.originalStartBeginning = this.startBeginning;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.force_start_beginning = this.startBeginning;
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busyExtra = false;
                })
                .onRequestError((err) => {
                    this.busyExtra = false;
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
                    this.busyExtra = false;
                });
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

        this.autoFocus();
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        Request.Abort("media-editor-busy-title");
        Request.Abort("media-editor-busy-description");
        Request.Abort("media-editor-busy-thumbnail");
        Request.Abort("media-editor-busy-extra");
    },
});
</script>
