<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Create account") }}</div>
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
                    <label>{{ $t("Account password") }}:</label>
                    <PasswordInput
                        v-model:val="accountPassword"
                        :name="'account-new-password'"
                        :disabled="busy"
                        :is-new-password="true"
                        @tab-skip="passwordTabSkip1"
                    ></PasswordInput>
                    <PasswordStrengthIndicator v-if="accountPassword" :password="accountPassword"></PasswordStrengthIndicator>
                </div>

                <div class="form-group">
                    <label>{{ $t("Account password") }} ({{ $t("Again") }}):</label>
                    <PasswordInput
                        v-model:val="accountPassword2"
                        :name="'account-new-password-repeat'"
                        :disabled="busy"
                        :is-new-password="true"
                        @tab-skip="passwordTabSkip2"
                    ></PasswordInput>
                </div>

                <div class="form-group">
                    <label>{{ $t("Account type") }}:</label>
                    <select
                        v-model="accountWrite"
                        :disabled="busy"
                        name="account-type"
                        class="form-control form-select form-control-full-width"
                    >
                        <option :value="false">{{ $t("Read only") }}</option>
                        <option :value="true">{{ $t("Read / Write") }}</option>
                    </select>
                </div>

                <div class="form-group form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button type="submit" :disabled="busy" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-plus" :loading="busy"></LoadingIcon> {{ $t("Create account") }}
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
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAdminCreateAccount } from "@/api/api-admin";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "@/components/utils/PasswordInput.vue";
import PasswordStrengthIndicator from "@/components/utils/PasswordStrengthIndicator.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "AccountCreateModal",
    components: {
        LoadingIcon,
        PasswordInput,
        PasswordStrengthIndicator,
        AuthConfirmationModal,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display", "account-created"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            accountUsername: "",
            accountPassword: "",
            accountPassword2: "",
            accountWrite: false,

            busy: false,
            error: "",

            displayAuthConfirmation: false,
            authConfirmationCooldown: 0,
            authConfirmationTfa: false,
            authConfirmationError: "",

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
            this.autoFocus();
        }
    },
    beforeUnmount: function () {},
    methods: {
        autoFocus: function () {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        },

        submit: function (e: Event) {
            e.preventDefault();

            this.submitInternal({});
        },

        close: function () {
            this.closeSignal++;
        },

        passwordTabSkip1: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector('[name="account-new-password-repeat"]');

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();

                if (typeof nextElement.select === "function") {
                    nextElement.select();
                }
            }
        },

        passwordTabSkip2: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector('[name="account-type"]');

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();
            }
        },

        submitInternal: function (confirmation: ProvidedAuthConfirmation) {
            if (this.busy) {
                return;
            }

            if (this.accountPassword !== this.accountPassword2) {
                this.error = this.$t("The passwords do not match");
                return;
            }

            const username = this.accountUsername;
            const password = this.accountPassword;
            const write = this.accountWrite;

            this.busy = true;
            this.error = "";

            makeApiRequest(apiAdminCreateAccount(username, password, write, confirmation))
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Account created") + ": " + username);
                    this.accountUsername = "";
                    this.accountPassword = "";
                    this.accountPassword2 = "";
                    this.accountWrite = false;
                    this.$emit("account-created");
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
                        invalidUserPassword: () => {
                            this.error = this.$t("Invalid password provided");
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
