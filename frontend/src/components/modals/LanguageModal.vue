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
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Choose your language") }}</div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
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
      this.$i18n.locale = l;
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
.modal-body.limited-height {
  max-height: 300px;
  overflow-y: auto;
}

.modal-menu-item-icon .unchecked {
  visibility: hidden;
}
</style>
