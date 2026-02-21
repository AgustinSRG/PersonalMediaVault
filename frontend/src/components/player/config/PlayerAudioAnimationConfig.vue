<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @click="goBack" @keydown="clickOnEnter">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Animation style") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr v-for="s in ANIM_STYLES" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="setAnimStyle(s)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== animColors }"></i>
                    {{ renderAnimStyle(s, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderAnimStyle } from "@/utils/player-config";

// Animation styles
const ANIM_STYLES = ["gradient", "rainbow", "", "none"];

// Translation
const { $t } = useI18n();

// Animation colors
const animColors = defineModel<string>("animColors");

/**
 * Sets animation style
 * @param s The style
 */
const setAnimStyle = (s: string) => {
    animColors.value = s;
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
