<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <div
      class="modal-dialog modal-sm"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Vault settings") }}</div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body with-menu">
        <table class="modal-menu">
          <tr
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('theme')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-moon"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Change theme (Dark / Light)") }}
            </td>
          </tr>

          <tr
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('lang')"
          >
            <td class="modal-menu-item-icon">
              <i class="fas fa-language"></i>
            </td>
            <td class="modal-menu-item-title">
              {{ $t("Change language") }}
            </td>
          </tr>

          <tr
            v-if="isRoot"
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('username')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-user"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Change username") }}
            </td>
          </tr>

          <tr
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('password')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-key"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Change password") }}
            </td>
          </tr>

          <tr
            v-if="isRoot"
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('admin')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-users"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Administrate accounts") }}
            </td>
          </tr>

          <tr
            v-if="isRoot"
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('tasks')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-bars-progress"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Tasks") }}
            </td>
          </tr>

          <tr
            v-if="isRoot"
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('advanced')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-cog"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Advanced settings") }}
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { defineComponent, nextTick } from "vue";
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
  data: function () {
    return {
      isRoot: AuthController.IsRoot,
      canWrite: AuthController.CanWrite,
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

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },

    updateAuthInfo: function () {
      this.isRoot = AuthController.IsRoot;
      this.canWrite = AuthController.CanWrite;
    },
  },
  mounted: function () {
    this.$options.authUpdateH = this.updateAuthInfo.bind(this);

    AppEvents.AddEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );
  },
  watch: {
    display: function () {
      if (this.display) {
        nextTick(() => {
          this.$el.focus();
        });
      }
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
  border-spacing: 0; /* Removes the cell spacing via CSS */
  border-collapse: collapse; /* Optional - if you don't want to have double border where cells touch */
}

.modal-body.with-menu {
  padding: 0;
}

.modal-menu-item {
  cursor: pointer;
}

.modal-menu-item-title {
  padding-top: 1rem;
  padding-right: 1rem;
  padding-bottom: 1rem;
  font-weight: bold;
}

.modal-menu-item-icon {
  padding: 1rem;
  text-align: center;
  width: 2rem;
}

.light-theme .modal-menu-item:hover {
  background: rgba(0, 0, 0, 0.1);
}

.dark-theme .modal-menu-item:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
