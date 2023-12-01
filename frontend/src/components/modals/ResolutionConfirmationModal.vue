<template>
    <ModalDialogContainer :closeSignal="closeSignal" v-model:display="displayStatus">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title" v-if="deleting">
                    {{ $t("Delete extra resolution") }}
                </div>
                <div class="modal-title" v-if="!deleting">
                    {{ $t("Encode to extra resolution") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group" v-if="!deleting">
                    <label>{{ $t("Do you want to encode the media to this resolution? It will take more space in your vault.") }}</label>
                </div>

                <div class="form-group" v-if="deleting">
                    <label>{{ $t("Do you want to delete this extra resolution?") }}</label>
                </div>

                <div class="form-group" v-if="resolution">
                    <label v-if="type === 1">{{ resolution.name }}: {{ resolution.width }}x{{ resolution.height }}</label>
                    <label v-if="type === 2">
                        {{ resolution.name }}: {{ resolution.width }}x{{ resolution.height }}, {{ resolution.fps }} fps
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
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { PropType } from "vue";
import { NamedResolution } from "@/api/models";

export default defineComponent({
    name: "ResolutionConfirmationModal",
    emits: ["update:display", "confirm"],
    props: {
        display: Boolean,
        resolution: Object as PropType<NamedResolution>,
        type: Number,
        deleting: Boolean,
    },
    data: function () {
        return {
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
