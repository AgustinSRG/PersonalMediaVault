<template>
  <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
    <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Search") }}
        </div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Search by tag or album name") }}:</label>
          <input type="text" name="pmv-search" v-model="search" :disabled="busy" maxlength="255" class="form-control form-control-full-width auto-focus" />
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer no-padding">
        <button :disabled="busy" type="submit" class="modal-footer-btn">
          <i class="fas fa-search"></i> {{ $t("Search") }}
        </button>
      </div>
    </form>
  </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { AppStatus } from "@/control/app-status";

export default defineComponent({
  name: "SearchInputModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      search: AppStatus.CurrentSearch,

      busy: false,
      error: "",
    };
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
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
      this.$refs.modalContainer.close();
    },

    submit: function (e) {
      e.preventDefault();

      AppEvents.Emit("search-modal-submit", this.search);
      this.close();
    },
  },
  mounted: function () {
    if (this.display) {
      this.search = AppStatus.CurrentSearch;
      this.autoFocus();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        this.search = AppStatus.CurrentSearch;
        this.autoFocus();
      }
    },
  },
});
</script>