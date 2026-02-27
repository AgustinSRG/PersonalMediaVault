<template>
    <div class="player-media-editor" tabindex="-1">
        <div
            class="horizontal-filter-menu"
            :class="{
                'can-write': canWrite,
                'image-opts-menu': type === MEDIA_TYPE_IMAGE,
                'video-opts-menu': type === MEDIA_TYPE_VIDEO,
                'audio-opts-menu': type == MEDIA_TYPE_AUDIO,
            }"
        >
            <a
                v-for="p in availablePages"
                :key="p.id"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="p.title"
                :class="{ selected: page === p.id }"
                @click="changePage(p.id)"
                >{{ p.title }}</a
            >
        </div>

        <EditorGeneral v-if="page === 'general'" @changed="onChanged"></EditorGeneral>
        <EditorDescription v-if="page === 'description'" @changed="onChanged"></EditorDescription>
        <EditorRelatedMedia v-else-if="page === 'related'" @changed="onChanged"></EditorRelatedMedia>
        <EditorSubtitles v-else-if="page === 'subtitles'" @changed="onChanged"></EditorSubtitles>
        <EditorAudios v-else-if="page === 'audios'" @changed="onChanged"></EditorAudios>
        <EditorAttachments v-else-if="page === 'attachments'" @changed="onChanged"></EditorAttachments>
        <EditorTimeSlices v-else-if="page === 'time-slices'" @changed="onChanged"></EditorTimeSlices>
        <EditorImageNotes v-else-if="page === 'image-notes'" @changed="onChanged"></EditorImageNotes>
        <EditorResolutions v-else-if="page === 'resolutions'" @changed="onChanged"></EditorResolutions>
        <EditorDangerZone v-else-if="page === 'danger'"></EditorDangerZone>
    </div>
</template>

<script setup lang="ts">
import type { MediaType } from "@/api/models";
import { MEDIA_TYPE_AUDIO, MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO } from "@/api/models";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { ExitPreventer } from "@/control/exit-prevent";
import { MediaController } from "@/control/media";
import { computed, defineAsyncComponent, ref } from "vue";

const EditorGeneral = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorGeneral.vue"),
});

const EditorDescription = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorDescription.vue"),
});

const EditorRelatedMedia = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorRelatedMedia.vue"),
});

const EditorSubtitles = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorSubtitles.vue"),
});

const EditorAudios = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorAudios.vue"),
});

const EditorAttachments = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorAttachments.vue"),
});

const EditorTimeSlices = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorTimeSlices.vue"),
});

const EditorImageNotes = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorImageNotes.vue"),
});

const EditorResolutions = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorResolutions.vue"),
});

const EditorDangerZone = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorDangerZone.vue"),
});

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Emits
const emit = defineEmits<{
    /**
     * Emitted when changes are made to the media
     */
    (e: "changed"): void;
}>();

// Media type
const type = ref<MediaType>(MediaController.MediaData?.type || 0);

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, () => {
    if (!MediaController.MediaData) {
        return;
    }

    type.value = MediaController.MediaData.type || 0;
});

// Editor pages
type EditorPage =
    | "general"
    | "description"
    | "related"
    | "subtitles"
    | "audios"
    | "attachments"
    | "time-slices"
    | "image-notes"
    | "resolutions"
    | "danger";

// Metadata for editor page
type EditorPageMetadata = {
    // ID
    id: EditorPage;

    /**
     * A function to get the title
     * @returns The title
     */
    title: () => string;

    /**
     * A function to check the visibility of the page
     * @returns True if the poge is visible
     */
    visible: () => boolean;
};

// List of pages, in order
const PAGES: EditorPageMetadata[] = [
    {
        id: "general",
        title: () => $t("General"),
        visible: () => true,
    },

    {
        id: "description",
        title: () => $t("Description"),
        visible: () => true,
    },

    {
        id: "attachments",
        title: () => $t("Attachments"),
        visible: () => true,
    },

    {
        id: "related",
        title: () => $t("Related media"),
        visible: () => true,
    },

    {
        id: "subtitles",
        title: () => $t("Subtitles"),
        visible: () => [MEDIA_TYPE_AUDIO, MEDIA_TYPE_VIDEO].includes(type.value),
    },

    {
        id: "audios",
        title: () => $t("Audio tracks"),
        visible: () => type.value === MEDIA_TYPE_VIDEO,
    },

    {
        id: "time-slices",
        title: () => $t("Time slices"),
        visible: () => [MEDIA_TYPE_AUDIO, MEDIA_TYPE_VIDEO].includes(type.value),
    },

    {
        id: "image-notes",
        title: () => $t("Image notes"),
        visible: () => type.value === MEDIA_TYPE_IMAGE,
    },

    {
        id: "resolutions",
        title: () => $t("Resolutions"),
        visible: () => [MEDIA_TYPE_VIDEO, MEDIA_TYPE_IMAGE].includes(type.value),
    },

    {
        id: "danger",
        title: () => $t("Danger zone"),
        visible: () => canWrite.value,
    },
];

// Current page
const page = ref<EditorPage>("general");

// Available pages
const availablePages = computed(() =>
    PAGES.map((p) => {
        return {
            id: p.id,
            title: p.title(),
            visible: p.visible(),
        };
    }).filter((p) => p.visible),
);

/**
 * Changes the page
 * @param p The page
 */
const changePage = (p: EditorPage) => {
    if (p === page.value) {
        return;
    }
    ExitPreventer.TryExit(() => {
        page.value = p;
    });
};

/**
 * Called when the media changes
 */
const onChanged = () => {
    emit("changed");
};
</script>
