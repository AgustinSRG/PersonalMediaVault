<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Change password") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Current password") }}:</label>
                    <input type="text" class="hidden-input" name="username" :value="username" />
                    <PasswordInput
                        v-model:val="currentPassword"
                        :name="'password'"
                        :disabled="busy"
                        :auto-focus="true"
                        @tab-skip="passwordTabSkip1"
                    ></PasswordInput>
                </div>
                <div class="form-group">
                    <label>{{ $t("New password") }}:</label>
                    <PasswordInput
                        v-model:val="password"
                        :name="'new-password'"
                        :disabled="busy"
                        :is-new-password="true"
                        @tab-skip="passwordTabSkip2"
                    ></PasswordInput>
                    <PasswordStrengthIndicator v-if="password" :password="password"></PasswordStrengthIndicator>
                </div>
                <div class="form-group">
                    <label>{{ $t("New password") }} ({{ $t("Repeat it for confirmation") }}):</label>
                    <PasswordInput
                        v-model:val="password2"
                        :name="'new-password-repeat'"
                        :disabled="busy"
                        :is-new-password="true"
                        @tab-skip="passwordTabSkip3"
                    ></PasswordInput>
                </div>

                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Change password") }}
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
import { apiAccountChangePassword } from "@/api/api-account";
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "@/components/utils/PasswordInput.vue";
import PasswordStrengthIndicator from "@/components/utils/PasswordStrengthIndicator.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "ChangePasswordModal",
    components: {
        LoadingIcon,
        PasswordInput,
        PasswordStrengthIndicator,
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
            username: AuthController.Username,

            currentPassword: "",
            password: "",
            password2: "",
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
                this.currentPassword = "";
                this.password = "";
                this.password2 = "";
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, () => {
            this.username = AuthController.Username;
        });

        if (this.display) {
            this.error = "";
            this.currentPassword = "";
            this.password = "";
            this.password2 = "";
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

        submit: function (e: Event) {
            e.preventDefault();

            this.submitInternal({});
        },

        submitInternal: function (confirmation: ProvidedAuthConfirmation) {
            if (this.busy) {
                return;
            }

            if (this.password !== this.password2) {
                this.error = this.$t("The passwords do not match");
                return;
            }

            this.busy = true;
            this.error = "";

            makeApiRequest(apiAccountChangePassword(this.currentPassword, this.password, confirmation))
                .onSuccess(() => {
                    this.busy = false;
                    this.currentPassword = "";
                    this.password = "";
                    this.password2 = "";
                    PagesController.ShowSnackBar(this.$t("Vault password changed!"));
                    this.forceCloseSignal++;
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
                        invalidNewPassword: () => {
                            this.error = this.$t("Invalid password provided");
                        },
                        invalidPassword: () => {
                            this.error = this.$t("Invalid password");
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

        close: function () {
            this.closeSignal++;
        },

        passwordTabSkip1: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector('[name="new-password"]');

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();

                if (typeof nextElement.select === "function") {
                    nextElement.select();
                }
            }
        },

        passwordTabSkip2: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector('[name="new-password-repeat"]');

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();

                if (typeof nextElement.select === "function") {
                    nextElement.select();
                }
            }
        },

        passwordTabSkip3: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector(".modal-footer-btn");

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();
            }
        },
    },
});
</script>
