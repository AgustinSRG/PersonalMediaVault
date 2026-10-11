<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            v-model:display="display"
            :title="$t('Edit image note')"
            :position-key="'image-notes-editor-helper-pos'"
            :action-buttons="[
                {
                    id: 'delete',
                    name: $t('Delete'),
                    icon: 'fas fa-trash-alt',
                },
                {
                    id: 'save',
                    name: $t('Save changes'),
                    icon: 'fas fa-check',
                },
            ]"
            :initial-width="480"
            :initial-height="240"
            :initial-position="'bottom-right'"
            @action-btn="onActionBtn"
        >
            <div v-if="display" class="image-notes-text-edit-body">
                <textarea
                    ref="autoFocusElement"
                    v-model="text"
                    class="form-control form-textarea form-control-full-width auto-focus"
                    :placeholder="$t('Type the notes text') + '...'"
                    @change="emit('save')"
                ></textarea>
            </div>
        </ResizableWidget>
    </div>
</template>

<script setup lang="ts">
import ResizableWidget from "@/components/widgets/common/ResizableWidget.vue";
import { useI18n } from "@/composables/use-i18n";
import { onMounted, nextTick, useTemplateRef, watch } from "vue";

// Emits
const emit = defineEmits<{
    /**
     * Save the image note
     */
    (e: "save"): void;

    /**
     * Delete the image note
     */
    (e: "delete"): void;
}>();

// Translation
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Text model
const text = defineModel<string>("text");

// Element to be auto-focused on load
const autoFocusElement = useTemplateRef("autoFocusElement");

/**
 * Automatically focuses the appropriate element on load
 */
const autoFocus = () => {
    nextTick(() => {
        autoFocusElement.value?.focus();
    });
};

onMounted(() => {
    if (display.value) {
        autoFocus();
    }
});

watch(display, () => {
    if (display.value) {
        autoFocus();
    }
});

/**
 * Action button handler
 * @param action
 */
const onActionBtn = (action: string) => {
    switch (action) {
        case "save":
            emit("save");
            break;
        case "delete":
            emit("delete");
            break;
    }
};
</script>
