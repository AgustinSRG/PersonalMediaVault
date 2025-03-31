<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Close invited session") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to close the invited session?") }}</label>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn auto-focus"><i class="fas fa-trash-alt"></i> {{ $t("Close") }}</button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "InviteCloseSessionConfirmationModal",
    props: {
        display: Boolean,
    },
    emits: ["update:display", "confirm"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.autoFocus();
        }
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
            this.closeSignal++;
        },

        submit: function (e) {
            e.preventDefault();

            this.$emit("confirm");

            this.close();
        },
    },
});
</script>
