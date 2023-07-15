<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display">
    <div v-if="display" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title" v-if="status === 'search'">
          {{ $t("Searching") }}...
        </div>
        <div class="modal-title" v-if="status === 'action'">
          {{ $t("Applying batch action") }}...
        </div>
        <div class="modal-title" v-if="status === 'confirmation' || status === 'confirmation-delete'">
          {{ $t("Confirmation") }}
        </div>
        <div class="modal-title" v-if="status === 'error'">
          {{ $t("Error") }}
        </div>
        <div class="modal-title" v-if="status === 'success'">
          {{ $t("Success") }}
        </div>
        <button v-if="status === 'search' || status === 'action'" type="button" class="modal-close-btn" :title="$t('Close')" @click="cancel">
          <i class="fas fa-times"></i>
        </button>
        <button v-if="status === 'confirmation' || status === 'confirmation-delete' || status === 'error' || status === 'success'" type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body" v-if="status === 'search' || status === 'action'">
        <div class="batch-progress-bar">
          <div class="batch-progress-bar-current" :style="{ width: cssProgress(progress) }"></div>
          <div class="batch-progress-bar-text">{{ renderStatus(status, progress) }}</div>
        </div>
      </div>
      <div class="modal-body" v-if="status === 'confirmation'">
        <div class="form-group">
          <label>{{ $t("Do you want to update $N elements?").replace("$N", "" + actionCount) }}</label>
        </div>
      </div>
      <div class="modal-body" v-if="status === 'confirmation-delete'">
        <div class="form-group">
          <label>{{ $t("Do you want to delete $N elements?").replace("$N", "" + actionCount) }}</label>
        </div>

        <div class="form-group">
          <label>{{ $t("Type 'confirm' for confirmation") }}:</label>
          <input type="text" name="confirmation" autocomplete="off" v-model="confirmationDelete" maxlength="255" class="form-control form-control-full-width auto-focus" />
        </div>
      </div>
      <div class="modal-body" v-if="status === 'error'">
        <div class="form-group">
          <label>{{ error }}</label>
        </div>
      </div>
      <div class="modal-body" v-if="status === 'success'">
        <div class="form-group">
          <label>{{ $t("The batch operation was completed successfully.") }}</label>
        </div>
      </div>
      <div class="modal-footer no-padding" v-if="status === 'confirmation'">
        <button type="button" @click="confirm" class="modal-footer-btn auto-focus">
          <i class="fas fa-check"></i> {{ $t("Continue") }}
        </button>
      </div>
      <div class="modal-footer no-padding" v-if="status === 'confirmation-delete'">
        <button type="button" @click="confirm" :disabled="confirmationDelete.toLowerCase() !== 'confirm'" class="modal-footer-btn auto-focus">
          <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
        </button>
      </div>
      <div class="modal-footer no-padding" v-if="status === 'error'">
        <button type="button" @click="close" class="modal-footer-btn auto-focus">
          <i class="fas fa-times"></i> {{ $t("Close") }}
        </button>
      </div>
      <div class="modal-footer no-padding" v-if="status === 'success'">
        <button type="button" @click="close" class="modal-footer-btn auto-focus">
          <i class="fas fa-check"></i> {{ $t("Close") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
  name: "BatchOperationProgressModal",
  emits: ["update:display", "confirm", "cancel"],
  props: {
    display: Boolean,

    status: String,
    progress: Number,

    actionCount: Number,

    error: String,
  },
  data: function () {
    return {
      confirmationDelete: "",
    };
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

    cancel: function () {
      this.$emit("cancel");
    },

    confirm: function () {
      this.$emit("confirm");
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    cssProgress: function (p: number) {
      return Math.round(p) + "%";
    },

    renderStatus(status: string, p: number) {
      const renderP = Math.round(p * 100) / 100;
      switch (status) {
        case "search":
          if (p > 0) {
            return this.$t("Searching") + "... (" + renderP + "%)";
          } else {
            return this.$t("Searching") + "...";
          }
        case "action":
          if (p > 0) {
            return this.$t("Applying") + "... (" + renderP + "%)";
          } else {
            return this.$t("Applying") + "...";
          }
        default:
          return "-";
      }
    },
  },
  mounted: function () {

  },
  beforeUnmount: function () {

  },
});
</script>
