<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Rename row") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Row name") }}:</label>
                    <input
                        v-model="name"
                        type="text"
                        name="row-name"
                        autocomplete="off"
                        :disabled="busy"
                        :placeholder="getDefaultGroupName(selectedRowType, $t)"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus auto-select"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Rename row") }}
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
import { apiHomeGroupRename } from "@/api/api-home";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
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
     * Emitted when the row is renamed
     */
    (e: "renamed", id: number, newName: string): void;

    /**
     * Emitted to indicate the home page should be reloaded
     */
    (e: "must-reload"): void;
}>();

// New name
const name = ref(props.selectedRowName);

/**
 * Resets the form
 */
const reset = () => {
    name.value = props.selectedRowName;
};

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Resets the error messages
const resetErrors = () => {
    error.value = "";
};

// Reset when modal opens
watch(display, () => {
    if (display.value) {
        reset();
        resetErrors();
    }
});

/**
 * Submits the form
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    if (busy.value) {
        return;
    }

    resetErrors();

    if (name.value === props.selectedRowName) {
        forceClose();
        return;
    }

    busy.value = true;

    const rowId = props.selectedRow;
    const newName = name.value;

    makeApiRequest(apiHomeGroupRename(rowId, newName))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Row renamed") + ": " + (newName || getDefaultGroupName(props.selectedRowType, $t)));

            name.value = "";

            forceClose();

            emit("renamed", rowId, newName);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidName: () => {
                    error.value = $t("Invalid album name provided");
                },
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
