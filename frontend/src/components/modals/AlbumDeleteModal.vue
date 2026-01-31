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
                    {{ $t("Delete album") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <table class="table no-margin no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="confirmation"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the album by accident you would have to recreate it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy || !confirmation" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Delete album") }}
                </button>
            </div>
        </form>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="submitInternal"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { emitAppEvent, EVENT_NAME_CURRENT_ALBUM_UPDATED, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import { PagesController } from "@/control/pages";
import { apiAlbumsDeleteAlbum } from "@/api/api-albums";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "AlbumDeleteModal",
    components: {
        LoadingIcon,
        ToggleSwitch,
        AuthConfirmationModal,
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
            currentAlbum: -1,
            oldName: "",

            confirmation: false,

            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,

            displayAuthConfirmation: false,
            authConfirmationCooldown: 0,
            authConfirmationTfa: false,
            authConfirmationError: "",
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.confirmation = false;
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, this.onAlbumUpdate.bind(this));

        this.onAlbumUpdate();
        if (this.display) {
            this.error = "";
            this.confirmation = false;
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

        onAlbumUpdate: function () {
            this.currentAlbum = AlbumsController.CurrentAlbum;
            if (AlbumsController.CurrentAlbumData) {
                this.oldName = AlbumsController.CurrentAlbumData.name;
            }
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e: Event) {
            e.preventDefault();

            this.submitInternal({});
        },

        submitInternal: function (confirmation: ProvidedAuthConfirmation) {
            if (this.busy) {
                return;
            }

            this.busy = true;
            this.error = "";

            const albumId = this.currentAlbum;

            makeApiRequest(apiAlbumsDeleteAlbum(albumId, confirmation))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Album deleted") + ": " + this.oldName);
                    this.busy = false;
                    this.confirmation = false;
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
                            emitAppEvent(EVENT_NAME_UNAUTHORIZED);
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
});
</script>
