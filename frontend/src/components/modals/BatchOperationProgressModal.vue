<template>
    <ModalDialogContainer ref="container" v-model:display="display" :static="true" :lock-close="status === 'search' || status === 'action'">
        <div class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div v-if="status === 'search'" class="modal-title">{{ $t("Searching") }}...</div>
                <div v-if="status === 'action'" class="modal-title">{{ $t("Applying batch action") }}...</div>
                <div v-if="status === 'confirmation' || status === 'confirmation-delete'" class="modal-title">
                    {{ $t("Confirmation") }}
                </div>
                <div v-if="status === 'error'" class="modal-title">
                    {{ $t("Error") }}
                </div>
                <div v-if="status === 'success'" class="modal-title">
                    {{ $t("Success") }}
                </div>
                <button
                    v-if="status === 'search' || status === 'action'"
                    type="button"
                    class="modal-close-btn"
                    :title="$t('Close')"
                    @click="cancel"
                >
                    <i class="fas fa-times"></i>
                </button>
                <button
                    v-if="status === 'confirmation' || status === 'confirmation-delete' || status === 'error' || status === 'success'"
                    type="button"
                    class="modal-close-btn"
                    :title="$t('Close')"
                    @click="close"
                >
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="status === 'search' || status === 'action'" class="modal-body">
                <div class="batch-progress-bar">
                    <div class="batch-progress-bar-current" :style="{ width: progressBarWidth }"></div>
                    <div class="batch-progress-bar-text">{{ statusDisplayString }}</div>
                </div>
            </div>
            <div v-if="status === 'confirmation'" class="modal-body">
                <div class="form-group">
                    <label>{{ stringMultiReplace($t("Do you want to update $N elements?"), { $N: "" + actionCount }) }}</label>
                </div>
            </div>
            <div v-if="status === 'confirmation-delete'" class="modal-body">
                <div class="form-group">
                    <label>{{ stringMultiReplace($t("Do you want to delete $N elements?"), { $N: "" + actionCount }) }}</label>
                </div>

                <table class="table no-margin no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="confirmationDelete"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the media by accident you would have to re-upload it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div v-if="status === 'error'" class="modal-body">
                <div class="form-group">
                    <label>{{ error }}</label>
                </div>
            </div>
            <div v-if="status === 'success'" class="modal-body">
                <div class="form-group">
                    <label>{{ $t("The batch operation was completed successfully.") }}</label>
                </div>
            </div>
            <div v-if="status === 'confirmation'" class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="confirm">
                    <i class="fas fa-check"></i> {{ $t("Continue") }}
                </button>
            </div>
            <div v-if="status === 'confirmation-delete'" class="modal-footer no-padding">
                <button type="button" :disabled="!confirmationDelete" class="modal-footer-btn auto-focus" @click="confirm">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                </button>
            </div>
            <div v-if="status === 'error'" class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="close">
                    <i class="fas fa-times"></i> {{ $t("Close") }}
                </button>
            </div>
            <div v-if="status === 'success'" class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="close">
                    <i class="fas fa-check"></i> {{ $t("Close") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import type { PropType } from "vue";
import { computed, ref, useTemplateRef } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { stringMultiReplace } from "@/utils/string-multi-replace";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose } = useModal(display, container);

// Statuses of the batch operation
type BatchOperationStatus = "" | "search" | "confirmation" | "confirmation-delete" | "action" | "success" | "error";

// Props
const props = defineProps({
    /**
     * Status of the operation
     */
    status: {
        type: String as PropType<BatchOperationStatus>,
        required: true,
    },

    /**
     * Progress (%)
     */
    progress: {
        type: Number,
        default: 0,
    },

    /**
     * Number of elements to apply the action
     */
    actionCount: Number,

    /**
     * Error message
     */
    error: String,
});

// Progress bar width
const progressBarWidth = computed(() => Math.round(props.progress) + "%");

// Events
const emit = defineEmits<{
    /**
     * Confirmation event
     */
    (e: "confirm"): void;

    /**
     * Cancellation event
     */
    (e: "cancel"): void;
}>();

/**
 * Confirms the operation
 */
const confirm = () => {
    emit("confirm");
};

/**
 * Cancels the operation
 */
const cancel = () => {
    emit("cancel");
    forceClose();
};

// Deletion confirmation
const confirmationDelete = ref(false);

// String to display the progress to the user
const statusDisplayString = computed(() => {
    const p = props.progress;
    const renderP = Math.round(p * 100) / 100;
    switch (props.status) {
        case "search":
            if (p > 0) {
                return $t("Searching") + "... (" + renderP + "%)";
            } else {
                return $t("Searching") + "...";
            }
        case "action":
            if (p > 0) {
                return $t("Applying") + "... (" + renderP + "%)";
            } else {
                return $t("Applying") + "...";
            }
        default:
            return "-";
    }
});
</script>
