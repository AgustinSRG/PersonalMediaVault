<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Auto next") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="b in AUTO_NEXT_DELAY_OPTIONS"
                :key="b"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeAutoNext(b)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': b !== autoNext }"></i>
                    {{ renderAutoNext(b, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderAutoNext } from "@/utils/player-config";

// Translation
const { $t } = useI18n();

// Emits
const emit = defineEmits<{
    /**
     * Users requests to go back
     */
    (e: "go-back"): void;
}>();

// Auto-next delay options
const AUTO_NEXT_DELAY_OPTIONS = [0, 3, 5, 10, 15, 20, 25, 30];

// Auto-next delay
const autoNext = defineModel<number>("autoNext");

/**
 * Goes back to the parent page
 */
const goBack = () => {
    emit("go-back");
};

/**
 * Changes auto-next value
 * @param v The new value
 */
const changeAutoNext = (v: number) => {
    autoNext.value = v;
};
</script>
