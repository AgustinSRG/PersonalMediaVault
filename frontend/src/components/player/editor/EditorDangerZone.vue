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

        <!--- Delete -->

        <div class="form-group" v-if="canWrite">
            <label>{{ $t("If you want to delete this media resource, click the button below.") }}</label>
        </div>
        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-danger" :disabled="busy" @click="deleteMedia">
                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
            </button>
        </div>

        <MediaDeleteModal v-model:display="displayMediaDelete"></MediaDeleteModal>
        <ReEncodeConfirmationModal v-model:display="displayReEncode" @confirm="doEncodeMedia"></ReEncodeConfirmationModal>
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
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaEncodeMedia } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        MediaDeleteModal,
        ReEncodeConfirmationModal,
    },
    name: "EditorDangerZone",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            busy: false,

            canWrite: AuthController.CanWrite,

            displayMediaDelete: false,

            displayReEncode: false,
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
                    PagesController.ShowSnackBar(this.$t("Successfully requested pending encoding tasks"));
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
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
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
