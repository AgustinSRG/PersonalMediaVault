<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Save changes") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <p>{{ $t("Do you want to save the changes you made?") }}</p>
            </div>
            <div class="modal-footer text-right">
                <button type="button" class="btn btn-primary btn-mr" @click="clickNo"><i class="fas fa-times"></i> {{ $t("No") }}</button>
                <button type="button" class="btn btn-primary auto-focus" @click="clickYes">
                    <i class="fas fa-check"></i> {{ $t("Yes") }}
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

// Events
const emit = defineEmits<{
    /**
     * The user chose yes
     */
    (e: "yes"): void;

    /**
     * The user chose no
     */
    (e: "no"): void;
}>();

/**
 * Call when the user selects 'Yes'
 */
const clickYes = () => {
    emit("yes");
    close();
};

/**
 * Call when the user selects 'No'
 */
const clickNo = () => {
    emit("no");
    close();
};
</script>
