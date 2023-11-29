<template>
    <ModalDialogContainer :closeSignal="closeSignal" v-model:display="displayStatus">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete attachment") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to delete this attachment file?") }}</label>
                </div>

                <div class="form-group">
                    <label>{{ attachmentToDelete ? attachmentToDelete.name : "" }}</label>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn auto-focus"><i class="fas fa-trash-alt"></i> {{ $t("Delete") }}</button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { PropType, defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { MediaAttachment } from "@/api/models";

export default defineComponent({
    name: "AttachmentDeleteModal",
    emits: ["update:display", "confirm"],
    props: {
        attachmentToDelete: Object as PropType<MediaAttachment>,
        display: Boolean,
    },
    data: function () {
        return {
            name: "",

            closeSignal: 0,
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
            this.closeSignal++;
        },

        submit: function (e) {
            e.preventDefault();

            this.$emit("confirm");

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
