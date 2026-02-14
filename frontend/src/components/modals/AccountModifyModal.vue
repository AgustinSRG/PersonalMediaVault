<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
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

                    <div v-if="accountUsernameError" class="form-error form-error-pt">{{ accountUsernameError }}</div>
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
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="performRequest"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { apiAdminUpdateAccount } from "@/api/api-admin";
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

// Props
const props = defineProps({
    /**
     * Username
     */
    username: {
        type: String,
        required: true,
    },

    /**
     * Write permission of the user
     */
    write: {
        type: Boolean,
        required: true,
    },
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when the request is done successfully.
     */
    (e: "done"): void;
}>();

// Username
const accountUsername = ref(props.username);

watch(
    () => props.username,
    () => {
        accountUsername.value = props.username;
    },
);

// Write permission
const accountWrite = ref(props.write);

watch(
    () => props.write,
    () => {
        accountWrite.value = props.write;
    },
);

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

// Resets the error messages
const resetErrors = () => {
    error.value = "";

    accountUsernameError.value = "";
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

    busy.value = true;

    makeApiRequest(apiAdminUpdateAccount(props.username, accountUsername.value, accountWrite.value, confirmation))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Account updated") + ": " + accountUsername.value);

            emit("done");

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
                badRequest,
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied,
                accountNotFound: () => {
                    error.value = $t("Not found");
                },
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
</script>
