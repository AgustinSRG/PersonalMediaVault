<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-lg" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Enable two factor authentication") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body">
                <div class="tfa-qr-container">
                    <div class="tfa-qr">
                        <img v-if="qr && !errorInvalidSettings" :src="qr" class="tfa-qr-image" />
                        <div v-if="loadingSettings" class="tfa-qr-loader">
                            <i class="fa fa-spinner fa-spin"></i>
                        </div>
                        <div v-if="!loadingSettings && errorInvalidSettings" class="tfa-qr-invalid">
                            <i class="fas fa-ban"></i>
                        </div>
                    </div>
                </div>

                <div class="form-group">
                    <label>{{ $t("Scan the QR code with an authenticator mobile application in order to add the profile.") }}</label>
                </div>

                <div class="form-group">
                    <label>{{ $t("Two factor authentication URL (back it up in case you lose access to your mobile device)") }}:</label>
                    <input type="text" readonly :value="url" class="form-control form-control-full-width" />
                </div>

                <div class="form-group">
                    <button type="button" class="btn btn-sm btn-primary btn-mr" :disabled="busy || !url" @click="copyUrl">
                        <i class="fas fa-clone"></i> {{ $t("Copy URL to clipboard") }}
                    </button>
                    <button type="button" class="btn btn-sm btn-primary btn-mr" :disabled="busy" @click="customizeSettings">
                        <i class="fas fa-cog"></i> {{ $t("Customize algorithm settings") }}
                    </button>
                </div>

                <div class="form-group">
                    <label>{{ $t("To confirm the operation, type your account password") }}:</label>
                    <PasswordInput
                        v-model:val="password"
                        :name="'password'"
                        :disabled="busy || loadingSettings || errorInvalidSettings"
                        :auto-focus="true"
                        @tab-skip="passwordTabSkip"
                    ></PasswordInput>

                    <div v-if="errorPassword" class="form-error form-error-pt">{{ errorPassword }}</div>
                </div>

                <div class="form-group">
                    <label>{{ $t("Finally, input your one-time code") }}:</label>
                    <SixDigitCodeInput
                        v-model:val="code"
                        :min="true"
                        :disabled="busy || loadingSettings || errorInvalidSettings"
                    ></SixDigitCodeInput>

                    <div v-if="errorCode" class="form-error form-error-pt">{{ errorCode }}</div>
                </div>

                <div v-if="error" class="form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy || loadingSettings || errorInvalidSettings">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Enable two factor authentication") }}
                </button>
            </div>
        </form>

        <AccountTfaSettingsModal
            v-if="displayCustomSettings"
            v-model:display="displayCustomSettings"
            v-model:issuer="issuer"
            v-model:account="account"
            v-model:algorithm="algorithm"
            v-model:period="period"
            v-model:skew="skew"
            @update:display="onCloseCustomSettingsModal"
        ></AccountTfaSettingsModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "../utils/PasswordInput.vue";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import { apiAccountTimeOtpEnable, apiAccountTimeOtpSettings, TimeOtpAlgorithm, TimeOtpPeriod } from "@/api/api-account";
import AccountTfaSettingsModal from "./AccountTfaSettingsModal.vue";

export default defineComponent({
    name: "AccountTfaEnableModal",
    components: {
        LoadingIcon,
        PasswordInput,
        SixDigitCodeInput,
        AccountTfaSettingsModal,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display", "done"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            loadRequestId: getUniqueStringId(),
            saveRequestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            issuer: "PMV",
            account: AuthController.Username,

            algorithm: "sha1" as TimeOtpAlgorithm,

            period: "60" as TimeOtpPeriod,

            skew: true,

            loadingSettings: false,

            displayCustomSettings: false,

            secret: "",
            method: "",
            url: "",
            qr: "",

            password: "",
            code: "",

            errorInvalidSettings: false,

            busy: false,

            error: "",
            errorPassword: "",
            errorCode: "",

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.loadSettings();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, () => {
            this.account = AuthController.Username;
            this.loadSettings();
        });

        if (this.display) {
            this.loadSettings();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);

        abortNamedApiRequest(this.saveRequestId);
    },
    methods: {
        autoFocus: function () {
            if (!this.display || this.displayCustomSettings) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        customizeSettings: function () {
            this.displayCustomSettings = true;
        },

        onCloseCustomSettingsModal: function () {
            if (!this.displayCustomSettings) {
                this.loadSettings();
            }
        },

        copyUrl: function () {
            navigator.clipboard.writeText(this.url);
            PagesController.ShowSnackBar(this.$t("Copied URL to clipboard"));
        },

        passwordTabSkip: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector(".form-control.code-char-0");

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();
            }
        },

        loadSettings: function () {
            if (!this.display) {
                return;
            }

            this.loadingSettings = true;
            this.errorInvalidSettings = false;

            clearNamedTimeout(this.loadRequestId);

            makeNamedApiRequest(
                this.loadRequestId,
                apiAccountTimeOtpSettings({
                    issuer: this.issuer,
                    account: this.account,
                    algorithm: this.algorithm,
                    period: this.period,
                    skew: this.skew ? "allow" : "disallow",
                }),
            )
                .onSuccess((response) => {
                    this.secret = response.secret;
                    this.method = response.method;
                    this.url = response.url;
                    this.qr = response.qr;

                    this.code = "";

                    this.loadingSettings = false;
                    this.errorInvalidSettings = false;

                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidSettings: () => {
                            this.loadingSettings = false;
                            this.error = this.$t("Invalid two factor authentication settings. Try with another configuration.");
                            this.errorInvalidSettings = true;
                        },
                        accessDenied: () => {
                            this.close();
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.loadSettings.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.loadSettings.bind(this));
                });
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e?: Event) {
            if (e) {
                e.preventDefault();
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(
                apiAccountTimeOtpEnable({
                    secret: this.secret,
                    method: this.method,
                    password: this.password,
                    code: this.code,
                }),
            )
                .onSuccess(() => {
                    this.busy = false;

                    this.error = "";
                    this.errorCode = "";
                    this.errorPassword = "";

                    PagesController.ShowSnackBar(this.$t("Two factor authentication enabled"));
                    this.$emit("done");
                    this.forceCloseSignal++;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;

                    this.error = "";
                    this.errorCode = "";
                    this.errorPassword = "";

                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidCode: () => {
                            this.errorCode = this.$t("Invalid one-time code");
                        },
                        invalidPassword: () => {
                            this.errorPassword = this.$t("Invalid password");
                        },
                        tfaAlreadyEnabled: () => {
                            PagesController.ShowSnackBar(this.$t("Two factor authentication is already enabled"));
                            this.$emit("done");
                            this.forceCloseSignal++;
                        },
                        invalidSecretOrMethod: () => {
                            this.error = this.$t("Invalid two factor authentication settings. Try with another configuration.");
                            this.errorInvalidSettings = true;
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
