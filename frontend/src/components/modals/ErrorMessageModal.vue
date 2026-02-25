<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Error") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <p>{{ message }}</p>
            </div>
            <div class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="close">
                    <i class="fas fa-times"></i> {{ $t("Close") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { useTemplateRef } from "vue";
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
     * The error message
     */
    message: String,
});
</script>
