<template>
    <div class="thumb-bottom-right-tag" :title="renderTitle(type, duration)" :class="{ small: !!small }">
        <i v-if="type === 3" class="fas fa-headphones"></i>
        {{ renderTimeSeconds(duration) }}
    </div>
</template>

<script setup lang="ts">
import { MEDIA_TYPE_AUDIO } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { renderTimeSeconds } from "@/utils/time";

const { $t } = useI18n();

defineProps({
    /**
     * Duration (seconds)
     */
    duration: {
        type: Number,
        required: true,
    },

    /**
     * Type of media
     */
    type: {
        type: Number,
        required: true,
    },

    /**
     * True to use small style
     */
    small: Boolean,
});

/**
 * Renders the title property of the duration tag
 * @param type Media type
 * @param duration The duration (seconds)
 */
const renderTitle = (type: number, duration: number): string => {
    const typeTagName = type === MEDIA_TYPE_AUDIO ? $t("Audio") : $t("Video");
    return typeTagName + " (" + renderTimeSeconds(duration) + ")";
};
</script>
