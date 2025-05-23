<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
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
                    <input
                        v-model.number="currentPos"
                        type="number"
                        name="album-position"
                        autocomplete="off"
                        step="1"
                        min="1"
                        class="form-control form-control-full-width auto-focus"
                    />
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
import { AlbumsController } from "@/control/albums";

export default defineComponent({
    name: "AlbumMovePosModal",
    props: {
        display: Boolean,
        positionToMove: Number,
        albumListLength: Number,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            currentPos: 0,
            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.currentPos = this.positionToMove + 1;
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
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e: Event) {
            e.preventDefault();

            let newPos = this.currentPos - 1;

            if (isNaN(newPos) || !isFinite(newPos)) {
                this.close();
                return;
            }
            newPos = Math.floor(newPos);
            newPos = Math.min(newPos, this.albumListLength - 1);
            newPos = Math.max(0, newPos);

            if (newPos === this.positionToMove) {
                this.close();
                return;
            }

            AlbumsController.MoveCurrentAlbumOrder(this.positionToMove, newPos, this.$t);

            this.close();
        },
    },
});
</script>
