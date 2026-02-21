<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Scale") }}</b>
                </td>
                <td class="td-right" @click="goToCustomScale">
                    <a href="#video-scale-custom" @click="goToCustomScale">{{ $t("Custom") }}</a>
                </td>
            </tr>
            <tr v-for="s in DEFAULT_PLAYER_SCALES" :key="s" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeScale(s)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== scale }"></i>
                    {{ renderScale(s, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-if="!DEFAULT_PLAYER_SCALES.includes(scale)"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeScale(scale)"
            >
                <td>
                    <i class="fas fa-check icon-config"></i>
                    {{ $t("Custom") }}: {{ renderScale(scale, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderScale } from "@/utils/player-config";

// Default values for player scale
const DEFAULT_PLAYER_SCALES = [1, 1.25, 1.5, 1.75, 2, 4, 8];

// Translation
const { $t } = useI18n();

// Player scale
const scale = defineModel<number>("scale");

/**
 * Changes the scale value
 * @param s The value
 */
const changeScale = (s: number) => {
    scale.value = s;
};

// Emits
const emit = defineEmits<{
    /**
     * Users requests to go back
     */
    (e: "go-back"): void;

    /**
     * Users requests to go to custom scale
     */
    (e: "go-custom-scale"): void;
}>();

/**
 * Goes back to the parent page
 */
const goBack = () => {
    emit("go-back");
};

/**
 * Goes to the custom peed page
 * @param e The click event
 */
const goToCustomScale = (e?: Event) => {
    if (e) {
        e.preventDefault();
        e.stopPropagation();
    }

    emit("go-custom-scale");
};
</script>
