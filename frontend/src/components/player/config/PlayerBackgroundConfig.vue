<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Background") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="b in BACKGROUND_OPTIONS"
                :key="b"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeBackground(b)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': b !== background }"></i>
                    {{ renderBackground(b, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderBackground } from "@/utils/player-config";

// Animation styles
const BACKGROUND_OPTIONS = ["default", "black", "white"];

// Translation
const { $t } = useI18n();

// Background
const background = defineModel<string>("background");

/**
 * Changes background
 * @param b The background
 */
const changeBackground = (b: string) => {
    background.value = b;
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
