<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
    @keydown="keyDownHandle"
  >
    <div
      class="modal-dialog modal-sm"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Help") }}</div>
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
            @click="clickOnOption('about')"
          >
            <td class="modal-menu-item-icon"><i class="fas fa-info"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("About PMV") }}
            </td>
          </tr>
          
          <tr
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="clickOnOption('keyboard')"
          >
            <td class="modal-menu-item-icon">
              <i class="fas fa-keyboard"></i>
            </td>
            <td class="modal-menu-item-title">
              {{ $t("Keyboard shortcuts") }}
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "HelpHubModal",
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
    return {};
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

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));
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
        nextTick(() => {
          this.$el.focus();
        });
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
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
