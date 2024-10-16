<template>
    <ModalDialogContainer
        :closeSignal="closeSignal"
        :forceCloseSignal="forceCloseSignal"
        v-model:display="displayStatus"
        :lock-close="busy"
    >
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Rename album") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Album name") }}:</label>
                    <input
                        type="text"
                        name="album-name"
                        autocomplete="off"
                        v-model="name"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Rename album") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsController, EVENT_NAME_CURRENT_ALBUM_UPDATED } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsRenameAlbum } from "@/api/api-albums";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";

export default defineComponent({
    components: {
        LoadingIcon,
    },
    name: "AlbumRenameModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            currentAlbum: -1,
            name: "",
            oldName: "",

            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,
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
                elem.focus();
                elem.select();
            });
        },

        onAlbumUpdate: function () {
            this.currentAlbum = AlbumsController.CurrentAlbum;
            if (AlbumsController.CurrentAlbumData) {
                this.oldName = AlbumsController.CurrentAlbumData.name;
                this.name = this.oldName;
            }
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            if (!this.name) {
                this.error = this.$t("Invalid album name provided");
                return;
            }

            if (this.name === this.oldName) {
                this.forceCloseSignal++;
                return;
            }

            this.busy = true;
            this.error = "";

            const albumId = this.currentAlbum;

            makeApiRequest(apiAlbumsRenameAlbum(albumId, this.name))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Album renamed") + ": " + this.name);
                    this.busy = false;
                    this.name = "";
                    this.forceCloseSignal++;
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
                        invalidName: () => {
                            this.error = this.$t("Invalid album name provided");
                        },
                        badRequest: () => {
                            this.error = this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        notFound: () => {
                            this.forceCloseSignal++;
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
            this.name = this.oldName;
            this.autoFocus();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.name = this.oldName;
                this.autoFocus();
            }
        },
    },
});
</script>
