<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            v-model:display="displayStatus"
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

<script lang="ts">
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";
import ResizableWidget from "@/components/player/ResizableWidget.vue";
import MediaTagsEditor from "@/components/utils/MediaTagsEditor.vue";

export default defineComponent({
    name: "TagsEditHelper",
    components: {
        ResizableWidget,
        MediaTagsEditor,
    },
    props: {
        display: Boolean,
        contextOpen: Boolean,
    },
    emits: ["update:display", "tags-update", "clicked"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    methods: {
        propagateClick: function () {
            this.$emit("clicked");
        },

        onTagUpdate: function () {
            this.$emit("tags-update");
        },

        close: function () {
            this.displayStatus = false;
        },
    },
});
</script>
