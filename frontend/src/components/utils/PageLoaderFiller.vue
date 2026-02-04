<template>
    <div class="search-results-loading-display">
        <div v-for="f in pageSize" :key="f" class="search-result-item">
            <div class="search-result-thumb">
                <div class="search-result-thumb-inner">
                    <div class="search-result-loader">
                        <i class="fa fa-spinner fa-spin"></i>
                    </div>
                </div>
            </div>
            <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";

const { $t } = useI18n();

defineProps({
    /**
     * Page size
     */
    pageSize: {
        type: Number,
        default: 0,
    },

    /**
     * True to display the titles
     */
    displayTitles: Boolean,
});
</script>
