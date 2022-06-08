<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <div class="modal-dialog" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Vault settings") }}</div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body with-menu">
        <div class="modal-menu">
          <div
            class="modal-menu-item"
            tabindex="0"
            @click="clickOnOption('theme')"
          >
            <i class="modal-menu-item-icon fas fa-moon"></i>
            {{ $t("Change theme (Dark / Light)") }}
          </div>

          <div
            class="modal-menu-item"
            tabindex="0"
            @click="clickOnOption('lang')"
          >
            <i class="modal-menu-item-icon fas fa-language"></i>
            {{ $t("Change language") }}
          </div>

          <div
            class="modal-menu-item"
            tabindex="0"
            @click="clickOnOption('username')"
          >
            <i class="modal-menu-item-icon fas fa-user"></i>
            {{ $t("Change username") }}
          </div>

          <div
            class="modal-menu-item"
            tabindex="0"
            @click="clickOnOption('password')"
          >
            <i class="modal-menu-item-icon fas fa-key"></i>
            {{ $t("Change password") }}
          </div>

          <div
            class="modal-menu-item"
            tabindex="0"
            @click="clickOnOption('advanced')"
          >
            <i class="modal-menu-item-icon fas fa-cog"></i>
            {{ $t("Advanced settings") }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "SettingsModal",
  emits: ["update:display", "goto"],
  props: {
    display: Boolean,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    clickOnOption: function (o: string) {
      this.$emit("goto", o);
      this.close();
    },
  },
});
</script>

<style>
.modal-container-settings {
  z-index: 250;
}

.modal-menu {
  width: 100%;
  display: flex;
  flex-direction: column;
}

.modal-body.with-menu {
  padding: 0;
}

.modal-menu-item {
  padding: 1rem;
  cursor: pointer;
  font-weight: bold;
}

.modal-menu-item-icon {
  width: 2rem;
}

.modal-menu-item:hover {
  background: rgba(0, 0, 0, 0.1);
}
</style>
