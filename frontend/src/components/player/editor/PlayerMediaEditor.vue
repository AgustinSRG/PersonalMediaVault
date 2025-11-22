<template>
    <div class="player-media-editor" tabindex="-1">
        <div
            class="horizontal-filter-menu"
            :class="{
                'can-write': canWrite,
                'image-opts-menu': type === 1,
                'video-opts-menu': type === 2,
                'audio-opts-menu': type == 3,
            }"
        >
            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('General')"
                :class="{ selected: page === 'general' }"
                @click="changePage('general')"
                >{{ $t("General") }}</a
            >

            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Description')"
                :class="{ selected: page === 'description' }"
                @click="changePage('description')"
                >{{ $t("Description") }}</a
            >

            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Attachments')"
                :class="{ selected: page === 'attachments' }"
                @click="changePage('attachments')"
                >{{ $t("Attachments") }}</a
            >

            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Related media')"
                :class="{ selected: page === 'related' }"
                @click="changePage('related')"
                >{{ $t("Related media") }}</a
            >

            <a
                v-if="type === 2 || type === 3"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Subtitles')"
                :class="{ selected: page === 'subtitles' }"
                @click="changePage('subtitles')"
                >{{ $t("Subtitles") }}</a
            >
            <a
                v-if="type === 2"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Audio tracks')"
                :class="{ selected: page === 'audios' }"
                @click="changePage('audios')"
                >{{ $t("Audio tracks") }}</a
            >
            <a
                v-if="type === 2 || type === 3"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Time slices')"
                :class="{ selected: page === 'time-slices' }"
                @click="changePage('time-slices')"
                >{{ $t("Time slices") }}</a
            >
            <a
                v-if="type === 1"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Image notes')"
                :class="{ selected: page === 'image-notes' }"
                @click="changePage('image-notes')"
                >{{ $t("Image notes") }}</a
            >

            <a
                v-if="type === 1 || type === 2"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Resolutions')"
                :class="{ selected: page === 'resolutions' }"
                @click="changePage('resolutions')"
                >{{ $t("Resolutions") }}</a
            >
            <a
                v-if="canWrite"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Danger zone')"
                :class="{ selected: page === 'danger' }"
                @click="changePage('danger')"
                >{{ $t("Danger zone") }}</a
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
        <EditorDangerZone v-else-if="page === 'danger'" @changed="onChanged"></EditorDangerZone>
    </div>
</template>

<script lang="ts">
import { AuthController, EVENT_NAME_AUTH_CHANGED } from "@/control/auth";
import { ExitPreventer } from "@/control/exit-prevent";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { defineAsyncComponent, defineComponent } from "vue";

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

export default defineComponent({
    name: "PlayerMediaEditor",
    components: {
        EditorGeneral,
        EditorDescription,
        EditorRelatedMedia,
        EditorSubtitles,
        EditorAudios,
        EditorAttachments,
        EditorTimeSlices,
        EditorImageNotes,
        EditorResolutions,
        EditorDangerZone,
    },
    emits: ["changed"],
    data: function () {
        return {
            page: "general",

            type: 0,

            busy: false,

            canWrite: AuthController.CanWrite,
        };
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    methods: {
        changePage: function (page: string) {
            if (page === this.page) {
                return;
            }
            ExitPreventer.TryExit(() => {
                this.page = page;
            });
        },

        onChanged: function () {
            this.$emit("changed");
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },
});
</script>
