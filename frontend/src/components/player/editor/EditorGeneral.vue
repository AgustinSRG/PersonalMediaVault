<template>
    <div class="player-editor-sub-content content-row">
        <div class="row-general-left">
            <!--- Title -->

            <form @submit="changeTitle">
                <div class="form-group">
                    <label>{{ $t("Title") }}:</label>
                    <input
                        v-model="title"
                        type="text"
                        autocomplete="off"
                        :readonly="!canWrite"
                        maxlength="255"
                        :disabled="busyTitle"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div v-if="canWrite" class="form-group">
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

            <!--- Extra config -->

            <div v-if="canWrite && (type === 2 || type === 3)" class="form-group">
                <label>{{ $t("Extra media configuration") }}:</label>
            </div>
            <div v-if="type === 2 || type === 3" class="table-responsive">
                <table class="table no-border">
                    <tbody>
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
                    </tbody>
                </table>
            </div>
            <div v-if="busyExtra" class="form-group loader-delayed-custom">
                <label><i class="fa fa-spinner fa-spin mr-1"></i> {{ $t("Saving changes") }}...</label>
            </div>
            <div v-if="errorExtraConfig" class="form-error form-error-pt">{{ errorExtraConfig }}</div>

            <!--- Tags -->

            <div class="form-group">
                <label>{{ $t("Tags") }}:</label>
            </div>

            <MediaTagsEditor @tags-update="onTagUpdate"></MediaTagsEditor>
        </div>

        <div class="row-general-right">
            <!--- Thumbnail -->

            <div class="form-group">
                <label>{{ $t("Thumbnail") }}:</label>
            </div>
            <div class="form-group" @drop="onDrop">
                <label v-if="!thumbnail">{{ $t("No thumbnail set for this media") }}</label>
                <ThumbImage v-if="thumbnail" :src="getThumbnail(thumbnail)" class-name="form-group-thumbnail"></ThumbImage>
            </div>
            <div v-if="canWrite" class="form-group">
                <input type="file" class="file-hidden" name="thumbnail-upload" @change="inputFileChanged" />
                <div class="text-center form-group-thumbnail-buttons">
                    <button
                        v-if="!busyThumbnail"
                        type="button"
                        class="btn btn-primary image-thumbnail-button"
                        @click="uploadThumbnail"
                        @drop="onDrop"
                    >
                        <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
                    </button>
                    <button v-if="busyThumbnail" type="button" class="btn btn-primary image-thumbnail-button" disabled>
                        <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading thumbnail") }}...
                    </button>
                </div>

                <div v-if="mediaElementReady" class="form-group-pt text-center form-group-thumbnail-buttons">
                    <button
                        v-if="type === 1"
                        type="button"
                        class="btn btn-primary btn-sm image-thumbnail-button"
                        :title="$t('Set current image as thumbnail')"
                        :disabled="busyThumbnail"
                        @click="setCurrentImageAsThumbnail"
                    >
                        <i class="fas fa-image"></i> {{ $t("Set current image as thumbnail") }}
                    </button>
                    <button
                        v-if="type === 2"
                        type="button"
                        class="btn btn-primary btn-sm image-thumbnail-button"
                        :title="$t('Set current frame as thumbnail')"
                        :disabled="busyThumbnail"
                        @click="setCurrentFrameAsThumbnail"
                    >
                        <i class="fas fa-image"></i> {{ $t("Set current frame as thumbnail") }}
                    </button>
                </div>
                <div v-if="errorThumbnail" class="form-error form-error-pt text-center">{{ errorThumbnail }}</div>
            </div>
        </div>

        <ThumbnailCropModal
            v-if="displayThumbnailModal"
            v-model:display="displayThumbnailModal"
            :image-url="thumbnailModalUrl"
            @done="changeThumbnail"
            @error="onThumbnailModalError"
        ></ThumbnailCropModal>

        <SaveChangesAskModal
            v-if="displayExitConfirmation"
            v-model:display="displayExitConfirmation"
            @yes="onExitSaveChanges"
            @no="onExitDiscardChanges"
        ></SaveChangesAskModal>
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
import { apiMediaChangeExtraParams, apiMediaChangeMediaThumbnail, apiMediaChangeMediaTitle } from "@/api/api-media-edit";
import ThumbImage from "@/components/utils/ThumbImage.vue";
import ThumbnailCropModal from "@/components/modals/ThumbnailCropModal.vue";
import MediaTagsEditor from "@/components/utils/MediaTagsEditor.vue";
import SaveChangesAskModal from "@/components/modals/SaveChangesAskModal.vue";
import { ExitPreventer } from "@/control/exit-prevent";

