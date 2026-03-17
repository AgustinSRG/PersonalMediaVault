<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Go to position") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="pos-input-container">
                    <div class="form-control-container">
                        <input
                            v-model.number="currentPos"
                            type="number"
                            name="album-position"
                            autocomplete="off"
                            step="1"
                            min="1"
                            :max="albumLength"
                            class="form-control form-control-full-width auto-focus"
                        />
                    </div>
                    <div v-if="albumLength > 0" class="form-control-suffix">
                        {{ "/ " + albumLength }}
                    </div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn">
                    <i class="fas fa-forward-step"></i>
                    {{ $t("Go") }}
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
import { getCurrentAlbumData, getCurrentAlbumMediaPosition } from "@/global-state/album";
import { navigationClickOnMedia } from "@/global-state/navigation";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Current album position
const currentPos = ref(getCurrentAlbumMediaPosition() + 1);

// Current album length
const albumLength = ref(getCurrentAlbumData()?.list.length || 0);

/**
 * Resets the form
 */
const reset = () => {
    currentPos.value = getCurrentAlbumMediaPosition() + 1;
    albumLength.value = getCurrentAlbumData()?.list.length || 0;
};

watch(display, () => {
    if (display.value) {
        reset();
    }
});

/**
 * Handler for the 'submit' event
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    const currentAlbumData = getCurrentAlbumData();

    if (currentAlbumData && currentAlbumData.list.length > 0) {
        const pos = Math.min(Math.max(0, Math.floor(currentPos.value - 1)), currentAlbumData.list.length - 1);

        navigationClickOnMedia(currentAlbumData.list[pos].id, false);
    }

    close();
};
</script>
