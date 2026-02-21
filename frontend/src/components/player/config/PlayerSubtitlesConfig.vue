<template>
    <table>
        <tbody>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                <td>
                    <i class="fas fa-chevron-left icon-config"></i>
                    <b>{{ $t("Subtitles") }}</b>
                </td>
                <td class="td-right" @click="goToSubtitlesOptions">
                    <a href="#subtitle-options" @click="goToSubtitlesOptions">{{ $t("Style options") }}</a>
                </td>
            </tr>
            <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeSubtitle('')">
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': '' !== effectiveSubtitles }"></i>
                    {{ renderSubtitle(metadata, "", rTick, $t) }}
                </td>
                <td class="td-right"></td>
            </tr>
            <tr
                v-for="sub in metadata.subtitles"
                :key="sub.id"
                class="tr-button"
                tabindex="0"
                @keydown="clickOnEnter"
                @click="changeSubtitle(sub.id)"
            >
                <td>
                    <i class="fas fa-check icon-config" :class="{ 'check-uncheck': sub.id !== effectiveSubtitles }"></i>
                    {{ sub.name }}
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
import { renderSubtitle } from "@/utils/player-config";
import type { PropType } from "vue";
import { ref, watch } from "vue";

// Translation
const { $t } = useI18n();

// Subtitles
const subtitles = defineModel<string>("subtitles");

/**
 * Changes the subtitle
 * @param s The subtitle ID
 */
const changeSubtitle = (s: string) => {
    subtitles.value = s;
};

// Props
const props = defineProps({
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

    /**
     * Users requests to go to subtitles options
     */
    (e: "go-subtitles-options"): void;
}>();

/**
 * Goes back to the parent page
 */
const goBack = () => {
    emit("go-back");
};

/**
 * Goes to the subtitles options page
 * @param e The click event
 */
const goToSubtitlesOptions = (e?: Event) => {
    if (e) {
        e.preventDefault();
        e.stopPropagation();
    }

    emit("go-subtitles-options");
};

// Effective subtitles
const effectiveSubtitles = ref("");

/**
 * Updates effective subtitles
 */
const updateEffectiveSubtitles = () => {
    if (!props.metadata || !props.metadata.subtitles || !subtitles.value) {
        effectiveSubtitles.value = "";
        return;
    }

    for (const sub of props.metadata.subtitles) {
        if (sub.id === subtitles.value) {
            effectiveSubtitles.value = sub.id;
            return;
        }
    }

    if (subtitles.value && props.metadata.subtitles.length > 0) {
        effectiveSubtitles.value = props.metadata.subtitles[0].id;
    } else {
        effectiveSubtitles.value = "";
    }
};

updateEffectiveSubtitles();
watch(() => props.rTick, updateEffectiveSubtitles);
watch(subtitles, updateEffectiveSubtitles);
</script>
