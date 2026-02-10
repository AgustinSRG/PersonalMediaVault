<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form ref="form" class="modal-dialog modal-lg" role="document" @submit="submit">
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
                    <input type="text" readonly name="totp-url" :value="url" class="form-control form-control-full-width" />
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
                    <input type="text" class="hidden-input" name="username" :value="originalAccount" />
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
            @done="loadSettings"
        ></AccountTfaSettingsModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, onMounted, ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "../utils/PasswordInput.vue";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import type { TimeOtpAlgorithm, TimeOtpPeriod } from "@/api/api-account";
import { apiAccountTimeOtpEnable, apiAccountTimeOtpSettings } from "@/api/api-account";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useRequestId } from "@/composables/use-request-id";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

const AccountTfaSettingsModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountTfaSettingsModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose, focus } = useModal(display, container);

// Events
const emit = defineEmits<{
    /**
     * Emitted when the request is done successfully.
     */
    (e: "done"): void;
}>();

// TFA TOTP issuer name
const issuer = ref("PMV");

// TFA account name
const account = ref(AuthController.Username);

// TFA original account name
const originalAccount = ref(AuthController.Username);

// TOTP algorithm
const algorithm = ref<TimeOtpAlgorithm>("sha1");

// TOTP period
const period = ref<TimeOtpPeriod>("30");

// Allow clock skew?
const skew = ref(true);

// TOTP secret
const secret = ref("");

// TFA method
const method = ref("");

// TFA URL
const url = ref("");

// QR code
const qr = ref("");

// Loading settings status
const loadingSettings = ref(true);

// Load request ID
const loadRequestId = useRequestId();

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the TFA settings
 */
const loadSettings = () => {
    loadingSettings.value = true;

    resetErrors();

    clearNamedTimeout(loadRequestId);

    makeNamedApiRequest(
        loadRequestId,
        apiAccountTimeOtpSettings({
            issuer: issuer.value,
            account: account.value,
            algorithm: algorithm.value,
            period: period.value,
            skew: skew.value ? "allow" : "disallow",
        }),
    )
        .onSuccess((response) => {
            secret.value = response.secret;
            method.value = response.method;
            url.value = response.url;
            qr.value = response.qr;

            code.value = "";

            loadingSettings.value = false;
            errorInvalidSettings.value = false;

            focus();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                invalidSettings: () => {
                    loadingSettings.value = false;
                    error.value = $t("Invalid two factor authentication settings. Try with another configuration.");
                    errorInvalidSettings.value = true;
                },
                accessDenied: () => {
                    close();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSettings);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSettings);
        });
};

onMounted(() => {
    if (display.value) {
        loadSettings();
    }
});

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    account.value = AuthController.Username;
    originalAccount.value = AuthController.Username;
});

// Display custom settings modal
const displayCustomSettings = ref(false);

watch(display, () => {
    if (display.value) {
        displayCustomSettings.value = false;

        loadSettings();
    }
});

/**
 * Opens the modal to customize the settings
 */
const customizeSettings = () => {
    displayCustomSettings.value = true;
};

// Account password (confirmation)
const password = ref("");

// TFA code (confirmation)
const code = ref("");

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Other errors
const errorInvalidSettings = ref(false);
const errorPassword = ref("");
const errorCode = ref("");

// Resets the error messages
const resetErrors = () => {
    error.value = "";

    errorInvalidSettings.value = false;
    errorPassword.value = "";
    errorCode.value = "";
};

/**
 * Performs the request
 */
const performRequest = () => {
    if (busy.value) {
        return;
    }

    resetErrors();

    busy.value = true;

    makeApiRequest(
        apiAccountTimeOtpEnable({
            secret: secret.value,
            method: method.value,
            password: password.value,
            code: code.value,
        }),
    )
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Two factor authentication enabled"));

            emit("done");

            forceClose();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidCode: () => {
                    errorCode.value = $t("Invalid one-time code");
                },
                invalidPassword: () => {
                    errorPassword.value = $t("Invalid password");
                },
                tfaAlreadyEnabled: () => {
                    PagesController.ShowSnackBar($t("Two factor authentication is already enabled"));

                    emit("done");

                    forceClose();
                },
                invalidSecretOrMethod: () => {
                    error.value = $t("Invalid two factor authentication settings. Try with another configuration.");
                    errorInvalidSettings.value = true;
                },
                accessDenied,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            error.value = err.message;

            console.error(err);
        });
};

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    performRequest();
};

/**
 * Copies TFA URL to clipboard
 */
const copyUrl = () => {
    navigator.clipboard.writeText(url.value);
    PagesController.ShowSnackBar($t("Copied URL to clipboard"));
};

// Form container
const form = useTemplateRef("form");

/**
 * Handler for tab focus skip on password input
 * @param e The keyboard event
 */
const passwordTabSkip = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector(".form-control.code-char-0") as HTMLElement;

    if (nextElement) {
        e.preventDefault();

        nextElement.focus();
    }
};
</script>
