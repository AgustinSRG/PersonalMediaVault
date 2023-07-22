<template>
    <div class="loading-overlay" :class="{ 'loading-overlay-fixed': fixed, hidden: !display }">
        <div class="loading-overlay-loader">
            <div></div>
            <div></div>
            <div></div>
            <div></div>
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
    },
});
</script>
