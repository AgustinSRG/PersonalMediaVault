<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Quality") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeResolution(-1)">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': -1 !== resolution }"></i>
                    {{ renderResolution(metadata, -1, rTick, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="(_r, i) in metadata.resolutions"
                :key="i"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeResolution(i)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': i !== resolution }"></i>
                    {{ renderResolution(metadata, i, rTick, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
        </tbody>
    </table>
</template>

<script setup lang="ts">
import type { MediaData } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { renderResolution } from "@/utils/player-config";
import type { PropType } from "vue";

// Translation
const { $t } = useI18n();

// Resolution index
const resolution = defineModel<number>("resolution");

/**
 * Changes the resolution
 * @param i The resolution index
 */
const changeResolution = (i: number) => {
    resolution.value = i;
};

// Props
defineProps({
    /**
     * Media metadata
     */
    metadata: {
        type: Object as PropType<MediaData>,
        required: true,
    },

    /**
     * Resolution load tick.
     * When updated, the resolutions should be updated.
     */
    rTick: Number,
});

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
