<template>
  <div
    class="modal-container modal-container-logout"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <div class="modal-dialog modal-md" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Close vault") }}</div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <p>{{ $t("Do you want to close your session and lock the vault?") }}</p>
      </div>
      <div class="modal-footer">
        <button type="button" class="modal-footer-btn" @click="logout">
          <i class="fas fa-sign-out-alt"></i> {{ $t("Close vault") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AuthController } from "@/control/auth";
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "LogoutModal",
  emits: ["update:display"],
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

    stopPropagationEvent: function(e) {
      e.stopPropagation()
    },

    logout: function () {
      AuthController.Logout();
      this.close();
    },
  },
});
</script>

<style>
.modal-container-logout {
  z-index: 250;
}
</style>
