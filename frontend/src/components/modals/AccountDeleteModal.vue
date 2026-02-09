<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete account") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to delete this account?") }}</label>
                </div>

                <div class="form-group">
                    <label>{{ username }}</label>
                </div>

                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" :disabled="busy" class="modal-footer-btn auto-focus">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Delete") }}
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
import { ref, useTemplateRef, watch } from "vue";
import { apiAdminDeleteAccount } from "@/api/api-admin";
import { PagesController } from "@/control/pages";
import { makeApiRequest } from "@asanrom/request-browser";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
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
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when the request is done successfully.
     */
    (e: "done"): void;
}>();

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
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Resets the error messages
const resetErrors = () => {
    error.value = "";
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

    makeApiRequest(apiAdminDeleteAccount(props.username, confirmation))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Account deleted") + ": " + props.username);

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
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied,
                accountNotFound: () => {
                    // Already deleted?
                    busy.value = false;

                    emit("done");

                    forceClose();
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
