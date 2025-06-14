<template>
    <div class="player-editor-sub-content">
        <!--- Re-Encode -->

        <div v-if="canWrite" class="form-group">
            <label>
                {{ $t("If the media resource did not encode properly, try using the button below.") }}
                {{ $t("If it still does not work, try re-uploading the media.") }}
            </label>
        </div>
        <div v-if="canWrite" class="form-group">
            <button type="button" class="btn btn-primary" :disabled="busyReEncode || busyReplace" @click="encodeMedia">
                <LoadingIcon icon="fas fa-sync-alt" :loading="busyReEncode"></LoadingIcon> {{ $t("Re-Encode") }}
            </button>
            <div v-if="errorReEncode" class="form-error form-error-pt">{{ errorReEncode }}</div>
        </div>

        <!--- Replace -->

        <div v-if="canWrite" class="form-group">
            <label>
                {{ $t("If you want to replace the media file, try using the button below.") }}
                {{ $t("You can use it to upgrade the media quality or fix any issues it may have.") }}
            </label>
        </div>
        <div v-if="canWrite" class="form-group">
            <input type="file" class="file-hidden replace-file-hidden" name="attachment-upload" @change="replaceFileChanged" />
            <button
                v-if="!replacing"
                type="button"
                class="btn btn-primary"
                :disabled="busyReEncode || busyReplace"
                @click="replaceMedia"
                @drop="replaceMediaDrop"
            >
                <i class="fas fa-upload"></i> {{ $t("Replace media") }}
            </button>
            <button v-else-if="replacing && replaceProgress < 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ replaceProgress }}%)
            </button>
            <button v-else type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
            <div v-if="errorReplace" class="form-error form-error-pt">{{ errorReplace }}</div>
        </div>

        <!--- Delete -->

        <div v-if="canWrite" class="form-group">
            <label>{{ $t("If you want to delete this media resource, click the button below.") }}</label>
        </div>
        <div v-if="canWrite" class="form-group">
            <button type="button" class="btn btn-danger" :disabled="busyReEncode || busyReplace" @click="deleteMedia">
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

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="doReplaceMediaInternal"
        ></AuthConfirmationModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import MediaDeleteModal from "@/components/modals/MediaDeleteModal.vue";
import ReEncodeConfirmationModal from "@/components/modals/ReEncodeConfirmationModal.vue";
import ReplaceMediaConfirmationModal from "@/components/modals/ReplaceMediaConfirmationModal.vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaEncodeMedia, apiMediaReplaceMedia } from "@/api/api-media-edit";
import AuthConfirmationModal from "@/components/modals/AuthConfirmationModal.vue";
import { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "EditorDangerZone",
    components: {
        MediaDeleteModal,
        ReEncodeConfirmationModal,
        ReplaceMediaConfirmationModal,
        LoadingIcon,
        AuthConfirmationModal,
    },
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
            fileRef: null as File | null,
        };
    },
    data: function () {
        return {
            type: 0,

            busyReEncode: false,
            busyReplace: false,

            canWrite: AuthController.CanWrite,

            displayMediaDelete: false,

            displayReEncode: false,

            displayReplace: false,

            replaceFileName: "",
            replaceFileSize: 0,

            replacing: false,
            replaceProgress: 0,

            errorReplace: "",
            errorReEncode: "",

            displayAuthConfirmation: false,
            authConfirmationCooldown: 0,
            authConfirmationTfa: false,
            authConfirmationError: "",
        };
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;
        },

        doEncodeMedia: function () {
            if (this.busyReEncode) {
                return;
            }

            this.busyReEncode = true;
            this.errorReEncode = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaEncodeMedia(mediaId))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully requested pending encoding tasks"));
                    this.busyReEncode = false;
                    MediaController.Load();
                })
                .onCancel(() => {
                    this.busyReEncode = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busyReEncode = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.errorReEncode = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.errorReEncode = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.errorReEncode = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.errorReEncode = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.errorReEncode = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorReEncode = this.$t("Error") + ": " + err.message;
                    console.error(err);
                    this.busyReEncode = false;
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

        replaceFileChanged: function (e: InputEvent) {
            const data = (e.target as HTMLInputElement).files;
            if (data && data.length > 0) {
                const file = data[0] as File;
                this.fileRef = file;
                this.replaceFileName = file.name;
                this.replaceFileSize = file.size;
                this.displayReplace = true;
            }
        },

        replaceMediaDrop: function (e: DragEvent) {
            e.preventDefault();

            if (this.busyReplace) {
                return;
            }

            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0] as File;
                this.fileRef = file;
                this.replaceFileName = file.name;
                this.replaceFileSize = file.size;
                this.displayReplace = true;
            }
        },

        doReplaceMedia: function () {
            this.doReplaceMediaInternal({});
        },

        doReplaceMediaInternal: function (confirmation: ProvidedAuthConfirmation) {
            if (this.busyReplace) {
                return;
            }

            const file = this.fileRef;

            if (!file) {
                return;
            }

            this.busyReplace = true;
            this.replacing = true;
            this.replaceProgress = 0;
            this.errorReplace = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaReplaceMedia(mediaId, file, confirmation))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully uploaded") + ": " + file.name);
                    this.busyReplace = false;
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
                    this.busyReplace = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                    this.fileRef = null;
                })
                .onRequestError((err, handleErr) => {
                    this.busyReplace = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                    handleErr(err, {
                        unauthorized: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidMedia: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Invalid media file provided");
                        },
                        invalidMediaType: () => {
                            this.fileRef = null;
                            this.errorReplace =
                                this.$t("Error") + ": " + this.$t("You must upload a file of the same type in order to replace the media");
                        },
                        badRequest: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Bad request");
                        },
                        requiredAuthConfirmationPassword: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = "";
                            this.authConfirmationTfa = false;
                        },
                        invalidPassword: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("Invalid password");
                            this.authConfirmationTfa = false;
                            this.authConfirmationCooldown = Date.now() + 5000;
                        },
                        requiredAuthConfirmationTfa: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = "";
                            this.authConfirmationTfa = true;
                        },
                        invalidTfaCode: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("Invalid one-time code");
                            this.authConfirmationTfa = true;
                            this.authConfirmationCooldown = Date.now() + 5000;
                        },
                        cooldown: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("You must wait 5 seconds to try again");
                        },
                        accessDenied: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.fileRef = null;
                            this.errorReplace = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorReplace = this.$t("Error") + ": " + err.message;
                    console.error(err);
                    this.busyReplace = false;
                    this.replacing = false;
                    this.replaceProgress = 0;
                });
        },
    },
});
</script>
