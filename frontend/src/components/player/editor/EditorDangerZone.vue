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
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
import { Request } from "@/utils/request";
import { defineComponent } from "vue";

import MediaDeleteModal from "@/components/modals/MediaDeleteModal.vue";
import ReEncodeConfirmationModal from "@/components/modals/ReEncodeConfirmationModal.vue";
import { EditMediaAPI } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        MediaDeleteModal,
        ReEncodeConfirmationModal,
    },
    name: "EditorDangerZone",
    emits: ["changed"],
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

            Request.Pending("media-editor-busy-danger", EditMediaAPI.EncodeMedia(mediaId))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully requested pending encoding tasks"));
                    this.busy = false;
                    MediaController.Load();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
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
        this._handles = Object.create(null);
        this.updateMediaData();

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this._handles.mediaUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        Request.Abort("media-editor-busy-danger");
    },
});
</script>
