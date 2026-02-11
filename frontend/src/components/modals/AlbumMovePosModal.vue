<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Change position") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Position in the album") }}:</label>
                    <input
                        v-model.number="currentPos"
                        type="number"
                        name="album-position"
                        autocomplete="off"
                        step="1"
                        min="1"
                        :max="albumListLength"
                        class="form-control form-control-full-width auto-focus auto-select"
                    />
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn">
                    <i class="fas fa-arrows-up-down-left-right"></i>
                    {{ $t("Change position") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { ref, useTemplateRef, watch } from "vue";
import { AlbumsController } from "@/control/albums";
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
     * Original position to be moved
     */
    positionToMove: {
        type: Number,
        required: true,
    },

    /**
     * Length of the album
     */
    albumListLength: {
        type: Number,
        required: true,
    },
});

// Position to be moved to
const currentPos = ref(props.positionToMove + 1);

watch(
    () => props.positionToMove,
    () => {
        currentPos.value = props.positionToMove + 1;
    },
);

watch(display, () => {
    if (display.value) {
        currentPos.value = props.positionToMove + 1;
    }
});

/**
 * Handler for the 'submit' event
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    let newPos = currentPos.value - 1;

    if (isNaN(newPos) || !isFinite(newPos)) {
        close();
        return;
    }

    newPos = Math.floor(newPos);
    newPos = Math.min(newPos, props.albumListLength - 1);
    newPos = Math.max(0, newPos);

    if (newPos === props.positionToMove) {
        close();
        return;
    }

    AlbumsController.MoveCurrentAlbumOrder(props.positionToMove, newPos, $t);

    close();
};
</script>
