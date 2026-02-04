<template>
    <div class="no-thumb">
        <i v-if="type === MEDIA_TYPE_IMAGE" class="fas fa-image"></i>
        <i v-else-if="type === MEDIA_TYPE_VIDEO" class="fas fa-video"></i>
        <i v-else-if="type === MEDIA_TYPE_AUDIO" class="fas fa-headphones"></i>
        <i v-else class="fas fa-ban"></i>
    </div>
</template>

<script setup lang="ts">
import type { MediaType } from "@/api/models";
import { MEDIA_TYPE_AUDIO, MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO } from "@/api/models";
import type { PropType } from "vue";

defineProps({
    /**
     * Media type
     */
    type: {
        type: Number as PropType<MediaType>,
        required: true,
    },
});
</script>
