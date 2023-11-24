<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
        <div v-if="display" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Page configuration") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Max items per page") }}:</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width auto-focus"
                        v-model.number="pageSize"
                        @change="onChangePageSize"
                        :max="maxPageSize"
                        :min="minPageSize"
                        step="1"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("Size of the thumbnails") }}:</label>
                    <select class="form-control form-select form-control-full-width" v-model="pageItemsSize" @change="onChangeItemsSize">
                        <option :value="'small'">{{ $t("Small thumbnails") }}</option>
                        <option :value="'normal'">{{ $t("Normal thumbnails") }}</option>
                        <option :value="'big'">{{ $t("Big thumbnails") }}</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>{{ $t("If possible, fit a number of elements in a single row") }}:</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width"
                        v-model.number="pageItemsFit"
                        @change="onChangePageItemsFit"
                        :max="256"
                        :min="1"
                        step="1"
                    />
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn" @click="close"><i class="fas fa-check"></i> {{ $t("Done") }}</button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { AppStatus } from "@/control/app-status";
import {
    getPageItemsFit,
    getPageItemsSize,
    getPageMaxItems,
    setPageItemsFit,
    setPageItemsSize,
    setPageMaxItems,
} from "@/control/app-preferences";

const MIN_PAGE_SIZE = 1;
const MAX_PAGE_SIZE = 256;

export default defineComponent({
    name: "PageConfigModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            page: AppStatus.CurrentPage,
            pageSize: getPageMaxItems(),

            maxPageSize: MAX_PAGE_SIZE,
            minPageSize: MIN_PAGE_SIZE,

            pageItemsFit: getPageItemsFit(),
            pageItemsSize: getPageItemsSize(),
        };
    },
    methods: {
        onChangePageSize() {
            const pageSize = parseInt(this.pageSize as any);
            const truePageSize = Math.min(256, Math.max(1, pageSize || 25));

            if (getPageMaxItems() !== truePageSize) {
                setPageMaxItems(truePageSize);
            }
        },

        onChangePageItemsFit: function () {
            const itemsFit = parseInt(this.pageItemsFit as any);
            const trueItemsFit = Math.min(256, Math.max(0, itemsFit || 5));

            if (getPageItemsFit() !== trueItemsFit) {
                setPageItemsFit(trueItemsFit);
            }
        },

        onChangeItemsSize: function () {
            setPageItemsSize(this.pageItemsSize);
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        reset: function () {
            this.page = AppStatus.CurrentPage;
            this.pageSize = getPageMaxItems();

            this.pageItemsFit = getPageItemsFit();
            this.pageItemsSize = getPageItemsSize();
        },

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                } else {
                    this.$el.focus();
                }
            });
        },
    },
    mounted: function () {
        this.reset();

        if (this.display) {
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
