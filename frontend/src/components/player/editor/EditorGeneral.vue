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
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { Request } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import { EVENT_NAME_MEDIA_METADATA_CHANGE, PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import {
    apiMediaChangeExtraParams,
    apiMediaChangeMediaDescription,
    apiMediaChangeMediaThumbnail,
    apiMediaChangeMediaTitle,
} from "@/api/api-media-edit";

export default defineComponent({
    components: {
        ToggleSwitch,
    },
    name: "EditorGeneral",
    emits: ["changed"],
    setup() {
        return {
            requestIdTitle: getUniqueStringId(),
            requestIdDescription: getUniqueStringId(),
            requestIdThumbnail: getUniqueStringId(),
            requestIdExtra: getUniqueStringId(),
        };
    },
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

            Request.Pending(this.requestIdThumbnail, apiMediaChangeMediaThumbnail(mediaId, file))
                .onSuccess((res) => {
                    PagesController.ShowSnackBar(this.$t("Successfully changed thumbnail"));
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
                .onRequestError((err, handleErr) => {
                    this.busyThumbnail = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidThumbnail: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid thumbnail provided"));
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

            Request.Pending(this.requestIdTitle, apiMediaChangeMediaTitle(mediaId, this.title))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Successfully changed title"));
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
                .onRequestError((err, handleErr) => {
                    this.busyTitle = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTitle: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid title provided"));
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
                    this.busyTitle = false;
                });
        },

        changeDescription: function () {
            if (this.busyDescription) {
                return;
            }

            this.busyDescription = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending(this.requestIdDescription, apiMediaChangeMediaDescription(mediaId, this.desc))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Successfully changed description"));
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
                .onRequestError((err, handleErr) => {
                    this.busyDescription = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidDescription: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid description provided"));
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
                    this.busyDescription = false;
                });
        },

        changeExtraParams: function () {
            if (this.busyExtra) {
                return;
            }

            this.busyExtra = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending(this.requestIdExtra, apiMediaChangeExtraParams(mediaId, this.startBeginning))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Successfully changed media extra params"));
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
                .onRequestError((err, handleErr) => {
                    this.busyExtra = false;
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
                    this.busyExtra = false;
                });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.autoFocus();
    },

    beforeUnmount: function () {
        Request.Abort(this.requestIdTitle);
        Request.Abort(this.requestIdDescription);
        Request.Abort(this.requestIdThumbnail);
        Request.Abort(this.requestIdExtra);
    },
});
</script>
