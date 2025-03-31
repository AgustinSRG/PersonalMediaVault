<template>
    <div class="player-media-editor" tabindex="-1">
        <div
            class="horizontal-filter-menu"
            :class="{
                'three-child': (!canWrite && type === 2) || (!canWrite && type === 3) || (type === 1 && canWrite),
                'four-child': (type === 2 && canWrite) || (type === 1 && !canWrite),
                'five-child': type == 3 && !canWrite,
            }"
        >
            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'general' }"
                @click="changePage('general')"
                >{{ $t("General") }}</a
            >
            <a href="javascript:;" class="horizontal-filter-menu-item" :class="{ selected: page === 'tags' }" @click="changePage('tags')">{{
                $t("Tags")
            }}</a>
            <a
                v-if="type === 2 || type === 3"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'subtitles' }"
                @click="changePage('subtitles')"
                >{{ $t("Subtitles") }}</a
            >
            <a
                v-if="type === 2"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'audios' }"
                @click="changePage('audios')"
                >{{ $t("Audio tracks") }}</a
            >
            <a
                v-if="type === 2 || type === 3"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'time-slices' }"
                @click="changePage('time-slices')"
                >{{ $t("Time slices") }}</a
            >
            <a
                v-if="type === 1"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'image-notes' }"
                @click="changePage('image-notes')"
                >{{ $t("Image notes") }}</a
            >
            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'attachments' }"
                @click="changePage('attachments')"
                >{{ $t("Attachments") }}</a
            >
            <a
                v-if="(type === 1 || type === 2) && canWrite"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'resolutions' }"
                @click="changePage('resolutions')"
                >{{ $t("Resolutions") }}</a
            >
            <a
                v-if="canWrite"
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'danger' }"
                @click="changePage('danger')"
                >{{ $t("Danger zone") }}</a
            >
        </div>

        <EditorGeneral v-if="page === 'general'" @changed="onChanged"></EditorGeneral>
        <EditorTags v-else-if="page === 'tags'" @changed="onChanged"></EditorTags>
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
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { defineAsyncComponent, defineComponent } from "vue";

const EditorGeneral = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorGeneral.vue"),
});

const EditorTags = defineAsyncComponent({
    loader: () => import("@/components/player/editor/EditorTags.vue"),
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
        EditorTags,
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
            this.page = page;
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
