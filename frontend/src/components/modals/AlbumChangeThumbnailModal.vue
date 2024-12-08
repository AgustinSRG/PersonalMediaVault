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
                <div class="form-group text-center">
                    <label v-if="!thumbnail">{{ $t("No thumbnail set for this album") }}</label>
                    <img
                        v-if="thumbnail"
                        :src="getThumbnail(thumbnail)"
                        :alt="$t('Thumbnail')"
                        class="form-group-thumbnail"
                        loading="lazy"
                    />
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
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsChangeAlbumThumbnail } from "@/api/api-albums";
import { getAssetURL } from "@/utils/api";
import { getUniqueStringId } from "@/utils/unique-id";

export default defineComponent({
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
        };
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            requestIdThumbnail: getUniqueStringId(),
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
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, this.onAlbumUpdate.bind(this));

        this.onAlbumUpdate();

        if (this.display) {
            this.errorThumbnail = "";
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        abortNamedApiRequest(this.requestIdThumbnail);
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
