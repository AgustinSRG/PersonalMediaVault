<template>
    <ModalDialogContainer
        :closeSignal="closeSignal"
        :forceCloseSignal="forceCloseSignal"
        v-model:display="displayStatus"
        :lock-close="busyThumbnail"
    >
        <div v-if="display" class="modal-dialog modal-md" role="document" @drop="onDrop">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Change album thumbnail") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group flex-center">
                    <label v-if="!thumbnail">{{ $t("No thumbnail set for this album") }}</label>
                    <ThumbImage v-if="thumbnail" :src="getThumbnail(thumbnail)" className="form-group-thumbnail"></ThumbImage>
                </div>
                <div class="form-group">
                    <input type="file" class="file-hidden" @change="inputFileChanged" name="thumbnail-upload" />
                    <div class="text-center">
                        <button v-if="!busyThumbnail" type="button" class="btn btn-primary image-thumbnail-button" @click="uploadThumbnail">
                            <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
                        </button>
                        <button v-if="busyThumbnail" type="button" class="btn btn-primary image-thumbnail-button" disabled>
                            <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading thumbnail") }}...
                        </button>
                    </div>
                    <div class="text-center form-group-pt" v-if="currentMediaThumbnail">
                        <button
                            type="button"
                            class="btn btn-primary btn-sm image-thumbnail-button"
                            :title="$t('Set current media thumbnail')"
                            @click="setCurrentMediaThumbnail"
                            :disabled="busyThumbnail"
                        >
                            <i class="fas fa-image"></i> {{ $t("Set current media thumbnail") }}
                        </button>
                    </div>
                    <div v-if="errorThumbnail" class="form-error form-error-pt text-center">{{ errorThumbnail }}</div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busyThumbnail" type="button" class="modal-footer-btn" @click="close">
                    <i class="fas fa-check"></i> {{ $t("Done") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsController, EVENT_NAME_CURRENT_ALBUM_UPDATED } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsChangeAlbumThumbnail } from "@/api/api-albums";
import { getAssetURL } from "@/utils/api";
import { getUniqueStringId } from "@/utils/unique-id";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import ThumbImage from "../utils/ThumbImage.vue";

export default defineComponent({
    components: {
        ThumbImage,
    },
    name: "AlbumChangeThumbnailModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            currentAlbum: -1,

            thumbnail: "",

            closeSignal: 0,
            forceCloseSignal: 0,

            busyThumbnail: false,
            errorThumbnail: "",

            currentMediaThumbnail: "",
        };
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            requestIdThumbnail: getUniqueStringId(),
            fetchMediaThumbnailAbortController: null as AbortController | null,
        };
    },
    methods: {
        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        onAlbumUpdate: function () {
            this.currentAlbum = AlbumsController.CurrentAlbum;
            if (AlbumsController.CurrentAlbumData) {
                this.thumbnail = AlbumsController.CurrentAlbumData.thumbnail || "";
            }
        },

        close: function () {
            this.closeSignal++;
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

            const albumId = this.currentAlbum;

            makeNamedApiRequest(this.requestIdThumbnail, apiAlbumsChangeAlbumThumbnail(albumId, file))
                .onSuccess((res) => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed thumbnail"));
                    this.busyThumbnail = false;
                    this.thumbnail = res.url;
                    AlbumsController.OnChangedAlbum(albumId);
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

        onMediaUpdate: function () {
            if (MediaController.MediaData) {
                this.currentMediaThumbnail = MediaController.MediaData.thumbnail || "";
            } else {
                this.currentMediaThumbnail = "";
            }
        },

        setCurrentMediaThumbnail: function () {
            if (this.busyThumbnail) {
                return;
            }

            if (!this.currentMediaThumbnail) {
                return;
            }

            const thumbnailUrl = getAssetURL(this.currentMediaThumbnail);

            this.busyThumbnail = true;

            const abortController = new AbortController();
            this.fetchMediaThumbnailAbortController = abortController;

            fetch(thumbnailUrl, {
                signal: abortController.signal,
                headers: {
                    "x-session-token": AuthController.Session,
                },
            })
                .then((response) => {
                    this.fetchMediaThumbnailAbortController = null;
                    if (response.status === 200) {
                        response
                            .blob()
                            .then((blob) => {
                                this.busyThumbnail = false;
                                this.changeThumbnail(new File([blob], "thumbnail.jpg"));
                            })
                            .catch((err) => {
                                console.error(err);
                                this.errorThumbnail = this.$t("Error") + ": " + err.message;
                                this.busyThumbnail = false;
                            });
                    } else {
                        this.errorThumbnail = this.$t("Error") + ": " + this.$t("Could not fetch media thumbnail");
                        this.busyThumbnail = false;
                    }
                })
                .catch((err) => {
                    this.fetchMediaThumbnailAbortController = null;
                    console.error(err);
                    this.errorThumbnail = this.$t("Error") + ": " + err.message;
                    this.busyThumbnail = false;
                });
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, this.onAlbumUpdate.bind(this));

        this.onAlbumUpdate();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.onMediaUpdate.bind(this));

        this.onMediaUpdate();

        if (this.display) {
            this.errorThumbnail = "";
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        abortNamedApiRequest(this.requestIdThumbnail);

        if (this.fetchMediaThumbnailAbortController) {
            this.fetchMediaThumbnailAbortController.abort();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.errorThumbnail = "";
                this.autoFocus();
            }
        },
    },
});
</script>
