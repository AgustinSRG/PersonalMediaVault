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
      v-if="display"
      class="modal-dialog modal-lg"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title about-modal-title">
          <img class="about-modal-logo" src="@/assets/favicon.png" alt="PMV" />
          {{ $t("Personal Media Vault") }}
        </div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body no-padding table-responsive">
        <table class="table table-text-overflow">
          <tr>
            <td>{{ $t("Version") }}</td>
            <td>{{ version }}</td>
          </tr>
          <tr>
            <td>{{ $t("Version date") }}</td>
            <td>{{ versionDate }}</td>
          </tr>

          <tr>
            <td>{{ $t("Home page") }}</td>
            <td>
              <a :href="homePage" target="_blank" rel="noopener noreferrer">{{
                homePage
              }}</a>
            </td>
          </tr>

          <tr>
            <td>{{ $t("Git repository") }}</td>
            <td>
              <a :href="gitRepo" target="_blank" rel="noopener noreferrer">{{
                gitRepo
              }}</a>
            </td>
          </tr>

          <tr>
            <td>{{ $t("License") }}</td>
            <td>
              <a :href="license" target="_blank" rel="noopener noreferrer">{{
                license
              }}</a>
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "AboutModal",
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
      version: process.env.VUE_APP_VERSION || "-",
      versionDate: process.env.VUE_APP_VERSION_DATE || "-",
      homePage: process.env.VUE_APP_HOME_URL || "#",
      gitRepo: process.env.VUE_APP_GIT_URL || "#",
      license: process.env.VUE_APP_LICENSE_URL || "#",
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
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
