<template>
  <div class="snackbar" :class="{ hidden: !shown }">{{ message }}</div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { defineComponent } from "vue";

export default defineComponent({
  name: "SnackBar",
  data: function () {
    return {
      shown: false,
      message: "",
    };
  },
  methods: {
    show: function (msg: string) {
      if (this.$options.timeout) {
        clearTimeout(this.$options.timeout);
        this.$options.timeout = null;
      }

      this.shown = true;
      this.message = msg;

      this.$options.timeout = setTimeout(() => {
        this.shown = false;
      }, 3000);
    },
  },
  mounted: function () {
    this.$options.showH = this.show.bind(this);
    AppEvents.AddEventListener("snack", this.$options.showH);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("snack", this.$options.showH);
    if (this.$options.timeout) {
      clearTimeout(this.$options.timeout);
      this.$options.timeout = null;
    }
  },
});
</script>

<style>
.snackbar {
  position: fixed;
  left: 1rem;
  bottom: 1rem;
  opacity: 1;
  background: #212121;
  box-shadow: 0 16px 24px 2px rgb(0 0 0 / 14%), 0 6px 30px 5px rgb(0 0 0 / 12%),
    0 8px 10px -5px rgb(0 0 0 / 40%);
  padding: 1rem;
  font-weight: bold;
  border-radius: 0.25rem;
  transition: bottom 0.3s, opacity 0.1s;
  z-index: 900;
}

.snackbar.hidden {
  bottom: -5rem;
  opacity: 0;
  pointer-events: none;
}
</style>
