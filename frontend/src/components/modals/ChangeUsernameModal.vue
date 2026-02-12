<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form ref="form" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Change username") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Current username") }}:</label>
                    <input
                        v-model="currentUsername"
                        type="text"
                        name="current-username"
                        :disabled="busy"
                        maxlength="255"
                        readonly
                        class="form-control form-control-full-width"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("New username") }}:</label>
                    <input
                        v-model="username"
                        type="text"
                        name="username"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />

                    <div v-if="usernameError" class="form-error form-error-pt">{{ usernameError }}</div>
                </div>
                <div class="form-group">
                    <label>{{ $t("Password") }}:</label>
                    <PasswordInput v-model:val="password" :name="'password'" :disabled="busy" @tab-skip="passwordTabSkip"></PasswordInput>

                    <div v-if="passwordError" class="form-error form-error-pt">{{ passwordError }}</div>
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Change username") }}
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
import { apiAccountChangeUsername } from "@/api/api-account";
import { EVENT_NAME_AUTH_CHANGED } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import PasswordInput from "@/components/utils/PasswordInput.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { onApplicationEvent } from "@/composables/on-app-event";
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

// Current username
const currentUsername = ref(AuthController.Username);

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    currentUsername.value = AuthController.Username;
});

// New username
const username = ref("");

// Account password (for confirmation)
const password = ref("");

/**
 * Resets the form
 */
const resetForm = () => {
    username.value = "";
    password.value = "";
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
const usernameError = ref("");
const passwordError = ref("");

// Resets the error messages
const resetErrors = () => {
    error.value = "";

    usernameError.value = "";
    passwordError.value = "";
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

    busy.value = true;

    makeApiRequest(apiAccountChangeUsername(username.value, password.value, confirmation))
        .onSuccess(() => {
            busy.value = false;

            AuthController.UpdateUsername(username.value);

            resetForm();

            PagesController.ShowSnackBar($t("Vault username changed!"));

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
                    usernameError.value = $t("Invalid username provided");
                },
                usernameInUse: () => {
                    usernameError.value = $t("The username is already in use");
                },
                invalidPassword: () => {
                    passwordError.value = $t("Invalid password");
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
 * Handler for tab focus skip on password input
 * @param e The keyboard event
 */
const passwordTabSkip = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector(".modal-footer-btn") as HTMLElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();
    }
};
</script>
