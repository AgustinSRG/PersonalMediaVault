<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form ref="form" class="modal-dialog modal-md" role="document" @submit="submit">
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

                    <div v-if="accountUsernameError" class="form-error form-error-pt">{{ accountUsernameError }}</div>
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

                    <div v-if="accountPasswordError" class="form-error form-error-pt">{{ accountPasswordError }}</div>
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

                    <div v-if="accountPassword2Error" class="form-error form-error-pt">{{ accountPassword2Error }}</div>
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

                <div v-if="error" class="form-group form-error">{{ error }}</div>
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
            @confirm="performRequest"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import { apiAdminCreateAccount } from "@/api/api-admin";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "@/components/utils/PasswordInput.vue";
import PasswordStrengthIndicator from "@/components/utils/PasswordStrengthIndicator.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
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

// Events
const emit = defineEmits<{
    /**
     * Emitted when the account is created and the modal closed.
     */
    (e: "account-created"): void;
}>();

// Username
const accountUsername = ref("");

// Password
const accountPassword = ref("");

// Password (repeat)
const accountPassword2 = ref("");

// Account write permission
const accountWrite = ref(false);

// Auth confirmation
const {
    displayAuthConfirmation,
    authConfirmationCooldown,
    authConfirmationTfa,
    authConfirmationError,
    requiredAuthConfirmationPassword,
    invalidPassword,
    requiredAuthConfirmationTfa,
    invalidTfaCode,
    cooldown,
} = useAuthConfirmation();

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, badRequest, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Other errors
const accountUsernameError = ref("");
const accountPasswordError = ref("");
const accountPassword2Error = ref("");

// Resets the error messages
const resetErrors = () => {
    error.value = "";

    accountUsernameError.value = "";
    accountPasswordError.value = "";
    accountPassword2Error.value = "";
};

// Reset error when modal opens
watch(display, () => {
    if (display.value) {
        resetErrors();
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

    const username = accountUsername.value;
    const password = accountPassword.value;
    const write = accountWrite.value;

    if (!username) {
        accountUsernameError.value = $t("Invalid username provided");
        return;
    }

    if (!password) {
        accountPasswordError.value = $t("Invalid password provided");
        return;
    }

    if (password !== accountPassword2.value) {
        accountPassword2Error.value = $t("The passwords do not match");
        return;
    }

    busy.value = true;

    makeApiRequest(apiAdminCreateAccount(username, password, write, confirmation))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Account created") + ": " + username);

            accountUsername.value = "";
            accountPassword.value = "";
            accountPassword2.value = "";
            accountWrite.value = false;

            emit("account-created");

            forceClose();
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidUsername: () => {
                    accountUsernameError.value = $t("Invalid username provided");
                },
                usernameInUse: () => {
                    accountUsernameError.value = $t("The username is already in use");
                },
                invalidUserPassword: () => {
                    accountPasswordError.value = $t("Invalid password provided");
                },
                badRequest,
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
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

    performRequest({});
};

// Form container
const form = useTemplateRef("form");

/**
 * Handler for tab focus skip on password input 1
 * @param e The keyboard event
 */
const passwordTabSkip1 = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector('[name="account-new-password-repeat"]') as HTMLInputElement;

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
    const nextElement = form.value?.querySelector('[name="account-type"]') as HTMLElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();
    }
};
</script>
