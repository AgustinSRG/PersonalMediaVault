<template>
  <div class="page-inner" :class="{hidden: !display}">
    
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "PageRandom",
  props: {
    display: Boolean,
  },
  data: function () {
    return {};
  },
  methods: {},
  mounted: function () {},
  beforeUnmount: function () {},
});
</script>

<style>



</style>