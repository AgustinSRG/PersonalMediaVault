<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @mousedown="close" @touchstart="close" @keydown="keyDownHandle">
    <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Delete subtitles") }}
        </div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Do you want to delete this subtitles file?") }}</label>
        </div>

        <div class="form-group">
          <label>{{ name }}</label>
        </div>
      </div>
      <div class="modal-footer no-padding">
        <button type="submit" class="modal-footer-btn auto-focus">
          <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "SubtitlesDeleteModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      name: "",

      callback: null,
    };
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
    onShow: function (options) {
      this.name = options.name;
      this.callback = options.callback;
      this.displayStatus = true;
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

    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    submit: function (e) {
      e.preventDefault();

      if (this.callback) {
        this.callback();
      }

      this.close();
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.$options.showH = this.onShow.bind(this);
    AppEvents.AddEventListener("subtitles-confirmation", this.$options.showH);

    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

    if (this.display) {
      this.$options.focusTrap.activate();
      this.autoFocus();
    }
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "subtitles-confirmation",
      this.$options.showH
    );

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
