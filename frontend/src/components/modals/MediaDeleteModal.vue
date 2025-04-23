<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete media") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>
                        {{ $t("Remember. If you delete the media by accident you would have to re-upload it.") }}
                        {{ $t("Make sure you actually want to delete it.") }}
                    </label>
                </div>
                <div class="form-group">
                    <label>{{ $t("Type 'confirm' for confirmation") }}:</label>
                    <input
                        v-model="confirmation"
                        type="text"
                        name="confirmation"
                        autocomplete="off"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Delete media") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiMediaDeleteMedia } from "@/api/api-media-edit";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";

export default defineComponent({
    name: "MediaDeleteModal",
    components: {
        LoadingIcon,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            currentMedia: -1,
            oldName: "",

            confirmation: "",

            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.confirmation = "";
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onMediaUpdate.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.onMediaUpdate.bind(this));

        this.onMediaUpdate();

        if (this.display) {
            this.error = "";
            this.confirmation = "";
            this.autoFocus();
        }
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

        onMediaUpdate: function () {
            this.currentMedia = AppStatus.CurrentMedia;
            if (MediaController.MediaData) {
                this.oldName = MediaController.MediaData.title;
            }
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e: Event) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            if (this.confirmation.toLowerCase() !== "confirm") {
                this.error = this.$t("You must type 'confirm' in order to confirm the deletion of the media");
                return;
            }

            this.busy = true;
            this.error = "";

            const mediaId = this.currentMedia;

            makeApiRequest(apiMediaDeleteMedia(mediaId))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Media deleted") + ": " + this.oldName);
                    this.busy = false;
                    this.confirmation = "";
                    this.forceCloseSignal++;
                    AlbumsController.LoadCurrentAlbum();
                    AppStatus.OnDeleteMedia();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        notFound: () => {
                            this.error = this.$t("Not found");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.busy = false;
                });
        },
    },
});
</script>
