<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Playback speed") }} ({{ $t("Custom") }})</b>
                </td>
                <td class="td-right"></td>
            </tr>

            <tr>
                <td colspan="2">
                    <input
                        v-model.number="speedNum"
                        type="range"
                        class="form-range"
                        :min="1"
                        :max="200"
                        :step="1"
                        @input="updateSpeedNum"
                    />
                </td>
            </tr>

            <tr>
                <td colspan="2" class="custom-size-row">
                    <input
                        v-model.number="speedNum"
                        type="number"
                        class="form-control custom-size-input"
                        :min="1"
                        :step="1"
                        @input="updateSpeedNum"
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

// Playback speed
const speed = defineModel<number>("speed");

// Speed number
const speedNum = ref(Math.floor(speed.value * 100));

watch(speed, () => {
    speedNum.value = Math.floor(speed.value * 100);
});

/**
 * Call when the speed number is updated,
 * in order to update the speed value.
 */
const updateSpeedNum = () => {
    if (typeof speedNum.value !== "number" || isNaN(speedNum.value) || speedNum.value < 0.1) {
        return;
    }

    speed.value = speedNum.value / 100;
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
