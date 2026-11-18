<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Volume multiplier") }} ({{ mediaTypeIndicator }})</b>
                </td>
                <td class="td-right"></td>
            </tr>

            <tr>
                <td colspan="2">
                    <input v-model.number="volumeMultiplier" type="range" class="form-range" :min="0" :max="2" :step="0.01" />
                </td>
            </tr>

            <tr>
                <td colspan="2" class="custom-size-row">
                    <b class="custom-size-unit-left">x</b>
                    <input
                        v-model.number="volumeMultiplier"
                        type="number"
                        class="form-control custom-size-input"
                        :min="0"
                        :max="100"
                        :step="0.01"
                    />
                    <a
                        href="#volume-multiplier-reset"
                        class="custom-pref-reset-link"
                        :class="{ hidden: volumeMultiplier === 1 }"
                        @click.prevent="resetVolumeMultiplier"
                        >{{ $t("Reset") }}</a
                    >
                </td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { MEDIA_TYPE_VIDEO } from "@/constants";
import { clickOnEnter } from "@/utils/events";
import { computed } from "vue";

// Translation
const { $t } = useI18n();

// Media type
const props = defineProps({
    /**
     * Media type
     */
    mediaType: {
        type: Number,
        required: true,
    },
});

// Media type indicator
const mediaTypeIndicator = computed<string>(() => {
    switch (props.mediaType) {
        case MEDIA_TYPE_VIDEO:
            return $t("for this video");
        default:
            return $t("for this audio");
    }
});

// Volume multiplier
const volumeMultiplier = defineModel<number>("volumeMultiplier");

/**
 * Resets volume multiplier
 */
const resetVolumeMultiplier = () => {
    volumeMultiplier.value = 1;
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
