<template>
    <div class="main-layout">
        <TopBar></TopBar>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import TopBar from "./TopBar.vue"

export default defineComponent({
    components: {
        TopBar,
    },
    name: "MainLayout",
})

</script>

<style>

*,
*::before,
*::after {
  box-sizing: border-box;
}

.main-layout {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: #181818;
    color: white;
}
</style>
