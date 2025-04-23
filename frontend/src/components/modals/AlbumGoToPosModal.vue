<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Go to position") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="pos-input-container">
                    <div class="form-control-container">
                        <input
                            v-model.number="currentPos"
                            type="number"
                            name="album-position"
                            autocomplete="off"
                            step="1"
                            min="1"
                            :max="albumLength"
                            class="form-control form-control-full-width auto-focus"
                        />
                    </div>
                    <div v-if="albumLength > 0" class="form-control-suffix">
                        {{ "/ " + albumLength }}
                    </div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn">
                    <i class="fas fa-forward-step"></i>
                    {{ $t("Go") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { AlbumsController } from "@/control/albums";
import { AppStatus } from "@/control/app-status";

export default defineComponent({
    name: "AlbumGoToPosModal",
    props: {
        display: Boolean,
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
            albumLength: 0,

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.reset();
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.reset();
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

        reset: function () {
            this.currentPos = AlbumsController.CurrentAlbumPos + 1;
            this.albumLength = AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.list.length : 0;
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e: Event) {
            e.preventDefault();

            if (AlbumsController.CurrentAlbumData && AlbumsController.CurrentAlbumData.list.length > 0) {
                const pos = Math.min(Math.max(0, Math.floor(this.currentPos - 1)), AlbumsController.CurrentAlbumData.list.length - 1);

                AppStatus.ClickOnMedia(AlbumsController.CurrentAlbumData.list[pos].id, false);
            }

            this.close();
        },
    },
});
</script>
