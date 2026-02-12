<template>
    <ModalDialogContainer ref="container" v-model:display="display" :static="true">
        <div ref="form" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Upload files") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Files to upload") }} ({{ files.length }}):</label>
                    <textarea class="form-control form-control-full-width form-textarea" readonly :value="renderedFiles"></textarea>
                </div>

                <div class="form-group">
                    <label>{{ $t("Total size") }}: {{ renderedTotalSize }}</label>
                </div>

                <div class="form-group">
                    <label>{{ $t("Select an album to add the uploaded media into") }}:</label>
                    <AlbumSelect v-model:album="album" :disabled="isFixedAlbum"></AlbumSelect>
                </div>
                <div v-if="!isFixedAlbum" class="form-group">
                    <button type="button" class="btn btn-primary btn-sm" @click="createAlbum">
                        <i class="fas fa-plus"></i> {{ $t("Create album") }}
                    </button>
                </div>

                <div class="form-group">
                    <label>{{ $t("Tags to automatically add to the uploaded media") }}:</label>
                </div>

                <TagNameList v-model:tags="tags" @tab-focus-skip="skipTagSuggestions"></TagNameList>
            </div>

            <div class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn" @click="doUpload" @keydown="onTagsSkipKeyDown">
                    <i class="fas fa-upload"></i> {{ $t("Upload") }}
                </button>
            </div>
        </div>

        <AlbumCreateModal v-if="displayAlbumCreate" v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { computed, defineAsyncComponent, ref, useTemplateRef, watch } from "vue";
import AlbumSelect from "../utils/AlbumSelect.vue";
import TagNameList from "../utils/TagNameList.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { renderSize } from "@/utils/size";

const AlbumCreateModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumCreateModal.vue"),
});

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * Fixed album
     */
    fixedAlbum: {
        type: Number,
        default: -1,
    },

    /**
     * List of files to upload
     */
    files: {
        type: Array as PropType<File[]>,
        required: true,
    },
});

// Is a fixed album
const isFixedAlbum = computed(() => props.fixedAlbum >= 0);

// Rendered list of files
const renderedFiles = computed(() =>
    props.files
        .map((file) => {
            return file.name + " (" + renderSize(file.size) + ")";
        })
        .join("\n"),
);

// Rendered total size
const renderedTotalSize = computed(() => {
    let size = 0;

    for (const file of props.files) {
        size += file.size || 0;
    }

    return renderSize(size);
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when upload is confirmed
     */
    (e: "upload", album: number, tags: string[]): void;
}>();

// Tags to be added to the uploaded media
const tags = ref<string[]>([]);

// Album for the uploaded media to be added
const album = ref(-1);

// Displays the modal to create an album
const displayAlbumCreate = ref(false);

/**
 * Resets the form
 */
const reset = () => {
    album.value = -1;
    tags.value = [];

    displayAlbumCreate.value = false;
};

watch(display, () => {
    if (display.value) {
        reset();
    }
});

/**
 * Confirms upload
 */
const doUpload = () => {
    emit("upload", album.value, tags.value);
    close();
};

/**
 * Displays the modal to create an album
 */
const createAlbum = () => {
    displayAlbumCreate.value = true;
};

/**
 * Called when a new album is created
 * @param albumId The id of the created album
 */
const onNewAlbum = (albumId: number) => {
    album.value = albumId;
};

// Form container
const form = useTemplateRef("form");

/**
 * Skips to the next element from tag suggestions
 */
const skipTagSuggestions = () => {
    const el = form.value?.querySelector(".modal-footer-btn") as HTMLElement;
    if (el) {
        el.focus();
    }
};

/**
 * Handler for 'keydown' on the next element to the tag suggestions
 * @param event The keyboard event
 */
const onTagsSkipKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && event.shiftKey) {
        const inputElem = form.value?.querySelector(".tags-input-search") as HTMLElement;
        if (inputElem) {
            event.preventDefault();
            inputElem.focus();
        }
    }
};
</script>
