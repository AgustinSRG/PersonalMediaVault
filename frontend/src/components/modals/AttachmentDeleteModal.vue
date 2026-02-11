<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete attachment") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to delete this attachment file?") }}</label>
                </div>

                <div class="form-group">
                    <label>{{ attachmentToDelete?.name || "" }}</label>
                </div>

                <table class="table no-margin no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="confirmation"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the attachment by accident you would have to re-upload it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" :disabled="!confirmation" class="modal-footer-btn auto-focus">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { ref, useTemplateRef } from "vue";
import type { MediaAttachment } from "@/api/models";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

defineProps({
    /**
     * Attachment to be deleted
     */
    attachmentToDelete: Object as PropType<MediaAttachment>,
});

// Events
const emit = defineEmits<{
    /**
     * Emitted on confirmation
     */
    (e: "confirm"): void;
}>();

// Confirmation flag
const confirmation = ref(false);

/**
 * Handler for the 'submit' event
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    emit("confirm");

    close();
};
</script>
