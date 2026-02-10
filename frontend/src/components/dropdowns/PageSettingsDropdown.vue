<template>
    <DropdownContainer
        v-model:display="display"
        :position-class="'modal-container-page-settings'"
        :focus-trap-exception-class="'page-header-btn'"
    >
        <div class="modal-header-corner">
            <div class="modal-header-corner-title">{{ $t("Page settings") }}</div>
        </div>
        <div class="modal-body with-menu limited-height">
            <div class="page-settings">
                <div class="page-settings-row">
                    <div class="page-settings-label">{{ $t("Max items per page") }}</div>
                    <div class="page-settings-range">
                        <input v-model.number="pageSize" type="range" class="form-range" :min="1" :max="256" :step="1" @input="markDirty" />
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
                            @input="markDirtyAndApplyImmediately"
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
                            @input="markDirtyAndApplyImmediately"
                        />
                    </div>
                </div>

                <div class="page-settings-row">
                    <div class="page-settings-label">{{ $t("Display titles?") }}</div>
                    <div class="page-settings-input page-settings-toggle">
                        <toggle-switch v-model:val="displayTitles" @update:val="markDirtyAndApplyImmediately"></toggle-switch>
                    </div>
                </div>

                <div class="page-settings-row">
                    <div class="page-settings-label">{{ $t("Use rounded borders for corners?") }}</div>
                    <div class="page-settings-input page-settings-toggle">
                        <toggle-switch v-model:val="roundedCorners" @update:val="markDirtyAndApplyImmediately"></toggle-switch>
                    </div>
                </div>
            </div>
            <div class="page-settings-btn-container">
                <button type="button" class="btn btn-primary btn-sm" @click="resetDefaultValues">
                    <i class="fas fa-sync-alt"></i> {{ $t("Reset to default values") }}
                </button>

                <button type="button" class="btn btn-primary btn-sm" @click="close"><i class="fas fa-check"></i> {{ $t("Done") }}</button>
            </div>
        </div>
    </DropdownContainer>
</template>

<script setup lang="ts">
import DropdownContainer from "./common/DropdownContainer.vue";
import { getPagePreferences, resetPagePreferences, setPagePreferences } from "@/control/app-preferences";
import type { PropType } from "vue";
import { computed, onBeforeUnmount, ref, watch } from "vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import { useI18n } from "@/composables/use-i18n";
import { useTimeout } from "@/composables/use-timeout";
import type { AppStatusPage } from "@/control/app-status";

// Translation function
const { $t } = useI18n();

const props = defineProps({
    /**
     * ID of the page
     */
    page: {
        type: String as PropType<AppStatusPage>,
        required: true,
    },
});

// Display
const display = defineModel<boolean>("display");

/**
 * Closes the dropdown
 */
const close = () => {
    display.value = false;
};

// True if there are unsaved changes
const dirty = ref(false);

// Timeout for saving changes
const saveChangesTimeout = useTimeout();

// Initial page preferences
const initialPagePreferences = getPagePreferences(props.page);

// Max number of items per page
const pageSize = ref(initialPagePreferences.pageSize);

// Max number of items per row
const rowSize = ref(initialPagePreferences.rowSize);

// Max number of items per row (when page width is small)
const rowSizeMin = ref(initialPagePreferences.rowSizeMin);

// Min size (px) per item
const minItemSize = ref(initialPagePreferences.minItemSize);

// Max size (px) per item
const maxItemSize = ref(initialPagePreferences.maxItemSize);

// Padding (px) of items
const padding = ref(initialPagePreferences.padding);

// True to display titles
const displayTitles = ref(initialPagePreferences.displayTitles);

// True for rounded corners for items
const roundedCorners = ref(initialPagePreferences.roundedCorners);

/**
 * Resets the refs
 * from the stored preferences
 */
const reset = () => {
    saveChangesTimeout.clear();

    const storedPreferences = getPagePreferences(props.page);

    pageSize.value = storedPreferences.pageSize;

    rowSize.value = storedPreferences.rowSize;
    rowSizeMin.value = storedPreferences.rowSizeMin;

    minItemSize.value = storedPreferences.minItemSize;
    maxItemSize.value = storedPreferences.maxItemSize;

    padding.value = storedPreferences.padding;

    displayTitles.value = storedPreferences.displayTitles;
    roundedCorners.value = storedPreferences.roundedCorners;

    dirty.value = false;
};

// Reset when page Id changes
watch(() => props.page, reset);

// Reset when the dropdown is shown
watch(display, () => {
    if (display.value) {
        reset();
    }
});

// Computed max row size depending on the window size and the min item size
const computedMaxRowSize = computed(() => {
    return Math.max(2, Math.floor(window.innerWidth / Math.max(1, minItemSize.value)));
});

// Computed min row size depending on the window size and the max item size
const computedMinRowSize = computed(() => {
    return Math.max(1, Math.floor(window.innerWidth / Math.max(1, maxItemSize.value)));
});

// Computed max row size depending on the window size and the min item size
// (when window width is small)
const computedMaxRowSizeMin = computed(() => {
    return Math.max(2, Math.floor(Math.min(500, window.innerWidth) / Math.max(1, minItemSize.value)));
});

// Computed min row size depending on the window size and the max item size
// (when window width is small)
const computedMinRowSizeMin = computed(() => {
    return Math.max(1, Math.floor(Math.min(500, window.innerWidth) / Math.max(1, maxItemSize.value)));
});

/**
 * Saves page preferences
 */
const saveChanges = () => {
    saveChangesTimeout.clear();

    if (!dirty.value) {
        return;
    }

    setPagePreferences(props.page, {
        pageSize: pageSize.value,
        rowSize: rowSize.value,
        rowSizeMin: rowSizeMin.value,
        minItemSize: minItemSize.value,
        maxItemSize: maxItemSize.value,
        padding: padding.value,
        displayTitles: displayTitles.value,
        roundedCorners: roundedCorners.value,
    });

    dirty.value = false;
};

// Make sure changes are saved before the component unmounts
onBeforeUnmount(saveChanges);

// Delay to save changes (milliseconds)
const SAVE_CHANGES_DELAY = 500;

/**
 * Marks settings as dirty.
 * They will be automatically saved after a delay
 * or when the component unmounts.
 */
const markDirty = () => {
    dirty.value = true;

    if (!saveChangesTimeout.isSet()) {
        saveChangesTimeout.set(saveChanges, SAVE_CHANGES_DELAY);
    }
};

/**
 * Marks settings as dirty.
 * Applies changes immediately.
 */
const markDirtyAndApplyImmediately = () => {
    dirty.value = true;
    saveChanges();
};

/**
 * Resets page settings to its default values
 */
const resetDefaultValues = () => {
    resetPagePreferences(props.page);
    reset();
};
</script>
