<template>
  <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
    <div v-if="display" class="modal-dialog modal-md" role="document">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Save changes") }}</div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <p>{{ $t("Do you want to save the changes you made?") }}</p>
      </div>
      <div class="modal-footer text-right">
        <button type="button" class="btn btn-primary btn-mr" @click="clickNo">
          <i class="fas fa-times"></i> {{ $t("No") }}
        </button>
        <button type="button" class="btn btn-primary auto-focus" @click="clickYes">
          <i class="fas fa-check"></i> {{ $t("Yes") }}
        </button>
      </div>
    </div>
  </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "SaveChangesAskModal",
    emits: ["update:display", "yes", "no"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    methods: {
        close: function () {
            this.$refs.modalContainer.close();
        },

        clickNo: function () {
            this.$emit("no");
            this.close();
        },

        clickYes: function () {
            this.$emit("yes");
            this.close();
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
