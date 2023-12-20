<template>
    <div class="player-editor-sub-content">
        <!--- Re-Encode -->

        <div class="form-group" v-if="canWrite">
            <label>
                {{ $t("If the media resource did not encode properly, try using the button below.") }}
                {{ $t("If it still does not work, try re-uploading the media.") }}
            </label>
        </div>
        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-primary" :disabled="busy" @click="encodeMedia">
                <i class="fas fa-sync-alt"></i> {{ $t("Re-Encode") }}
            </button>
        </div>

        <!--- Replace -->

        <div class="form-group" v-if="canWrite">
            <label>
                {{ $t("If you want to replace the media file, try using the button below.") }}
                {{ $t("You can use it to upgrade the media quality or fix any issues it may have.") }}
            </label>
        </div>
        <div class="form-group" v-if="canWrite">
            <input type="file" class="file-hidden replace-file-hidden" @change="replaceFileChanged" name="attachment-upload" />
            <button v-if="!replacing" type="button" class="btn btn-primary" :disabled="busy" @click="replaceMedia">
                <i class="fas fa-upload"></i> {{ $t("Replace media") }}
            </button>
            <button v-else-if="replacing && replaceProgress < 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ replaceProgress }}%)
            </button>
            <button v-else type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
        </div>

        <!--- Delete -->

        <div class="form-group" v-if="canWrite">
            <label>{{ $t("If you want to delete this media resource, click the button below.") }}</label>
        </div>
        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-danger" :disabled="busy" @click="deleteMedia">
                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
            </button>
        </div>

        <MediaDeleteModal v-if="displayMediaDelete" v-model:display="displayMediaDelete"></MediaDeleteModal>
        <ReEncodeConfirmationModal
            v-if="displayReEncode"
            v-model:display="displayReEncode"
            @confirm="doEncodeMedia"
        ></ReEncodeConfirmationModal>
        <ReplaceMediaConfirmationModal
            v-if="displayReplace"
            v-model:display="displayReplace"
            :file-name="replaceFileName"
            :file-size="replaceFileSize"
            @confirm="doReplaceMedia"
        ></ReplaceMediaConfirmationModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";

import MediaDeleteModal from "@/components/modals/MediaDeleteModal.vue";
import ReEncodeConfirmationModal from "@/components/modals/ReEncodeConfirmationModal.vue";
import ReplaceMediaConfirmationModal from "@/components/modals/ReplaceMediaConfirmationModal.vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaEncodeMedia, apiMediaReplaceMedia } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        MediaDeleteModal,
        ReEncodeConfirmationModal,
        ReplaceMediaConfirmationModal,
    },
    name: "EditorDangerZone",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
            fileRef: null,
        };
    },
    data: function () {
        return {
            type: 0,

            busy: false,

            canWrite: AuthController.CanWrite,

            displayMediaDelete: false,

            displayReEncode: false,

            displayReplace: false,

            replaceFileName: "",
            replaceFileSize: 0,

            replacing: false,
            replaceProgress: 0,
        };
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;
        },

        doEncodeMedia: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaEncodeMedia(mediaId))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully requested pending encoding tasks"));
                    this.busy = false;
                    MediaController.Load();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        encodeMedia: function () {
            this.displayReEncode = true;
        },

        deleteMedia: function () {
            this.displayMediaDelete = true;
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        replaceMedia: function () {
            const fileElem = this.$el.querySelector(".replace-file-hidden");
            if (fileElem) {
                fileElem.value = null;
                fileElem.click();
            }
        },

        replaceFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0] as File;
                this.fileRef = file;
                this.replaceFileName = file.name;
                this.replaceFileSize = file.size;
                this.displayReplace = true;
            }
        },

        doReplaceMedia: function () {
            if (this.busy) {
                return;
            }

            const file = this.fileRef;

            if (!file) {
                return;
            }

            this.busy = true;
            this.replacing = true;
            this.replaceProgress = 0;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaReplaceMedia(mediaId, file))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully uploaded") + ": " + file.name);
                    this.busy = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                    this.fileRef = null;

                    MediaController.Load();
                })
                .onUploadProgress((loaded, total) => {
                    if (total) {
                        this.replaceProgress = Math.floor(((loaded * 100) / total) * 100) / 100;
                    }
                })
                .onCancel(() => {
                    this.busy = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                    this.fileRef = null;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                    this.fileRef = null;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidMedia: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid media file provided"));
                        },
                        invalidMediaType: () => {
                            PagesController.ShowSnackBarRight(
                                this.$t("Error") + ": " + this.$t("You must upload a file of the same type in order to replace the media"),
                            );
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.busy = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                });
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
