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
                    <select class="form-control form-select form-control-full-width" v-model="pageSize" @change="onChangePageSize">
                        <option v-for="po in pageSizeOptions" :key="po" :value="po">{{ po }} {{ $t("items per page") }}</option>
                        <option :value="0">{{ $t("Use a custom value") }}</option>
                    </select>
                </div>
                <div class="form-group" v-if="pageSize === 0">
                    <input
                        type="number"
                        class="form-control form-control-full-width"
                        v-model.number="pageSizeCustom"
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
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";

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
            pageSize: AppPreferences.PageMaxItems,
            pageSizeCustom: AppPreferences.PageMaxItems,

            maxPageSize: MAX_PAGE_SIZE,
            minPageSize: MIN_PAGE_SIZE,

            pageSizeOptions: [],

            pageItemsFit: AppPreferences.PageItemsFit,
            pageItemsSize: AppPreferences.PageItemsSize,
        };
    },
    methods: {
        onChangePageSize() {
            const pageSize = parseInt(this.pageSize === 0 ? this.pageSizeCustom : this.pageSize);
            const truePageSize = Math.min(256, Math.max(1, pageSize || 25));

            if (AppPreferences.PageMaxItems !== truePageSize) {
                AppPreferences.SetPageMaxItems(truePageSize);
            }
        },

        onChangePageItemsFit: function () {
            const itemsFit = parseInt(this.pageItemsFit);
            const trueItemsFit = Math.min(256, Math.max(0, itemsFit || 5));

            if (AppPreferences.PageItemsFit !== trueItemsFit) {
                AppPreferences.SetPageItemsFit(trueItemsFit);
            }
        },

        onChangeItemsSize: function () {
            AppPreferences.SetPageItemsSize(this.pageItemsSize);
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        reset: function () {
            this.page = AppStatus.CurrentPage;
            this.pageSize = AppPreferences.PageMaxItems;

            if (this.pageSizeOptions.includes(this.pageSize)) {
                this.pageSizeCustom = this.pageSize;
            } else {
                this.pageSizeCustom = this.pageSize;
                this.pageSize = 0;
            }

            this.pageItemsFit = AppPreferences.PageItemsFit;
            this.pageItemsSize = AppPreferences.PageItemsSize;
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
        this.pageSizeOption = [];

        for (let i = 1; i <= 20; i++) {
            this.pageSizeOptions.push(5 * i);
        }

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