export default defineComponent({
    name: "EditorGeneral",
    components: {
        ToggleSwitch,
        LoadingIcon,
        ThumbImage,
        ThumbnailCropModal,
        MediaTagsEditor,
        SaveChangesAskModal,
    },
    emits: ["changed"],
    setup() {
        return {
            requestIdTitle: getUniqueStringId(),
            requestIdThumbnail: getUniqueStringId(),
            requestIdExtra: getUniqueStringId(),

            mediaElementCheckTimer: null as ReturnType<typeof setInterval> | null,

            tempImage: null as HTMLImageElement,

            exitCallback: null as () => void,
        };
    },
    data: function () {
        return {
            type: 0,

            title: "",
            originalTitle: "",

            thumbnail: "",

            originalStartBeginning: false,
            startBeginning: false,

            originalIsAnimation: false,
            isAnimation: false,

            busyTitle: false,
            busyThumbnail: false,
            busyExtra: false,

            savedTitle: false,
            savedExtra: false,

            errorTitle: "",
            errorExtraConfig: "",
            errorThumbnail: "",

            canWrite: AuthController.CanWrite,

            mediaElementReady: false,

            displayThumbnailModal: false,
            thumbnailModalUrl: "",

            displayExitConfirmation: false,
            exitOnSave: false,
        };
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.checkMediaElement();

        this.autoFocus();

        ExitPreventer.SetupExitPrevent(this.checkExitPrevent.bind(this), this.onExit.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestIdTitle);
        abortNamedApiRequest(this.requestIdThumbnail);
        abortNamedApiRequest(this.requestIdExtra);

        if (this.mediaElementCheckTimer) {
            clearTimeout(this.mediaElementCheckTimer);
        }

        if (this.tempImage) {
            delete this.tempImage.onload;
            delete this.tempImage.onerror;
        }

        ExitPreventer.RemoveExitPrevent();
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
                this.prepareChangeThumbnail(file);
            }
        },

        onDrop: function (e: DragEvent) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.prepareChangeThumbnail(file);
            }
        },

        prepareChangeThumbnail: function (file: File) {
            if (this.tempImage) {
                delete this.tempImage.onload;
                delete this.tempImage.onerror;
                this.tempImage = null;
            }

            const url = URL.createObjectURL(file);

            const img = new Image();
            this.tempImage = img;

            img.onload = () => {
                this.tempImage = null;

                if (img.width === img.height) {
                    this.changeThumbnail(file);
                    return;
                }

                this.thumbnailModalUrl = url;
                this.errorThumbnail = "";
                this.displayThumbnailModal = true;
            };

            img.onerror = (err) => {
                this.tempImage = null;
                console.error(err);
                this.onThumbnailModalError();
            };

            img.src = url;
        },

        onThumbnailModalError: function () {
            this.errorThumbnail = this.$t("Error") + ": " + this.$t("Invalid thumbnail provided");
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

        checkMediaElement: function () {
            if (this.mediaElementCheckTimer) {
                clearTimeout(this.mediaElementCheckTimer);
                this.mediaElementCheckTimer = null;
            }

            if (this.type === 1) {
                const imageElement = document.querySelector(".player-container .player-img-element") as HTMLImageElement;
                this.mediaElementReady = imageElement && imageElement.complete;
            } else if (this.type === 2) {
                const videoElement = document.querySelector(".player-container .player-video-element") as HTMLVideoElement;
                this.mediaElementReady = videoElement && videoElement.readyState === 4;
            } else {
                return;
            }

            if (!this.mediaElementReady) {
                this.mediaElementCheckTimer = setTimeout(this.checkMediaElement.bind(this), 1500);
            }
        },

        setCurrentFrameAsThumbnail: function () {
            if (this.busyThumbnail) {
                return;
            }

            const videoElement = document.querySelector(".player-container .player-video-element") as HTMLVideoElement;

            if (!videoElement || videoElement.readyState !== 4) {
                this.errorThumbnail =
                    this.$t("Error") +
                    ": " +
                    this.$t("Could not take the current frame") +
                    ". " +
                    this.$t("Maybe the media is not yet loaded?");
            }

            try {
                // Create canvas
                const canvas = document.createElement("canvas") as HTMLCanvasElement;

                canvas.height = videoElement.videoHeight;
                canvas.width = videoElement.videoWidth;

                //  Draw video frame to the canvas
                const ctx = canvas.getContext("2d");
                ctx.drawImage(videoElement, 0, 0, canvas.width, canvas.height);

                // Get frame as blob
                canvas.toBlob((blob) => {
                    // Convert to file
                    const file = new File([blob], "thumbnail.png");

                    // Change thumbnail
                    this.prepareChangeThumbnail(file);
                }, "image/png");
            } catch (ex) {
                console.error(ex);
                this.errorThumbnail =
                    this.$t("Error") +
                    ": " +
                    this.$t("Could not take the current frame") +
                    ". " +
                    this.$t("Maybe the media is not yet loaded?");
            }
        },

        setCurrentImageAsThumbnail: function () {
            if (this.busyThumbnail) {
                return;
            }

            const imageElement = document.querySelector(".player-container .player-img-element") as HTMLImageElement;

            if (!imageElement || !imageElement.complete) {
                this.errorThumbnail =
                    this.$t("Error") +
                    ": " +
                    this.$t("Could not find the current image") +
                    ". " +
                    this.$t("Maybe the media is not yet loaded?");
            }

            try {
                // Create canvas
                const canvas = document.createElement("canvas") as HTMLCanvasElement;

                canvas.width = imageElement.width;
                canvas.height = imageElement.height;

                //  Draw image to the canvas
                const ctx = canvas.getContext("2d");
                ctx.drawImage(imageElement, 0, 0, canvas.width, canvas.height);

                // Get image as blob
                canvas.toBlob((blob) => {
                    // Convert to file
                    const file = new File([blob], "thumbnail.png");

                    // Change thumbnail
                    this.prepareChangeThumbnail(file);
                }, "image/png");
            } catch (ex) {
                console.error(ex);
                this.errorThumbnail =
                    this.$t("Error") +
                    ": " +
                    this.$t("Could not find the current image") +
                    ". " +
                    this.$t("Maybe the media is not yet loaded?");
            }
        },

        changeTitle: function (e?: Event) {
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

                    if (this.exitOnSave) {
                        this.exitOnSave = false;
                        if (this.exitCallback) {
                            this.exitCallback();
                        }
                    }
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

        onTagUpdate: function () {
            this.$emit("changed");
        },

        checkExitPrevent: function (): boolean {
            return this.originalTitle !== this.title;
        },

        onExit: function (callback: () => void) {
            this.exitCallback = callback;
            this.displayExitConfirmation = true;
        },

        onExitSaveChanges: function () {
            if (this.originalTitle !== this.title) {
                this.exitOnSave = true;
                this.changeTitle();
            } else {
                if (this.exitCallback) {
                    this.exitCallback();
                }
            }
        },

        onExitDiscardChanges: function () {
            if (this.exitCallback) {
                this.exitCallback();
            }
        },
    },
});
</script>
