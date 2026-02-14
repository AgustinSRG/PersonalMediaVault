<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete row") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <p>
                    <b>{{ selectedRowName || getDefaultGroupName(selectedRowType, $t) }}</b>
                </p>
                <table class="table no-margin no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="deleteConfirmation"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the row by accident you would have to recreate it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy || !deleteConfirmation" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Delete row") }}
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
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { getDefaultGroupName } from "@/utils/home";
import { apiHomeGroupDelete } from "@/api/api-home";
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
     * Id of the selected row
     */
    selectedRow: {
        type: Number,
        required: true,
    },

    /**
     * Type of the selected row
     */
    selectedRowType: {
        type: Number,
        required: true,
    },

    /**
     * Name of the selected row
     */
    selectedRowName: {
        type: String,
        required: true,
    },
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when the row is deleted
     */
    (e: "row-deleted", id: number): void;

    /**
     * Emitted to indicate the home page should be reloaded
     */
    (e: "must-reload"): void;
}>();

// Delete confirmation
const deleteConfirmation = ref(false);

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
        deleteConfirmation.value = false;
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

    const rowId = props.selectedRow;
    const oldName = props.selectedRowName || getDefaultGroupName(props.selectedRowType, $t);

    makeApiRequest(apiHomeGroupDelete(rowId, confirmation))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Row deleted") + ": " + oldName);

            deleteConfirmation.value = false;

            forceClose();

            emit("row-deleted", rowId);
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
                notFound: () => {
                    error.value = $t("Not found");

                    forceClose();

                    emit("must-reload");
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
