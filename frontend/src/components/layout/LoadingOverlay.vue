<template>
    <div class="loading-overlay" :class="{ 'loading-overlay-fixed': fixed, hidden: !display }">
        <div class="loading-overlay-loader">
            <div></div>
            <div></div>
            <div></div>
            <div></div>
        </div>
        <div v-if="issues" class="load-issues">
            {{ $t("There are problems connecting to the server.") }} <br />
            {{ $t("Check your Internet connection or the server status.") }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";

const { $t } = useI18n();

defineProps({
    /**
     * True to display the loading overlay
     */
    display: {
        type: Boolean,
        default: true,
    },

    /**
     * True if the position is fixed
     */
    fixed: Boolean,

    /**
     * True to display the connections issues message.
     */
    issues: {
        type: Boolean,
        default: false,
    },
});
</script>
