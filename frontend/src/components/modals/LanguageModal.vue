<template>
  <div
    class="modal-container modal-container-corner no-transition"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
    @keydown="keyDownHandle"
  >
    <div
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header-corner">
        <div class="modal-header-corner-title">{{ $t("Choose your language") }}</div>
      </div>
      <div class="modal-body with-menu limited-height">
        <table class="modal-menu">
          <tr
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="changeLocale('en')"
          >
            <td class="modal-menu-item-icon">
              <i class="fas fa-check" :class="{ unchecked: lang !== 'en' }"></i>
            </td>
            <td class="modal-menu-item-title">English ({{ $t("Default") }})</td>
          </tr>
          <tr
            class="modal-menu-item"
            tabindex="0"
            @keydown="clickOnEnter"
            @click="changeLocale('es')"
          >
            <td class="modal-menu-item-icon">
              <i class="fas fa-check" :class="{ unchecked: lang !== 'es' }"></i>
            </td>
            <td class="modal-menu-item-title">Español (Internacional)</td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AppPreferences } from "@/control/app-preferences";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";
import { AppEvents } from "@/control/app-events";

export default defineComponent({
  name: "LanguageModal",
  emits: ["update:display"],
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
      lang: AppPreferences.Language,
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    changeLocale: function (l: string) {
      this.lang = l;
      AppPreferences.SetLanguage(l);
      AppEvents.Emit("set-locale", l);
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
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "top-bar-button-dropdown");
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
.modal-body.limited-height {
  max-height: 300px;
  overflow-y: auto;
}

.modal-menu-item-icon .unchecked {
  visibility: hidden;
}
</style>
