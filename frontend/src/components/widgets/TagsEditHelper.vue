<template>
    <div class="resizable-widget-container" :class="{ focused: focused }">
        <ResizableWidget
            v-model:display="display"
            v-model:focused="focused"
            :title="$t('Tags')"
            :context-open="contextOpen"
            :position-key="'tags-edit-helper-pos'"
            @clicked="propagateClick"
        >
            <div v-if="display" class="tags-editor-body">
                <MediaTagsEditor :allow-navigation="true" @tags-update="onTagUpdate"></MediaTagsEditor>
            </div>
        </ResizableWidget>
    </div>
</template>

<script setup lang="ts">
import ResizableWidget from "@/components/widgets/common/ResizableWidget.vue";
import MediaTagsEditor from "@/components/utils/MediaTagsEditor.vue";
import { useI18n } from "@/composables/use-i18n";
import { ref } from "vue";

// Translation
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Focused?
const focused = ref(false);

defineProps({
    /**
     * True if a context menu is opened
     */
    contextOpen: Boolean,
});

// Emits
const emit = defineEmits<{
    /**
     * The widget was clicked
     */
    (e: "clicked"): void;

    /**
     * The list of tags was updated
     */
    (e: "tags-update"): void;
}>();

/**
 * Propagates the click event to the parent element
 */
const propagateClick = () => {
    emit("clicked");
};

/**
 * Called when the tags list change
 */
const onTagUpdate = () => {
    emit("tags-update");
};
</script>
