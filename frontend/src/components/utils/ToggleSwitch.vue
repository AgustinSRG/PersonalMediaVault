<template>
  <div class="switch-button-container" tabindex="0" :disabled="disabled" @keydown="keyToggle">
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
  name: "ToggleSwitch",
  emits: ["update:val"],
  props: {
    val: Boolean,
    disabled: Boolean,
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
      if (this.disabled) {
        return;
      }
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

.switch-button-container:disabled {
  opacity: 0.7;
}

.light-theme .switch-button-container  {
  --switch-color: black;
  --switch-shadow-color: rgba(255, 255, 255, 0.33);
  --switch-disabled-color: lightgray;
}

.dark-theme .switch-button-container  {
  --switch-color: white;
  --switch-shadow-color: rgba(0, 0, 0, 0.33);
  --switch-disabled-color: gray;
}

.switch-button-control {
  display: flex;
  flex-direction: row;
  align-items: center;
}

.switch-button-control .switch-button {
  height: 16px;
  width: 48px;
  box-shadow: inset 0px 0px 2px 0px var(--switch-shadow-color);
  border-radius: 16px;
  transition: all 0.1s ease-in-out;
  cursor: pointer;
  background: var(--switch-disabled-color);
}

.switch-button-control .switch-button .button {
  height: 26px;
  width: 26px;
  border-radius: 26px;
  background: var(--switch-color);
  transition: all 0.1s ease-in-out;
  transform: translate(0, -6px);
}

.switch-button-control .switch-button.enabled {
  background-color: red;
  box-shadow: none;
}

.switch-button-control .switch-button.enabled .button {
  background: var(--switch-color);
  transform: translate(24px, -6px);
}
</style>