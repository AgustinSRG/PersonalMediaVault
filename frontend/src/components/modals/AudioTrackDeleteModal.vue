<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete audio track") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to delete this audio track file?") }}</label>
                </div>

                <div class="form-group">
                    <label>{{ trackToDelete ? trackToDelete.name : "" }}</label>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn auto-focus"><i class="fas fa-trash-alt"></i> {{ $t("Delete") }}</button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { useTemplateRef } from "vue";
import type { PropType } from "vue";
import type { MediaAudioTrack } from "@/api/models";
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
    trackToDelete: Object as PropType<MediaAudioTrack>,
});

// Events
const emit = defineEmits<{
    /**
     * Emitted on confirmation
     */
    (e: "confirm"): void;
}>();

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
