<template>
  <div class="player-container">
    <LoadingOverlay></LoadingOverlay>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import LoadingOverlay from "./LoadingOverlay.vue";

export default defineComponent({
  components: {
    LoadingOverlay,
  },
  name: "PlayerContainerLoader",
  props: {
  },
});
</script>
