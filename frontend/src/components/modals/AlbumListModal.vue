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
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body with-menu">
        <table class="modal-menu">
          <tr
            class="modal-menu-item"
            tabindex="0"
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
            class="modal-menu-item"
            tabindex="0"
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
            @click="clickOnOption('password')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-key"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("Change password") }}
            </td>
          </tr>

          <tr
            class="modal-menu-item"
            tabindex="0"
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
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "AlbumListModal",
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

</style>
