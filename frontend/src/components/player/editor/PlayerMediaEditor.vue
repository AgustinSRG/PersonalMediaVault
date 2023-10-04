<template>
    <div class="player-media-editor" tabindex="-1">
        <div
            class="horizontal-filter-menu"
            :class="{
                'three-child': (!canWrite && type === 2) || (!canWrite && type === 3) || (type === 1 && canWrite),
                'four-child': (type === 2 && canWrite) || (type === 1 && !canWrite),
                'five-child': (type == 3 && !canWrite),
            }"
        >
            <a
                href="javascript:;"
                @click="changePage('general')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'general' }"
                >{{ $t("General") }}</a
            >
            <a href="javascript:;" @click="changePage('tags')" class="horizontal-filter-menu-item" :class="{ selected: page === 'tags' }">{{
                $t("Tags")
            }}</a>
            <a
                v-if="type === 2 || type === 3"
                href="javascript:;"
                @click="changePage('subtitles')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'subtitles' }"
                >{{ $t("Subtitles") }}</a
            >
            <a
                v-if="type === 2"
                href="javascript:;"
                @click="changePage('audios')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'audios' }"
                >{{ $t("Audio tracks") }}</a
            >
            <a
                v-if="type === 2 || type === 3"
                href="javascript:;"
                @click="changePage('time-slices')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'time-slices' }"
                >{{ $t("Time slices") }}</a
            >
            <a
                v-if="type === 1"
                href="javascript:;"
                @click="changePage('image-notes')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'image-notes' }"
                >{{ $t("Image notes") }}</a
            >
            <a
                href="javascript:;"
                @click="changePage('attachments')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'attachments' }"
                >{{ $t("Attachments") }}</a
            >
            <a
                v-if="(type === 1 || type === 2) && canWrite"
                href="javascript:;"
                @click="changePage('resolutions')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'resolutions' }"
                >{{ $t("Resolutions") }}</a
            >
            <a
                v-if="canWrite"
                href="javascript:;"
                @click="changePage('danger')"
                class="horizontal-filter-menu-item"
                :class="{ selected: page === 'danger' }"
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
import { AppEvents } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
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
    name: "PlayerMediaEditor",
    emits: ["changed"],
    data: function () {
        return {
            page: "general",

            type: 0,

            busy: false,

            canWrite: AuthController.CanWrite,
        };
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

    mounted: function () {
        this._handles = Object.create(null);
        this.updateMediaData();

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this._handles.mediaUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);
    },
});
</script>
