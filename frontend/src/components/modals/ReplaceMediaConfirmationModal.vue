<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Replace media") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to upload the file and replace the current media file?") }}</label>
                </div>
                <div class="form-group">
                    <label>{{ $t("File") }}: {{ fileName }}</label>
                </div>
                <div class="form-group">
                    <label>{{ $t("Size") }}: {{ renderSize(fileSize || 0) }}</label>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn auto-focus">
                    <i class="fas fa-upload"></i> {{ $t("Upload and replace") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { useTemplateRef } from "vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { renderSize } from "@/utils/size";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Props
defineProps({
    /**
     * Name of the file
     */
    fileName: String,

    /**
     * Size of the file
     */
    fileSize: Number,
});

// Events
const emit = defineEmits<{
    /**
     * Confirmation event
     */
    (e: "confirm"): void;
}>();

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();
    close();
    emit("confirm");
};
</script>
