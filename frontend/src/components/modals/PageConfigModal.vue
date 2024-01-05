<template>
    <ModalDialogContainer :closeSignal="closeSignal" v-model:display="displayStatus">
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
                        @input="markDirty"
                        :max="1"
                        :min="256"
                        step="1"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("Preferred row size") }}:</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width auto-focus"
                        v-model.number="rowSize"
                        @input="markDirty"
                        :max="1"
                        :min="256"
                        step="1"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Preferred row size") }} ({{ $t("For screen split mode") }}):</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width auto-focus"
                        v-model.number="rowSizeMin"
                        @input="markDirty"
                        :max="1"
                        :min="256"
                        step="1"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Min size of cells (pixels)") }}:</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width auto-focus"
                        v-model.number="minItemSize"
                        @input="markDirty"
                        :max="1"
                        :min="1000"
                        step="1"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Max size of cells (pixels)") }}:</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width auto-focus"
                        v-model.number="maxItemSize"
                        @input="markDirty"
                        :max="1"
                        :min="1000"
                        step="1"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Cell padding (pixels)") }}:</label>
                    <input
                        type="number"
                        class="form-control form-control-full-width auto-focus"
                        v-model.number="padding"
                        @input="markDirty"
                        :max="0"
                        :min="64"
                        step="1"
                    />
                </div>
                <div class="form-group table-responsive">
                    <table class="table no-border no-margin">
                        <tr>
                            <td class="text-right td-shrink">
                                <toggle-switch v-model:val="displayTitles" @update:val="markDirty"></toggle-switch>
                            </td>
                            <td>
                                {{ $t("Display titles?") }}
                            </td>
                        </tr>
                        <tr>
                            <td class="text-right td-shrink">
                                <toggle-switch v-model:val="roundedCorners" @update:val="markDirty"></toggle-switch>
                            </td>
                            <td>
                                {{ $t("Use rounded borders for corners?") }}
                            </td>
                        </tr>
                    </table>
                </div>
                <div>
                    <button type="button" class="btn btn-primary btn-sm" @click="resetDefaultValues">
                        <i class="fas fa-sync-alt"></i> {{ $t("Reset to default values") }}
                    </button>
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
import { getPagePreferences, resetPagePreferences, setPagePreferences } from "@/control/app-preferences";
import ToggleSwitch from "../utils/ToggleSwitch.vue";

export default defineComponent({
    name: "PageConfigModal",
    components: {
        ToggleSwitch,
    },
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            saveChangesTimer: null,
        };
    },
    data: function () {
        const pagePreferences = getPagePreferences();
        return {
            page: AppStatus.CurrentPage,

            pageSize: pagePreferences.pageSize,

            rowSize: pagePreferences.rowSize,
            rowSizeMin: pagePreferences.rowSizeMin,

            minItemSize: pagePreferences.minItemSize,
            maxItemSize: pagePreferences.maxItemSize,

            padding: pagePreferences.padding,

            displayTitles: pagePreferences.displayTitles,
            roundedCorners: pagePreferences.roundedCorners,

            closeSignal: 0,

            dirty: false,
        };
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        reset: function () {
            if (this.saveChangesTimer) {
                clearTimeout(this.saveChangesTimer);
                this.saveChangesTimer = null;
            }

            this.page = AppStatus.CurrentPage;

            const pagePreferences = getPagePreferences();

            this.pageSize = pagePreferences.pageSize;

            this.rowSize = pagePreferences.rowSize;
            this.rowSizeMin = pagePreferences.rowSizeMin;

            this.minItemSize = pagePreferences.minItemSize;
            this.maxItemSize = pagePreferences.maxItemSize;

            this.padding = pagePreferences.padding;

            this.displayTitles = pagePreferences.displayTitles;
            this.roundedCorners = pagePreferences.roundedCorners;

            this.dirty = false;
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

        markDirty: function () {
            this.dirty = true;
            if (!this.saveChangesTimer) {
                this.saveChangesTimer = setTimeout(() => {
                    this.saveChangesTimer = null;
                    this.saveChanges();
                }, 1000);
            }
        },

        saveChanges: function () {
            if (this.saveChangesTimer) {
                clearTimeout(this.saveChangesTimer);
                this.saveChangesTimer = null;
            }

            if (!this.dirty) {
                return;
            }

            setPagePreferences({
                pageSize: this.pageSize,
                rowSize: this.rowSize,
                rowSizeMin: this.rowSizeMin,
                minItemSize: this.minItemSize,
                maxItemSize: this.maxItemSize,
                padding: this.padding,
                displayTitles: this.displayTitles,
                roundedCorners: this.roundedCorners,
            });

            this.dirty = false;
        },

        resetDefaultValues: function () {
            resetPagePreferences();
            this.reset();
        },
    },
    mounted: function () {
        this.reset();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        if (this.saveChangesTimer) {
            clearTimeout(this.saveChangesTimer);
            this.saveChangesTimer = null;

            this.saveChanges();
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
