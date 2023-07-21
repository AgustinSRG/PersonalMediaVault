<template>
  <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
    <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Change position") }}
        </div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Position in the album") }}:</label>
          <input type="text" name="album-position" autocomplete="off" v-model.number="currentPos" maxlength="255" class="form-control form-control-full-width auto-focus" />
        </div>
      </div>
      <div class="modal-footer no-padding">
        <button type="submit" class="modal-footer-btn">
          <i class="fas fa-arrows-up-down-left-right"></i>
          {{ $t("Change position") }}
        </button>
      </div>
    </form>
  </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

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
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        },

        show: function (options: { pos: number; callback: () => void }) {
            this.currentPos = options.pos + 1;
            this.callback = options.callback;
            this.displayStatus = true;
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        submit: function (e) {
            e.preventDefault();

            if (this.callback) {
                this.callback(this.currentPos - 1);
            }

            this.close();
        },
    },
    mounted: function () {
        if (this.display) {
            this.autoFocus();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
            }
        },
    },
});
</script>
