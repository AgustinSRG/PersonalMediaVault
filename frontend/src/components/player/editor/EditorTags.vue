<template>
    <div class="player-editor-sub-content">
        <MediaTagsEditor @tags-update="onTagUpdate"></MediaTagsEditor>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import MediaTagsEditor from "@/components/utils/MediaTagsEditor.vue";

export default defineComponent({
    name: "EditorTags",
    components: {
        MediaTagsEditor,
    },
    emits: ["changed"],
    methods: {
        onTagUpdate: function () {
            this.$emit("changed");
        },
    },
});
</script>
