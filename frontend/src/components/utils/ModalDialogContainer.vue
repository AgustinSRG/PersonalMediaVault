<template>
  <div class="modal-container modal-container-dialog" :class="{ hidden: !display, closing: closing }" tabindex="-1" role="dialog" :aria-hidden="!display" @keydown="keyDownHandle" @animationend="onAnimationEnd" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent" @click="stopPropagationEvent">
    <div class="modal-out-close-area" @mousedown="onMouseDown" @touchstart="onTouchOutside"></div>
    <slot @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent" @click="stopPropagationEvent"></slot>
  </div>
</template>

<script lang="ts">
import { FocusTrap } from "@/utils/focus-trap";
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

export default defineComponent({
  name: "ModalDialogContainer",
  emits: ["update:display", "close"],
  props: {
    display: Boolean,
    lockClose: Boolean,
    static: Boolean,
    closeCallback: Function,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      closing: false,
    };
  },
  methods: {
    close: function (forced: boolean) {
      if (this.lockClose && forced !== true) {
        return;
      }
      if (this.closeCallback && forced !== true) {
        this.closeCallback(() => {
          this.$emit("close");
          this.closing = true;
        });
      } else {
        this.$emit("close");
        this.closing = true;
      }
    },

    onAnimationEnd: function (e: AnimationEvent) {
      e.stopPropagation();
      if (e.animationName === "modal-close-animation") {
        this.closing = false;
        this.displayStatus = false;
      }
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape" && this.display && !this.closing) {
        this.close();
      }
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    onTouchOutside: function (e: TouchEvent) {
      e.stopPropagation();
      if (!this.static) {
        this.close();
      }
    },

    onMouseDown: function (e: MouseEvent) {
      e.stopPropagation();
      if (e.button === 0 && !this.static) {
        this.close();
      }
    },

    focusLost: function () {
      this.close();
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.focusLost.bind(this));

    if (this.display) {
      this.$options.focusTrap.activate();
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
        this.closing = false;
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },
  },
});
</script>
