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

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
    name: "LoadingOverlay",
    props: {
        display: {
            type: Boolean,
            default: true,
        },
        fixed: Boolean,
        issues: {
            type: Boolean,
            default: false,
        },
    },
});
</script>
