<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
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
                            type="number"
                            name="album-position"
                            autocomplete="off"
                            v-model.number="currentPos"
                            step="1"
                            min="1"
                            :max="albumLength"
                            class="form-control form-control-full-width auto-focus"
                        />
                    </div>
                    <div class="form-control-suffix" v-if="albumLength > 0">
                        {{ '/ ' + albumLength }}
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
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            currentPos: 0,
            albumLength: 0,
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

        reset: function () {
            this.currentPos = AlbumsController.CurrentAlbumPos + 1;
            this.albumLength = AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.list.length : 0;
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        submit: function (e) {
            e.preventDefault();

            if (AlbumsController.CurrentAlbumData && AlbumsController.CurrentAlbumData.list.length > 0) {
                const pos = Math.min(Math.max(0, Math.floor(this.currentPos - 1)), AlbumsController.CurrentAlbumData.list.length - 1);

                AppStatus.ClickOnMedia(AlbumsController.CurrentAlbumData.list[pos].id, false);
            }

            this.close();
        },
    },
    mounted: function () {
        if (this.display) {
            this.reset();
            this.autoFocus();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.reset();
                this.autoFocus();
            }
        },
    },
});
</script>

<style scoped>
.pos-input-container {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
}

.form-control-container {
    flex: 1;
}
.form-control-suffix {
    padding-left: 0.75rem;
    white-space: nowrap;
}

</style>
