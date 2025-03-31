<template>
    <div
        class="modal-container modal-container-corner no-transition modal-container-page-settings"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
        @mousedown="close"
        @touchstart="close"
        @keydown="keyDownHandle"
    >
        <div
            v-if="display"
            class="modal-dialog modal-lg"
            role="document"
            @click="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
        >
            <div class="modal-header-corner">
                <div class="modal-header-corner-title">{{ $t("Page settings") }}</div>
            </div>
            <div class="modal-body with-menu limited-height">
                <div class="page-settings">
                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Max items per page") }}</div>
                        <div class="page-settings-range">
                            <input
                                v-model.number="pageSize"
                                type="range"
                                class="form-range"
                                :min="1"
                                :max="256"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                v-model.number="pageSize"
                                type="number"
                                class="form-control form-control-full-width"
                                :min="1"
                                :max="256"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Preferred row size") }}</div>
                        <div class="page-settings-range">
                            <input
                                v-model.number="rowSize"
                                type="range"
                                class="form-range"
                                :min="computedMinRowSize"
                                :max="computedMaxRowSize"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                v-model.number="rowSize"
                                type="number"
                                class="form-control form-control-full-width"
                                :min="1"
                                :max="256"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Preferred row size") }} ({{ $t("For screen split mode") }})</div>
                        <div class="page-settings-range">
                            <input
                                v-model.number="rowSizeMin"
                                type="range"
                                class="form-range"
                                :min="computedMinRowSizeMin"
                                :max="computedMaxRowSizeMin"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                v-model.number="rowSizeMin"
                                type="number"
                                class="form-control form-control-full-width"
                                :min="1"
                                :max="256"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Min size of cells (pixels)") }}</div>
                        <div class="page-settings-range">
                            <input
                                v-model.number="minItemSize"
                                type="range"
                                class="form-range"
                                :min="25"
                                :max="250"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                v-model.number="minItemSize"
                                type="number"
                                class="form-control form-control-full-width"
                                :min="25"
                                :max="500"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Max size of cells (pixels)") }}</div>
                        <div class="page-settings-range">
                            <input
                                v-model.number="maxItemSize"
                                type="range"
                                class="form-range"
                                :min="250"
                                :max="750"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                v-model.number="maxItemSize"
                                type="number"
                                class="form-control form-control-full-width"
                                :min="25"
                                :max="1000"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Cell padding (pixels)") }}</div>
                        <div class="page-settings-range">
                            <input
                                v-model.number="padding"
                                type="range"
                                class="form-range"
                                :min="0"
                                :max="32"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                v-model.number="padding"
                                type="number"
                                class="form-control form-control-full-width"
                                :min="0"
                                :max="64"
                                :step="1"
                                @input="markDirty"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Display titles?") }}</div>
                        <div class="page-settings-input page-settings-toggle">
                            <toggle-switch v-model:val="displayTitles" @update:val="markDirty"></toggle-switch>
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Use rounded borders for corners?") }}</div>
                        <div class="page-settings-input page-settings-toggle">
                            <toggle-switch v-model:val="roundedCorners" @update:val="markDirty"></toggle-switch>
                        </div>
                    </div>
                </div>
                <div class="page-settings-btn-container">
                    <button type="button" class="btn btn-primary btn-sm" @click="resetDefaultValues">
                        <i class="fas fa-sync-alt"></i> {{ $t("Reset to default values") }}
                    </button>

                    <button type="button" class="btn btn-primary btn-sm" @click="close">
                        <i class="fas fa-check"></i> {{ $t("Done") }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { getPagePreferences, resetPagePreferences, setPagePreferences } from "@/control/app-preferences";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "@/utils/v-model";
import { FocusTrap } from "@/utils/focus-trap";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";

export default defineComponent({
    name: "PageSettingsDropdown",
    components: {
        ToggleSwitch,
    },
    props: {
        display: Boolean,
        page: String,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
            saveChangesTimer: null,
        };
    },
    data: function () {
        const pagePreferences = getPagePreferences(this.page);
        return {
            dirty: false,

            pageSize: pagePreferences.pageSize,

            rowSize: pagePreferences.rowSize,
            rowSizeMin: pagePreferences.rowSizeMin,

            minItemSize: pagePreferences.minItemSize,
            maxItemSize: pagePreferences.maxItemSize,

            padding: pagePreferences.padding,

            displayTitles: pagePreferences.displayTitles,
            roundedCorners: pagePreferences.roundedCorners,
        };
    },
    computed: {
        computedMaxRowSize() {
            return Math.max(2, Math.floor(window.innerWidth / Math.max(1, this.minItemSize)));
        },

        computedMinRowSize() {
            return Math.max(1, Math.floor(window.innerWidth / Math.max(1, this.maxItemSize)));
        },

        computedMaxRowSizeMin() {
            return Math.max(2, Math.floor(Math.min(500, window.innerWidth) / Math.max(1, this.minItemSize)));
        },

        computedMinRowSizeMin() {
            return Math.max(1, Math.floor(Math.min(500, window.innerWidth) / Math.max(1, this.maxItemSize)));
        },
    },
    watch: {
        display: function () {
            if (this.display) {
                this.reset();
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },
        page: function () {
            if (this.display) {
                this.reset();
            }
        },
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "page-header-btn");

        this.reset();

        if (this.display) {
            this.focusTrap.activate();
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        if (this.saveChangesTimer) {
            clearTimeout(this.saveChangesTimer);
            this.saveChangesTimer = null;

            this.saveChanges();
        }

        this.focusTrap.destroy();
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        stopPropagationEvent: function (e: Event) {
            e.stopPropagation();
        },

        keyDownHandle: function (e: KeyboardEvent) {
            e.stopPropagation();
            if (e.key === "Escape") {
                this.close();
            }
        },

        reset: function () {
            if (this.saveChangesTimer) {
                clearTimeout(this.saveChangesTimer);
                this.saveChangesTimer = null;
            }

            const pagePreferences = getPagePreferences(this.page);

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

        markDirty: function () {
            this.dirty = true;
            if (!this.saveChangesTimer) {
                this.saveChangesTimer = setTimeout(() => {
                    this.saveChangesTimer = null;
                    this.saveChanges();
                }, 500);
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

            setPagePreferences(this.page, {
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
            resetPagePreferences(this.page);
            this.reset();
        },
    },
});
</script>
