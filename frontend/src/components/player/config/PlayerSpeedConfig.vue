<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Playback speed") }}</b>
                </td>
                <td class="td-right" @click="goToCustomSpeed">
                    <a href="#playback-speed-custom" @click="goToCustomSpeed">{{ $t("Custom") }}</a>
                </td>
            </tr>
            <tr
                v-for="s in DEFAULT_PLAYBACK_SPEEDS"
                :key="s"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeSpeed(s)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': s !== speed }"></i>
                    {{ renderSpeed(s, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-if="!DEFAULT_PLAYBACK_SPEEDS.includes(speed)"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeSpeed(speed)"
            >
                <td>
                    <i class="fas fa-check icon-config"></i>
                    {{ $t("Custom") }}: {{ renderSpeed(speed, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderSpeed } from "@/utils/player-config";

// Default playback speeds
const DEFAULT_PLAYBACK_SPEEDS = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2];

// Translation
const { $t } = useI18n();

// Playback speed
const speed = defineModel<number>("speed");

/**
 * Changes the playback speed
 * @param s The speed
 */
const changeSpeed = (s: number) => {
    speed.value = s;
};

// Emits
const emit = defineEmits<{
    /**
     * Users requests to go back
     */
    (e: "go-back"): void;

    /**
     * Users requests to go to custom speed
     */
    (e: "go-custom-speed"): void;
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
const goToCustomSpeed = (e?: Event) => {
    if (e) {
        e.preventDefault();
        e.stopPropagation();
    }

    emit("go-custom-speed");
};
</script>
