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
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Select a theme for the app") }}</div>
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
            @click="changeTheme('dark')"
            @keydown="clickOnEnter"
          >
            <td class="modal-menu-item-icon">
              <i
                class="fas fa-check"
                :class="{ unchecked: theme !== 'dark' }"
              ></i>
            </td>
            <td class="modal-menu-item-title">{{ $t("Dark Theme") }}</td>
          </tr>
          <tr
            class="modal-menu-item"
            tabindex="0"
            @click="changeTheme('light')"
            @keydown="clickOnEnter"
          >
            <td class="modal-menu-item-icon">
              <i
                class="fas fa-check"
                :class="{ unchecked: theme !== 'light' }"
              ></i>
            </td>
            <td class="modal-menu-item-title">{{ $t("Light Theme") }}</td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "ThemeModal",
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
      theme: AppPreferences.Theme,
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    changeTheme: function (t: string) {
      AppPreferences.SetTheme(t);
    },

    themeUpdated: function () {
      this.theme = AppPreferences.Theme;
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },
  },
  mounted: function () {
    this.$options.themeHandler = this.themeUpdated.bind(this);
    AppEvents.AddEventListener("theme-changed", this.$options.themeHandler);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("theme-changed", this.$options.themeHandler);
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
</style>
