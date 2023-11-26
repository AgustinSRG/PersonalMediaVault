<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" :lock-close="busy">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete album") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>
                        {{ $t("Remember. If you delete the album by accident you would have to recreate it.") }}
                        {{ $t("Make sure you actually want to delete it.") }}
                    </label>
                </div>
                <div class="form-group">
                    <label>{{ $t("Type 'confirm' for confirmation") }}:</label>
                    <input
                        type="text"
                        name="confirmation"
                        autocomplete="off"
                        v-model="confirmation"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete album") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsController, EVENT_NAME_CURRENT_ALBUM_UPDATED } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { Request } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsDeleteAlbum } from "@/api/api-albums";

export default defineComponent({
    name: "AlbumDeleteModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            currentAlbum: -1,
            oldName: "",

            confirmation: "",

            busy: false,
            error: "",
        };
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
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
                this.oldName = AlbumsController.CurrentAlbumData.name;
            }
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        submit: function (e) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            if (this.confirmation.toLowerCase() !== "confirm") {
                this.error = this.$t("You must type 'confirm' in order to confirm the deletion of the album");
                return;
            }

            this.busy = true;
            this.error = "";

            const albumId = this.currentAlbum;

            Request.Do(apiAlbumsDeleteAlbum(albumId))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Album deleted") + ": " + this.oldName);
                    this.busy = false;
                    this.confirmation = "";
                    this.$refs.modalContainer.close(true);
                    AlbumsController.OnChangedAlbum(albumId);
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
                            AlbumsController.OnChangedAlbum(albumId);
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
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, this.onAlbumUpdate.bind(this));

        this.onAlbumUpdate();
        if (this.display) {
            this.error = "";
            this.confirmation = "";
            this.autoFocus();
        }
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
});
</script>
