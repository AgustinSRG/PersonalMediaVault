<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Scale") }} ({{ $t("Custom") }})</b>
                </td>
                <td class="td-right"></td>
            </tr>

            <tr>
                <td colspan="2">
                    <input
                        v-model.number="scaleNum"
                        type="range"
                        class="form-range"
                        :min="100"
                        :max="800"
                        :step="1"
                        @input="updateScaleNum"
                    />
                </td>
            </tr>

            <tr>
                <td colspan="2" class="custom-size-row">
                    <input
                        v-model.number="scaleNum"
                        type="number"
                        class="form-control custom-size-input"
                        :min="1"
                        :step="1"
                        @input="updateScaleNum"
                    />
                    <b class="custom-size-unit">%</b>
                </td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { ref, watch } from "vue";

// Translation
const { $t } = useI18n();

// Player scale
const scale = defineModel<number>("scale");

// Scale number
const scaleNum = ref(Math.floor(scale.value * 100));

watch(scale, () => {
    scaleNum.value = Math.floor(scale.value * 100);
});

/**
 * Call when the scale number is updated,
 * in order to update the scale value.
 */
const updateScaleNum = () => {
    if (typeof scaleNum.value !== "number" || isNaN(scaleNum.value) || scaleNum.value < 0.1) {
        return;
    }

    scale.value = scaleNum.value / 100;
};

// Emits
const emit = defineEmits<{
    /**
     * Users requests to go back
     */
    (e: "go-back"): void;
}>();

/**
 * Goes back to the parent page
 */
const goBack = () => {
    emit("go-back");
};
</script>
