<template>
  <img alt="Vue logo" src="./assets/logo.png" />
  <div style="display: flex; background: black">
    <VolumeControl
      :width="60"
      :min="false"
      v-model:muted="muted"
      v-model:volume="volume"
    ></VolumeControl>
  </div>

  <HelloWorld msg="Welcome to Your Vue.js + TypeScript App" />
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";
import HelloWorld from "./components/HelloWorld.vue";
import VolumeControl from "./components/player/VolumeControl.vue";

@Options({
  components: {
    HelloWorld,
    VolumeControl,
  },
  data: function () {
    return {
      test: 1,
      volume: 0.5,
      muted: false,
    };
  },
  methods: {
    runTest: function () {
      this.test++;
    },
  },
})
export default class App extends Vue {}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
