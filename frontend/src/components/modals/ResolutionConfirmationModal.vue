<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div v-if="deleting" class="modal-title">
                    {{ $t("Delete extra resolution") }}
                </div>
                <div v-if="!deleting" class="modal-title">
                    {{ $t("Encode to extra resolution") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div v-if="!deleting" class="form-group">
                    <label>{{ $t("Do you want to encode the media to this resolution? It will take more space in your vault.") }}</label>
                </div>

                <div v-if="deleting" class="form-group">
                    <label>{{ $t("Do you want to delete this extra resolution?") }}</label>
                </div>

                <div v-if="resolution" class="form-group">
                    <label v-if="type === MEDIA_TYPE_IMAGE">{{ resolution.name }}: {{ resolution.width }}x{{ resolution.height }}</label>
                    <label v-if="type === MEDIA_TYPE_VIDEO">
                        {{ resolution.name }}: {{ resolution.width }}x{{ resolution.height }}, {{ resolution.fps }} fps
                    </label>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button v-if="!deleting" type="submit" class="modal-footer-btn auto-focus">
                    <i class="fas fa-plus"></i> {{ $t("Encode") }}
                </button>
                <button v-if="deleting" type="submit" class="modal-footer-btn auto-focus">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { useTemplateRef } from "vue";
import type { PropType } from "vue";
import { MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO, type MediaType, type NamedResolution } from "@/api/models";
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

// Props
defineProps({
    /**
     * The resolution
     */
    resolution: Object as PropType<NamedResolution>,

    /**
     * Media type
     */
    type: Number as PropType<MediaType>,

    /**
     * True if deleting
     */
    deleting: Boolean,
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
