<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Toggle play delay") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="s in TOGGLE_DELAY_OPTIONS"
                :key="s"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeToggleDelay(s)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== toggleDelay }"></i>
                    {{ renderToggleDelay(s, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderToggleDelay } from "@/utils/player-config";

// Options for toggle play delay
const TOGGLE_DELAY_OPTIONS = [0, 250, 500, 750, 1000];

// Translation
const { $t } = useI18n();

// Toggle play delay
const toggleDelay = defineModel<number>("toggleDelay");

/**
 * Changes toggle play delay
 * @param d The delay
 */
const changeToggleDelay = (d: number) => {
    toggleDelay.value = d;
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
