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
                    <button
                        v-if="originalTitle !== title || busyTitle || !savedTitle"
                        type="submit"
                        class="btn btn-primary"
                        :disabled="busyTitle || originalTitle === title || !title"
                    >
                        <LoadingIcon icon="fas fa-pencil-alt" :loading="busyTitle"></LoadingIcon> {{ $t("Change title") }}
                    </button>
                    <button v-else type="button" disabled class="btn btn-primary">
                        <i class="fas fa-check"></i> {{ $t("Saved title") }}
                    </button>
                    <div v-if="errorTitle" class="form-error form-error-pt">{{ errorTitle }}</div>
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
                    v-if="originalDesc !== desc || busyDescription || !savedDescription"
                    type="button"
                    class="btn btn-primary"
                    :disabled="busyDescription || originalDesc === desc"
                    @click="changeDescription"
                >
                    <LoadingIcon icon="fas fa-pencil-alt" :loading="busyDescription"></LoadingIcon> {{ $t("Change description") }}
                </button>
                <button v-else type="button" disabled class="btn btn-primary">
                    <i class="fas fa-check"></i> {{ $t("Saved description") }}
                </button>
                <div v-if="errorDescription" class="form-error form-error-pt">{{ errorDescription }}</div>
            </div>

            <!--- Extra config -->

            <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
                <label>{{ $t("Extra media configuration") }}:</label>
            </div>
            <div class="table-responsive" v-if="type === 2 || type === 3">
                <table class="table no-border">
                    <tr v-if="type === 2 || type === 3">
                        <td class="text-right td-shrink">
                            <toggle-switch
                                v-model:val="startBeginning"
                                :disabled="busyExtra || !canWrite"
                                @update:val="changeExtraParams"
                            ></toggle-switch>
                        </td>
                        <td class="">
                            {{ $t("Reset time to the beginning every time the media reloads?") }}
                        </td>
                    </tr>

                    <tr v-if="type === 2">
                        <td class="text-right td-shrink">
                            <toggle-switch
                                v-model:val="isAnimation"
                                :disabled="busyExtra || !canWrite"
                                @update:val="changeExtraParams"
                            ></toggle-switch>
                        </td>
                        <td class="">
                            {{ $t("Is animation? (Force loop and disable keyboard time skipping)") }}
                        </td>
                    </tr>
                </table>
            </div>
            <div class="form-group loader-delayed-custom" v-if="busyExtra">
                <label><i class="fa fa-spinner fa-spin mr-1"></i> {{ $t("Saving changes") }}...</label>
            </div>
            <div v-if="errorExtraConfig" class="form-error form-error-pt">{{ errorExtraConfig }}</div>
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
                <button v-if="!busyThumbnail" type="button" class="btn btn-primary" @click="uploadThumbnail" @drop="onDrop">
                    <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
                </button>
                <button v-if="busyThumbnail" type="button" class="btn btn-primary" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading thumbnail") }}...
                </button>
                <div v-if="errorThumbnail" class="form-error form-error-pt">{{ errorThumbnail }}</div>
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
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
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
        LoadingIcon,
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

            savedTitle: false,
            savedDescription: false,
            savedExtra: false,

            canWrite: AuthController.CanWrite,

            originalStartBeginning: false,
            startBeginning: false,

            originalIsAnimation: false,
            isAnimation: false,

            errorTitle: "",
            errorDescription: "",
            errorExtraConfig: "",
            errorThumbnail: "",
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

            this.originalIsAnimation = MediaController.MediaData.is_anim;
            this.isAnimation = this.originalIsAnimation;

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

        inputFileChanged: function (e: InputEvent) {
            const data = (e.target as HTMLInputElement).files;
            if (data && data.length > 0) {
                const file = data[0];
                this.changeThumbnail(file);
            }
        },

        onDrop: function (e: DragEvent) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.changeThumbnail(file);
            }
        },

        changeThumbnail: function (file: File) {
            if (this.busyThumbnail) {
                return;
            }

            this.busyThumbnail = true;
            this.errorThumbnail = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestIdThumbnail, apiMediaChangeMediaThumbnail(mediaId, file))
                .onSuccess((res) => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed thumbnail"));
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
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidThumbnail: () => {
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Invalid thumbnail provided");
                        },
                        badRequest: () => {
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorThumbnail = this.$t("Error") + ": " + err.message;
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
            this.errorTitle = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestIdTitle, apiMediaChangeMediaTitle(mediaId, this.title))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed title"));
                    this.busyTitle = false;
                    this.savedTitle = true;
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
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTitle: () => {
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Invalid title provided");
                        },
                        badRequest: () => {
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.errorTitle = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorTitle = this.$t("Error") + ": " + err.message;
                    console.error(err);
                    this.busyTitle = false;
                });
        },

        changeDescription: function () {
            if (this.busyDescription) {
                return;
            }

            this.busyDescription = true;
            this.errorDescription = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestIdDescription, apiMediaChangeMediaDescription(mediaId, this.desc))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed description"));
                    this.busyDescription = false;
                    this.savedDescription = true;
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
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidDescription: () => {
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Invalid description provided");
                        },
                        badRequest: () => {
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.errorDescription = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorDescription = this.$t("Error") + ": " + err.message;
                    console.error(err);
                    this.busyDescription = false;
                });
        },

        changeExtraParams: function () {
            if (this.busyExtra) {
                return;
            }

            this.busyExtra = true;
            this.errorExtraConfig = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestIdExtra, apiMediaChangeExtraParams(mediaId, this.startBeginning, this.isAnimation))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed media extra params"));
                    this.busyExtra = false;
                    this.originalStartBeginning = this.startBeginning;
                    this.originalIsAnimation = this.isAnimation;
                    if (MediaController.MediaData) {
                        MediaController.MediaData.force_start_beginning = this.startBeginning;
                        MediaController.MediaData.is_anim = this.isAnimation;
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
                            this.errorExtraConfig = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            this.errorExtraConfig = this.$t("Error") + ": " + this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.errorExtraConfig = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.errorExtraConfig = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.errorExtraConfig = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.errorExtraConfig = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorExtraConfig = this.$t("Error") + ": " + err.message;
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
        abortNamedApiRequest(this.requestIdTitle);
        abortNamedApiRequest(this.requestIdDescription);
        abortNamedApiRequest(this.requestIdThumbnail);
        abortNamedApiRequest(this.requestIdExtra);
    },
});
</script>
