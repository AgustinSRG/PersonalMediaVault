<template>
  <div class="loading-overlay" :class="{ hidden: !display }">
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
    display: Boolean,
  },
});
</script>

<style>
/* Loader */

.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;

  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;

  transition: opacity 0.2s;
  opacity: 1;

  z-index: 300;

  background: var(--theme-bg-color);
}

.loading-overlay.hidden {
  pointer-events: none;
  opacity: 0;
}

.loading-overlay-loader {
  display: inline-block;
  position: relative;
  width: 120px;
  height: 120px;
  --loader-color: var(--theme-fg-color);
}

.loading-overlay-loader div {
  box-sizing: border-box;
  display: block;
  position: absolute;
  width: 110px;
  height: 110px;
  margin: 8px;
  border: 8px solid var(--loader-color);
  border-radius: 50%;
  animation: loading-overlay-loader 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
  border-color: var(--loader-color) transparent transparent transparent;
}

@media (max-width: 400px) {
  .loading-overlay-loader {
    width: 48px;
    height: 48px;
  }
  .loading-overlay-loader div {
    width: 42px;
    height: 42px;
    margin: 4px;
    border: 4px solid var(--loader-color);
    border-color: var(--loader-color) transparent transparent transparent;
  }
}

.loading-overlay-loader div:nth-child(1) {
  animation-delay: -0.45s;
}

.loading-overlay-loader div:nth-child(2) {
  animation-delay: -0.3s;
}

.loading-overlay-loader div:nth-child(3) {
  animation-delay: -0.15s;
}

@keyframes loading-overlay-loader {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
