<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
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

                <div class="form-group">
                    <label v-if="type === 1">{{ name }}: {{ width }}x{{ height }}</label>
                    <label v-if="type === 2"> {{ name }}: {{ width }}x{{ height }}, {{ fps }} fps </label>
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

export default defineComponent({
    name: "ResolutionConfirmationModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            deleting: false,
            name: "",
            type: 2,
            width: 0,
            height: 0,
            fps: 0,

            callback: null,
        };
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    methods: {
        show: function (options: {
            type: number;
            deleting: boolean;
            name: string;
            width: number;
            height: number;
            fps?: number;
            callback: () => void;
        }) {
            this.type = options.type;
            this.deleting = options.deleting;
            this.name = options.name;
            this.width = options.width;
            this.height = options.height;
            this.fps = options.fps;
            this.callback = options.callback;
            this.displayStatus = true;
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

        close: function () {
            this.$refs.modalContainer.close();
        },

        submit: function (e) {
            e.preventDefault();

            if (this.callback) {
                this.callback();
            }

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
