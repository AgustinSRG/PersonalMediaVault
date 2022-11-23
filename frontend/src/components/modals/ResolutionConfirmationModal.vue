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
    <form
      v-if="display"
      @submit="submit"
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title" v-if="deleting">
          {{ $t("Delete extra resolution") }}
        </div>
        <div class="modal-title" v-if="!deleting">
          {{ $t("Encode to extra resolution") }}
        </div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group" v-if="!deleting">
          <label>{{
            $t(
              "Do you want to encode the media to this resolution? It will take more space in your vault."
            )
          }}</label>
        </div>

        <div class="form-group" v-if="deleting">
          <label>{{
            $t("Do you want to delete this extra resolution?")
          }}</label>
        </div>

        <div class="form-group">
          <label v-if="type === 1">{{ name }}: {{ width }}x{{ height }}</label>
          <label v-if="type === 2">
            {{ name }}: {{ width }}x{{ height }}, {{ fps }} fps
          </label>
        </div>
      </div>
      <div class="modal-footer no-padding">
        <button v-if="!deleting" type="submit" class="modal-footer-btn auto-focus">
          <i class="fas fa-plus"></i> {{ $t("Encode") }}
        </button>
        <button v-if="deleting" type="submit" class="modal-footer-btn auto-focus">
          <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { MediaController } from "@/control/media";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "ResolutionConfirmationModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      currentMedia: -1,

      deleting: false,
      name: "",
      type: 2,
      width: 0,
      height: 0,
      fps: 0,

      callback: null,
    };
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
    onMediaUpdate: function () {
      this.currentMedia = AppStatus.CurrentMedia;
      if (MediaController.MediaData) {
        this.oldName = MediaController.MediaData.title;
      }
    },

    onShow: function (options) {
      this.type = options.type;
      this.deleting = options.deleting;
      this.name = options.name;
      this.width = options.width;
      this.height = options.height;
      this.fps = options.fps;
      this.callback = options.callback;
      this.displayStatus = true;
    },

    autoFocus: function () {
      if (!this.display) {
        return;
      }
      nextTick(() => {
        const elem = this.$el.querySelector(".auto-focus");
        elem.focus();
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
    AppEvents.AddEventListener("resolution-confirmation", this.$options.showH);
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "resolution-confirmation",
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
