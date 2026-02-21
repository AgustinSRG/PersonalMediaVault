<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Audio") }}</b>
                </td>
                <td class="td-right"></td>
            </tr>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeAudioTrack('')">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': '' !== audioTrack }"></i>
                    {{ renderAudio(metadata, "", rTick, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="aud in metadata.audios"
                :key="aud.id"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeAudioTrack(aud.id)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': aud.id !== audioTrack }"></i>
                    {{ aud.name }}
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
import { renderAudio } from "@/utils/player-config";
import type { PropType } from "vue";

// Translation
const { $t } = useI18n();

// Selected audio track (for videos)
const audioTrack = defineModel<string>("audioTrack");

/**
 * Changes the resolution
 * @param a The audio track ID
 */
const changeAudioTrack = (a: string) => {
    audioTrack.value = a;
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
