<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
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

                <table class="table no-margin no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="confirmation"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the attachment by accident you would have to re-upload it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" :disabled="!confirmation" class="modal-footer-btn auto-focus">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import type { MediaAttachment } from "@/api/models";
import ToggleSwitch from "../utils/ToggleSwitch.vue";

export default defineComponent({
    name: "AttachmentDeleteModal",
    components: {
        ToggleSwitch,
    },
    props: {
        attachmentToDelete: Object as PropType<MediaAttachment>,
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
            name: "",

            confirmation: false,

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

        submit: function (e: Event) {
            e.preventDefault();

            this.$emit("confirm");

            this.close();
        },
    },
});
</script>
