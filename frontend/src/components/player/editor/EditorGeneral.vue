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
                <img v-if="thumbnail" :src="getThumbnail(thumbnail)" :alt="$t('Thumbnail')" class="form-group-thumbnail" loading="lazy" />
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
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { Request } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import { EditMediaAPI } from "@/api/api-media-edit";
import { EVENT_NAME_MEDIA_METADATA_CHANGE } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";

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
            return getAssetURL(thumb);
        },

        uploadThumbnail: function () {
            const fileElem = this.$el.querySelector(".file-hidden");
            if (fileElem) {
                fileElem.value = null;
                fileElem.click();
            }
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

            Request.Pending(this._handles.requestIdThumbnail, EditMediaAPI.ChangeMediaThumbnail(mediaId, file))
                .onSuccess((res) => {
                    AppEvents.ShowSnackBar(this.$t("Successfully changed thumbnail"));
                    this.busyThumbnail = false;
                    this.thumbnail = res.url;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.thumbnail = res.url;
                    }
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit(EVENT_NAME_MEDIA_METADATA_CHANGE);
                })
                .onCancel(() => {
                    this.busyThumbnail = false;
                })
                .onRequestError((err) => {
                    this.busyThumbnail = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Invalid thumbnail provided"));
                        })
                        .add(401, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.ShowSnackBar(err.message);
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

            Request.Pending(this._handles.requestIdTitle, EditMediaAPI.ChangeMediaTitle(mediaId, this.title))
                .onSuccess(() => {
                    AppEvents.ShowSnackBar(this.$t("Successfully changed title"));
                    this.busyTitle = false;
                    this.originalTitle = this.title;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.title = this.title;
                    }
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit(EVENT_NAME_MEDIA_METADATA_CHANGE);
                })
                .onCancel(() => {
                    this.busyTitle = false;
                })
                .onRequestError((err) => {
                    this.busyTitle = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.ShowSnackBar(err.message);
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

            Request.Pending(this._handles.requestIdDescription, EditMediaAPI.ChangeMediaDescription(mediaId, this.desc))
                .onSuccess(() => {
                    AppEvents.ShowSnackBar(this.$t("Successfully changed description"));
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
                            AppEvents.ShowSnackBar(this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.ShowSnackBar(err.message);
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

            Request.Pending(this._handles.requestIdExtra, EditMediaAPI.ChangeExtraParams(mediaId, this.startBeginning))
                .onSuccess(() => {
                    AppEvents.ShowSnackBar(this.$t("Successfully changed media extra params"));
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
                            AppEvents.ShowSnackBar(this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.ShowSnackBar(this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.ShowSnackBar(err.message);
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
        this._handles.requestIdTitle = getUniqueStringId();
        this._handles.requestIdDescription = getUniqueStringId();
        this._handles.requestIdThumbnail = getUniqueStringId();
        this._handles.requestIdExtra = getUniqueStringId();

        this.updateMediaData();

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        MediaController.AddUpdateEventListener(this._handles.mediaUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AuthController.AddChangeEventListener(this._handles.authUpdateH);

        this.autoFocus();
    },

    beforeUnmount: function () {
        MediaController.RemoveUpdateEventListener(this._handles.mediaUpdateH);

        AuthController.RemoveChangeEventListener(this._handles.authUpdateH);

        Request.Abort(this._handles.requestIdTitle);
        Request.Abort(this._handles.requestIdDescription);
        Request.Abort(this._handles.requestIdThumbnail);
        Request.Abort(this._handles.requestIdExtra);
    },
});
</script>
