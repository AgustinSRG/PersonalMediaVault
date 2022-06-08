<template>
  <MainLayout></MainLayout>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import MainLayout from "./components/layout/MainLayout.vue";

@Options({
  components: {
    MainLayout,
  },
  data: function () {
    return {};
  },
})
export default class App extends Vue {}
</script>

<style>

/* Add here any global / shared styles */ 

</style>
