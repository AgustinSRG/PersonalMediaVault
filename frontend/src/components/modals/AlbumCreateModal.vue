<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" :lock-close="busy">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Create new album") }}
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
                    <i class="fas fa-plus"></i> {{ $t("Create album") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsAPI } from "@/api/api-albums";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "AlbumCreateModal",
    emits: ["update:display", "new-album"],
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            name: "",

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

        close: function () {
            this.$refs.modalContainer.close();
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

            if (AlbumsController.FindDuplicatedName(this.name)) {
                this.error = this.$t("There is already another album with the same name");
                return;
            }

            this.busy = true;
            this.error = "";

            const albumName = this.name;

            Request.Do(AlbumsAPI.CreateAlbum(albumName))
                .onSuccess((response) => {
                    AppEvents.Emit("snack", this.$t("Album created") + ": " + albumName);
                    this.busy = false;
                    this.name = "";
                    this.close();
                    AppEvents.Emit("albums-list-change");
                    AlbumsController.Load();
                    this.$emit("new-album", response.album_id, albumName);
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            this.error = this.$t("Invalid album name provided");
                        })
                        .add(401, "*", () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            this.error = this.$t("Access denied");
                        })
                        .add(500, "*", () => {
                            this.error = this.$t("Internal server error");
                        })
                        .add("*", "*", () => {
                            this.error = this.$t("Could not connect to the server");
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.busy = false;
                });
        },
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
            this.autoFocus();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.autoFocus();
            }
        },
    },
});
</script>
