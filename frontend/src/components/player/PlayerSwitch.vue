<template>
  <div class="switch-button-container" tabindex="0" @keydown="keyToggle">
    <div class="switch-button-control">
      <div
        class="switch-button"
        :class="{ enabled: val }"
        @click="toggle"
      >
        <div class="button"></div>
      </div>
    </div>
    <div class="switch-button-label">
      <slot></slot>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "PlayerSwitch",
  emits: ["update:val"],
  props: {
    val: Boolean,
  },
  setup(props) {
    return {
      valState: useVModel(props, "val"),
    };
  },
  data: function () {
    return {
    };
  },
  methods: {
    toggle: function () {
      this.valState = !this.valState;
    },
    keyToggle: function (e) {
      if (e.key === " " || e.key === "Enter") {
        this.toggle();
      }
      e.stopPropagation();
    },
  },
  mounted: function () {},
  beforeUnmount: function () {},
});
</script>

<style>
.switch-button-container {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: right;
  width: 100%;
}

.switch-button-control {
  display: flex;
  flex-direction: row;
  align-items: center;
}

.switch-button-control .switch-button {
  height: 16px;
  width: 48px;
  box-shadow: inset 0px 0px 2px 0px rgba(0, 0, 0, 0.33);
  border-radius: 16px;
  transition: all 0.1s ease-in-out;
  cursor: pointer;
  background: gray;
}

.switch-button-control .switch-button .button {
  height: 26px;
  width: 26px;
  border-radius: 26px;
  background: white;
  transition: all 0.1s ease-in-out;
  transform: translate(0, -6px);
}

.switch-button-control .switch-button.enabled {
  background-color: red;
  box-shadow: none;
}

.switch-button-control .switch-button.enabled .button {
  background: white;
  transform: translate(24px, -6px);
}
</style>