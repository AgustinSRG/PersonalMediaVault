<template>
    <ModalDialogContainer
        :closeSignal="closeSignal"
        :forceCloseSignal="forceCloseSignal"
        v-model:display="displayStatus"
        :lock-close="busy"
    >
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Change password") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Current password") }}:</label>
                    <input
                        type="password"
                        name="password"
                        v-model="currentPassword"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("New password") }}:</label>
                    <input
                        type="password"
                        name="new-password"
                        v-model="password"
                        :disabled="busy"
                        autocomplete="new-password"
                        maxlength="255"
                        class="form-control form-control-full-width"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("New password") }} ({{ $t("Repeat it for confirmation") }}):</label>
                    <input
                        type="password"
                        name="new-password-repeat"
                        v-model="password2"
                        :disabled="busy"
                        autocomplete="new-password"
                        maxlength="255"
                        class="form-control form-control-full-width"
                    />
                </div>

                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Change password") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { apiAccountChangePassword } from "@/api/api-account";
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";

export default defineComponent({
    components: {
        LoadingIcon,
    },
    name: "ChangePasswordModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            currentPassword: "",
            password: "",
            password2: "",
            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,
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

        submit: function (e) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            if (this.password !== this.password2) {
                this.error = this.$t("The passwords do not match");
                return;
            }

            this.busy = true;
            this.error = "";

            makeApiRequest(apiAccountChangePassword(this.currentPassword, this.password))
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
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
            this.currentPassword = "";
            this.password = "";
            this.password2 = "";
            this.autoFocus();
        }
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
});
</script>
