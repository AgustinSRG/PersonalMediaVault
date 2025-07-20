<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Move element") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Position in the row") }}:</label>
                    <input
                        v-model.number="currentPos"
                        type="number"
                        name="home-page-element-position"
                        autocomplete="off"
                        step="1"
                        min="1"
                        :max="maxPosition + 1"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn">
                    <i class="fas fa-arrows-up-down-left-right"></i> {{ $t("Move element") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { getDefaultGroupName } from "@/utils/home";

export default defineComponent({
    name: "HomePageMoveElementModal",
    props: {
        display: Boolean,

        selectedPosition: Number,

        maxPosition: Number,
    },
    emits: ["update:display", "move-element"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            currentPos: this.selectedPosition + 1,

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.currentPos = this.selectedPosition + 1;
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.currentPos = this.selectedPosition + 1;
            this.autoFocus();
        }
    },
    methods: {
        getDefaultGroupName: getDefaultGroupName,

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                elem.focus();
                elem.select();
            });
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e: Event) {
            e.preventDefault();

            const position = this.currentPos - 1;

            if (position === this.selectedPosition) {
                this.close();
                return;
            }

            this.$emit("move-element", this.selectedPosition, position);
            this.close();
        },
    },
});
</script>
