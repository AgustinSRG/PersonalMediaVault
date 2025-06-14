<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-lg" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Account security settings") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-else class="modal-body">
                <div class="form-group">
                    <label
                        >{{ $t("Two factor authentication") }}:
                        <span class="tfa-status" :class="{ enabled: tfa }">{{
                            tfa ? $t("Enabled") + " (" + tfaMethod + ")" : $t("Disabled")
                        }}</span></label
                    >
                </div>

                <div class="form-group">
                    <button v-if="tfa" type="button" class="btn btn-danger btn-sm auto-focus" @click="disableTfa">
                        <i class="fas fa-trash-alt"></i> {{ $t("Disable two factor authentication") }}
                    </button>
                    <button v-else type="button" class="btn btn-primary btn-sm auto-focus" @click="enableTfa">
                        <i class="fas fa-check"></i> {{ $t("Enable two factor authentication") }}
                    </button>
                </div>

                <div class="form-group">
                    <table class="table no-border">
                        <tbody>
                            <tr>
                                <td class="text-right td-shrink no-padding">
                                    <ToggleSwitch v-model:val="authConfirmation" @update:val="onChangesMade"></ToggleSwitch>
                                </td>
                                <td>
                                    {{
                                        $t(
                                            "Enable auth confirmation. (when enabled, you will be asked for a confirmation when performing dangerous actions)",
                                        )
                                    }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="form-group">
                    <table class="table no-border">
                        <tbody>
                            <tr>
                                <td class="text-right td-shrink no-padding">
                                    <ToggleSwitch v-model:val="authConfirmationPreferTfa" @update:val="onChangesMade"></ToggleSwitch>
                                </td>
                                <td>
                                    {{ $t("Prefer two factor authentication (if enabled) for auth confirmation.") }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="form-group">
                    <label
                        >{{
                            $t(
                                "Auth confirmation period, in seconds. (After a successful auth confirmation, you won't be asked for it again for a period of time)",
                            )
                        }}:</label
                    >
                    <input
                        v-model.number="authConfirmationPeriodSeconds"
                        type="number"
                        autocomplete="off"
                        :disabled="busy"
                        min="0"
                        max="2147483647"
                        class="form-control form-control-full-width"
                        @change="onChangesMade"
                    />
                </div>

                <div class="form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy || !dirty">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Save changes") }}
                </button>
            </div>
        </form>

        <AccountTfaEnableModal v-if="displayTfaEnableModal" v-model:display="displayTfaEnableModal" @done="load"></AccountTfaEnableModal>
        <AccountTfaDisableModal
            v-if="displayTfaDisableModal"
            v-model:display="displayTfaDisableModal"
            @done="load"
        ></AccountTfaDisableModal>

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
import { EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { apiAccountGetSecuritySettings, apiAccountSetSecuritySettings } from "@/api/api-account";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import { ProvidedAuthConfirmation } from "@/api/api-auth";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";

const AccountTfaEnableModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountTfaEnableModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountTfaDisableModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountTfaDisableModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

export default defineComponent({
    name: "AccountSecuritySettingsModal",
    components: {
        LoadingIcon,
        ToggleSwitch,
        AccountTfaEnableModal,
        AccountTfaDisableModal,
        AuthConfirmationModal,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            loadRequestId: getUniqueStringId(),
            saveRequestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            loading: true,

            tfa: false,
            tfaMethod: "TOTP",

            authConfirmation: false,
            authConfirmationPreferTfa: true,
            authConfirmationPeriodSeconds: 120,

            busy: false,

            error: "",

            dirty: false,

            closeSignal: 0,
            forceCloseSignal: 0,

            displayTfaEnableModal: false,
            displayTfaDisableModal: false,

            displayAuthConfirmation: false,
            authConfirmationCooldown: 0,
            authConfirmationTfa: false,
            authConfirmationError: "",
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.load();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.load.bind(this));

        if (this.display) {
            this.load();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);

        abortNamedApiRequest(this.saveRequestId);
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

        enableTfa: function () {
            this.displayTfaEnableModal = true;
        },

        disableTfa: function () {
            this.displayTfaDisableModal = true;
        },

        onChangesMade: function () {
            this.dirty = true;
        },

        load: function () {
            if (!this.display) {
                return;
            }

            this.loading = true;

            clearNamedTimeout(this.loadRequestId);

            makeNamedApiRequest(this.loadRequestId, apiAccountGetSecuritySettings())
                .onSuccess((response) => {
                    this.tfa = response.tfa;
                    this.tfaMethod = ((response.tfaMethod || "").split(":")[0] || "").toUpperCase();

                    this.authConfirmation = response.authConfirmation;
                    this.authConfirmationPreferTfa = response.authConfirmationMethod !== "pw";
                    this.authConfirmationPeriodSeconds = response.authConfirmationPeriodSeconds || 0;

                    this.loading = false;
                    this.dirty = false;
                    this.error = "";

                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.close();
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e?: Event) {
            if (e) {
                e.preventDefault();
            }

            this.submitInternal({});
        },

        submitInternal: function (confirmation: ProvidedAuthConfirmation) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(
                apiAccountSetSecuritySettings(
                    this.authConfirmation,
                    this.authConfirmationPreferTfa ? "tfa" : "pw",
                    this.authConfirmationPeriodSeconds,
                    confirmation,
                ),
            )
                .onSuccess(() => {
                    this.busy = false;
                    this.error = "";
                    this.dirty = false;
                    PagesController.ShowSnackBar(this.$t("Saved security settings"));
                    this.forceCloseSignal++;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    this.error = "";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidSettings: () => {
                            this.error = this.$t("Invalid security settings");
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
                    this.busy = false;
                    console.error(err);
                    this.error = err.message;
                });
        },
    },
});
</script>
