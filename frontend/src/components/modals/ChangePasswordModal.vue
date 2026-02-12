<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form ref="form" class="modal-dialog modal-md" role="document" @submit="submit">
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

                    <div v-if="currentPasswordError" class="form-error form-error-pt">{{ currentPasswordError }}</div>
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

                    <div v-if="passwordError" class="form-error form-error-pt">{{ passwordError }}</div>
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

                    <div v-if="password2Error" class="form-error form-error-pt">{{ password2Error }}</div>
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
            @confirm="performRequest"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { apiAccountChangePassword } from "@/api/api-account";
import { EVENT_NAME_AUTH_CHANGED } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { AuthController } from "@/control/auth";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "@/components/utils/PasswordInput.vue";
import PasswordStrengthIndicator from "@/components/utils/PasswordStrengthIndicator.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose } = useModal(display, container);

// Username
const username = ref(AuthController.Username);

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    username.value = AuthController.Username;
});

// Current password
const currentPassword = ref("");

// New password
const password = ref("");

// New password (confirmation)
const password2 = ref("");

/**
 * Resets the form
 */
const resetForm = () => {
    currentPassword.value = "";
    password.value = "";
    password2.value = "";
};

// Auth confirmation
const {
    displayAuthConfirmation,
    authConfirmationCooldown,
    authConfirmationTfa,
    authConfirmationError,
    requiredAuthConfirmationTfa,
    invalidTfaCode,
    cooldown,
} = useAuthConfirmation();

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, serverError, networkError } = useCommonRequestErrors();

// Other errors
const currentPasswordError = ref("");
const passwordError = ref("");
const password2Error = ref("");

// Resets the error messages
const resetErrors = () => {
    error.value = "";

    currentPasswordError.value = "";
    passwordError.value = "";
    password2Error.value = "";
};

watch(display, () => {
    if (display.value) {
        resetErrors();
        resetForm();
    }
});

/**
 * Performs the request
 * @param confirmation The auth confirmation
 */
const performRequest = (confirmation: ProvidedAuthConfirmation) => {
    if (busy.value) {
        return;
    }

    resetErrors();

    if (!password.value) {
        passwordError.value = $t("Invalid password provided");
        return;
    }

    if (password.value !== password2.value) {
        password2Error.value = $t("The passwords do not match");
        return;
    }

    busy.value = true;

    makeApiRequest(apiAccountChangePassword(currentPassword.value, password.value, confirmation))
        .onSuccess(() => {
            busy.value = false;

            resetForm();

            PagesController.ShowSnackBar($t("Vault password changed!"));

            forceClose();
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidNewPassword: () => {
                    passwordError.value = $t("Invalid password provided");
                },
                invalidPassword: () => {
                    currentPasswordError.value = $t("Invalid password");
                },
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
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

    performRequest({});
};

// Form container
const form = useTemplateRef("form");

/**
 * Handler for tab focus skip on password input 1
 * @param e The keyboard event
 */
const passwordTabSkip1 = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector('[name="new-password"]') as HTMLInputElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();

        if (typeof nextElement.select === "function") {
            nextElement.select();
        }
    }
};

/**
 * Handler for tab focus skip on password input 2
 * @param e The keyboard event
 */
const passwordTabSkip2 = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector('[name="new-password-repeat"]') as HTMLInputElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();

        if (typeof nextElement.select === "function") {
            nextElement.select();
        }
    }
};

/**
 * Handler for tab focus skip on password input 3
 * @param e The keyboard event
 */
const passwordTabSkip3 = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector(".modal-footer-btn") as HTMLElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();
    }
};
</script>
