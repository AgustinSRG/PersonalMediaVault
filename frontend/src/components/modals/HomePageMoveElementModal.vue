<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Move element") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Position in the row") }}:</label>
                    <input
                        v-model.number="currentPos"
                        type="number"
                        name="home-page-element-position"
                        autocomplete="off"
                        step="1"
                        min="1"
                        :max="maxPosition + 1"
                        class="form-control form-control-full-width auto-focus auto-select"
                    />
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn">
                    <i class="fas fa-arrows-up-down-left-right"></i> {{ $t("Move element") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { ref, useTemplateRef, watch } from "vue";
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
const props = defineProps({
    /**
     * Selected position
     */
    selectedPosition: {
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
     * Emitted when the user chooses to move the element
     * to a new position
     */
    (e: "move-element", oldPos: number, newPos: number): void;
}>();

// Current position
const currentPos = ref(props.selectedPosition + 1);

/**
 * Resets the position
 */
const reset = () => {
    currentPos.value = props.selectedPosition;
};

watch(display, () => {
    if (display.value) {
        reset();
    }
});

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    const position = currentPos.value - 1;

    if (position === props.selectedPosition) {
        close();
        return;
    }

    emit("move-element", props.selectedPosition, position);

    close();
};
</script>
