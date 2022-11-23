<template>
  <div class="snackbar" :class="{ hidden: !shown }" @mouseenter="hide" @click="hide">{{ message }}</div>
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

    hide: function () {
      if (this.$options.timeout) {
        clearTimeout(this.$options.timeout);
        this.$options.timeout = null;
      }
      this.shown = false;
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
