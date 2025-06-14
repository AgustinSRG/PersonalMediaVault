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
                    {{ $t("Modify account") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Account name") }}:</label>
                    <input
                        v-model="accountUsername"
                        type="text"
                        autocomplete="off"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Account type") }}:</label>
                    <select v-model="accountWrite" :disabled="busy" class="form-control form-select form-control-full-width">
                        <option :value="false">{{ $t("Read only") }}</option>
                        <option :value="true">{{ $t("Read / Write") }}</option>
                    </select>
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy || (accountWrite === write && accountUsername === username)" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Modify account") }}
                </button>
            </div>
        </form>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            @confirm="submitInternal"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { apiAdminUpdateAccount } from "@/api/api-admin";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "AccountModifyModal",
    components: {
        LoadingIcon,
        AuthConfirmationModal,
    },
    props: {
        display: Boolean,
        username: String,
        write: Boolean,
    },
    emits: ["update:display", "done"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            name: "",

            busy: false,
            error: "",

            accountUsername: this.username || "",
            accountWrite: this.write || false,

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
                this.autoFocus();
            }
        },
        username: function () {
            this.accountUsername = this.username || "";
        },
        write: function () {
            this.accountWrite = this.write || false;
        },
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
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

            makeApiRequest(apiAdminUpdateAccount(this.username, this.accountUsername, this.accountWrite, confirmation))
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Account updated") + ": " + this.accountUsername);
                    this.$emit("done");
                    this.close();
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
                        invalidUsername: () => {
                            this.error = this.$t("Invalid username provided");
                        },
                        usernameInUse: () => {
                            this.error = this.$t("The username is already in use");
                        },
                        badRequest: () => {
                            this.error = this.$t("Bad request");
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
                        accountNotFound: () => {
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
