<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <form @submit="submit" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Delete account") }}
        </div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Do you want to delete this account?") }}</label>
        </div>
        
        <div class="form-group">
          <label>{{name}}</label>
        </div>
      </div>
      <div class="modal-footer">
        <button type="submit" class="modal-footer-btn">
          <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "AccountDeleteModal",
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
  },
  mounted: function () {
    this.$options.showH = this.onShow.bind(this);
    AppEvents.AddEventListener(
      "account-del-confirmation",
      this.$options.showH
    );
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "account-del-confirmation",
      this.$options.showH
    );
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