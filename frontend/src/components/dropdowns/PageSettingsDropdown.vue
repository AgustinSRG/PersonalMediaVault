<template>
    <div
        class="modal-container modal-container-corner no-transition modal-container-page-settings"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
        :aria-hidden="!display"
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
                                type="range"
                                class="form-range"
                                v-model.number="pageSize"
                                @input="markDirty"
                                :min="1"
                                :max="256"
                                :step="1"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                type="number"
                                class="form-control form-control-full-width"
                                v-model.number="pageSize"
                                @input="markDirty"
                                :min="1"
                                :max="256"
                                :step="1"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Preferred row size") }}</div>
                        <div class="page-settings-range">
                            <input
                                type="range"
                                class="form-range"
                                v-model.number="rowSize"
                                @input="markDirty"
                                :min="computedMinRowSize"
                                :max="computedMaxRowSize"
                                :step="1"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                type="number"
                                class="form-control form-control-full-width"
                                v-model.number="rowSize"
                                @input="markDirty"
                                :min="1"
                                :max="256"
                                :step="1"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Preferred row size") }} ({{ $t("For screen split mode") }})</div>
                        <div class="page-settings-range">
                            <input
                                type="range"
                                class="form-range"
                                v-model.number="rowSizeMin"
                                @input="markDirty"
                                :min="computedMinRowSizeMin"
                                :max="computedMaxRowSizeMin"
                                :step="1"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                type="number"
                                class="form-control form-control-full-width"
                                v-model.number="rowSizeMin"
                                @input="markDirty"
                                :min="1"
                                :max="256"
                                :step="1"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Min size of cells (pixels)") }}</div>
                        <div class="page-settings-range">
                            <input
                                type="range"
                                class="form-range"
                                v-model.number="minItemSize"
                                @input="markDirty"
                                :min="25"
                                :max="250"
                                :step="1"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                type="number"
                                class="form-control form-control-full-width"
                                v-model.number="minItemSize"
                                @input="markDirty"
                                :min="25"
                                :max="500"
                                :step="1"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Max size of cells (pixels)") }}</div>
                        <div class="page-settings-range">
                            <input
                                type="range"
                                class="form-range"
                                v-model.number="maxItemSize"
                                @input="markDirty"
                                :min="250"
                                :max="750"
                                :step="1"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                type="number"
                                class="form-control form-control-full-width"
                                v-model.number="maxItemSize"
                                @input="markDirty"
                                :min="25"
                                :max="1000"
                                :step="1"
                            />
                        </div>
                    </div>

                    <div class="page-settings-row">
                        <div class="page-settings-label">{{ $t("Cell padding (pixels)") }}</div>
                        <div class="page-settings-range">
                            <input
                                type="range"
                                class="form-range"
                                v-model.number="padding"
                                @input="markDirty"
                                :min="0"
                                :max="32"
                                :step="1"
                            />
                        </div>
                        <div class="page-settings-input">
                            <input
                                type="number"
                                class="form-control form-control-full-width"
                                v-model.number="padding"
                                @input="markDirty"
                                :min="0"
                                :max="64"
                                :step="1"
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
import { AppStatus } from "@/control/app-status";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";

export default defineComponent({
    name: "PageSettingsDropdown",
    components: {
        ToggleSwitch,
    },
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
            saveChangesTimer: null,
        };
    },
    data: function () {
        const pagePreferences = getPagePreferences();
        return {
            dirty: false,

            page: AppStatus.CurrentPage,

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
    },
});
</script>
<style scoped>
.page-settings {
    display: flex;
    flex-direction: column;
}

.page-settings-row {
    display: flex;
    flex-direction: row;
    align-items: center;
    padding: 0.5rem 1rem;
}

.page-settings-label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
}

.page-settings-range {
    width: 16rem;
    padding-left: 1rem;
}

.page-settings-input {
    width: 8rem;
    padding-left: 1rem;
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
}

.page-settings-toggle {
    height: 2rem;
}

.page-settings-btn-container {
    border-top: solid 1px var(--theme-border-color);
    padding-top: 1rem;
    padding-left: 1rem;
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: flex-end;
    align-items: center;
}

.page-settings-btn-container .btn {
    margin-right: 1rem;
    margin-bottom: 1rem;
}

@media (max-width: 600px) {
    .page-settings-row {
        flex-wrap: wrap;
    }

    .page-settings-label {
        width: 100%;
        flex: auto;
        padding-bottom: 0.5rem;
    }

    .page-settings-range {
        flex: 1;
    }

    .page-settings-input {
        justify-content: flex-start;
        width: fit-content;
    }
}

@media (max-width: 350px) {
    .page-settings-range {
        flex: auto;
        width: 100%;
        padding-bottom: 0.5rem;
    }

    .page-settings-input {
        width: 100%;
    }
}
</style>
