<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Move row") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label
                        >{{ $t("Row name") }}: <b>{{ selectedRowName || getDefaultGroupName(selectedRowType, $t) }}</b></label
                    >
                </div>
                <div class="form-group">
                    <label>{{ $t("Position in the home page") }}:</label>
                    <input
                        v-model.number="currentPos"
                        type="number"
                        name="home-page-position"
                        autocomplete="off"
                        step="1"
                        min="1"
                        :max="maxPosition + 1"
                        class="form-control form-control-full-width auto-focus auto-select"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-arrows-up-down-left-right" :loading="busy"></LoadingIcon> {{ $t("Move row") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { getDefaultGroupName } from "@/utils/home";
import { apiHomeGroupMove } from "@/api/api-home";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
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

    /**
     * Position of the selected row
     */
    selectedRowPosition: {
        type: Number,
        required: true,
    },

    /**
     * Max position
     */
    maxPosition: {
        type: Number,
        required: true,
    },
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when the row is moved
     */
    (e: "moved", id: number, position: number): void;

    /**
     * Emitted to indicate the home page should be reloaded
     */
    (e: "must-reload"): void;
}>();

// Current position
const currentPos = ref(props.selectedRowPosition + 1);

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
        currentPos.value = props.selectedRowPosition + 1;
        resetErrors();
    }
});

/**
 * Submits the form
 * @param e The submit event
 */
const submit = (e: Event) => {
    e.preventDefault();

    if (busy.value) {
        return;
    }

    resetErrors();

    const position = currentPos.value - 1;

    if (position === props.selectedRowPosition) {
        forceClose();
        return;
    }

    busy.value = true;

    const rowId = props.selectedRow;

    makeApiRequest(apiHomeGroupMove(rowId, position))
        .onSuccess(() => {
            PagesController.ShowSnackBar(
                $t("Row moved") + ": " + (props.selectedRowName || getDefaultGroupName(props.selectedRowType, $t)),
            );

            busy.value = false;

            forceClose();

            emit("moved", rowId, position);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                accessDenied,
                notFound: () => {
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
</script>
