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
      @submit="submit"
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Change position") }}
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
        <div class="form-group">
          <label>{{ $t("Position in the album") }}:</label>
          <input
            type="text"
            name="album-position"
            autocomplete="off"
            v-model.number="currentPos"
            maxlength="255"
            class="form-control form-control-full-width auto-focus"
          />
        </div>
      </div>
      <div class="modal-footer">
        <button type="submit" class="btn btn-primary">
          <i class="fas fa-arrows-up-down-left-right"></i>
          {{ $t("Change position") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "AlbumMovePosModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      currentPos: 0,
      callback: null,
    };
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
    autoFocus: function () {
      const elem = this.$el.querySelector(".auto-focus");
      if (elem) {
        setTimeout(() => {
          elem.focus();
          elem.select();
        }, 200);
      }
    },

    onShow: function (options: { pos: number; callback: () => void }) {
      this.currentPos = options.pos + 1;
      this.callback = options.callback;
      this.displayStatus = true;
      this.autoFocus();
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
        this.callback(this.currentPos - 1);
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
    AppEvents.AddEventListener("album-user-request-pos", this.$options.showH);
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "album-user-request-pos",
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
</style>
