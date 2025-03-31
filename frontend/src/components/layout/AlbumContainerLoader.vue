<template>
    <div class="album-container">
        <LoadingOverlay></LoadingOverlay>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import LoadingOverlay from "./LoadingOverlay.vue";

export default defineComponent({
    name: "AlbumContainerLoader",
    components: {
        LoadingOverlay,
    },
    props: {},
});
</script>
