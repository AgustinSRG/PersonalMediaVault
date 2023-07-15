<template>
  <div class="modal-container modal-container-logout" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @mousedown="close" @touchstart="close" @keydown="keyDownHandle">
    <div v-if="display" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Save changes") }}</div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <p>{{ $t("Do you want to save the changes you made?") }}</p>
      </div>
      <div class="modal-footer text-right">
        <button type="button" class="btn btn-primary btn-mr" @click="clickNo">
          <i class="fas fa-times"></i> {{ $t("No") }}
        </button>
        <button type="button" class="btn btn-primary auto-focus" @click="clickYes">
          <i class="fas fa-check"></i> {{ $t("Yes") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "SaveChangesAskModal",
  emits: ["update:display", "yes", "no"],
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

    clickNo: function () {
      this.$emit("no");
      this.close();
    },

    clickYes: function () {
      this.$emit("yes");
      this.close();
    },

    autoFocus: function () {
      if (!this.display) {
        return;
      }
      nextTick(() => {
        const elem = this.$el.querySelector(".auto-focus");
        if (elem) {
          elem.focus();
        }
      });
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

    if (this.display) {
      this.$options.focusTrap.activate();
      this.autoFocus();
    }
  },
  beforeUnmount: function () {
    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        this.autoFocus();
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },
  },
});
</script>
