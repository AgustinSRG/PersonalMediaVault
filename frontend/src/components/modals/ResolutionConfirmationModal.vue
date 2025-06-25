<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div v-if="deleting" class="modal-title">
                    {{ $t("Delete extra resolution") }}
                </div>
                <div v-if="!deleting" class="modal-title">
                    {{ $t("Encode to extra resolution") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div v-if="!deleting" class="form-group">
                    <label>{{ $t("Do you want to encode the media to this resolution? It will take more space in your vault.") }}</label>
                </div>

                <div v-if="deleting" class="form-group">
                    <label>{{ $t("Do you want to delete this extra resolution?") }}</label>
                </div>

                <div v-if="resolution" class="form-group">
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
import type { PropType } from "vue";
import type { NamedResolution } from "@/api/models";

export default defineComponent({
    name: "ResolutionConfirmationModal",
    props: {
        display: Boolean,
        resolution: Object as PropType<NamedResolution>,
        type: Number,
        deleting: Boolean,
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

        submit: function (e: Event) {
            e.preventDefault();

            this.$emit("confirm");

            this.close();
        },
    },
});
</script>
