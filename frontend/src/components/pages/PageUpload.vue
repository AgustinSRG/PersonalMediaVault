<template>
    <div class="page-inner-padded" :class="{ 'page-inner': !inModal, 'scrollbar-stable': !inModal }" @drop="onDrop">
        <div class="upload-option-container">
            <div class="upload-option-label">{{ $t("Max number of uploads in parallel") }}:</div>
            <button
                v-for="v in MAX_PARALLEL_UPLOADS_OPTIONS"
                :key="v"
                type="button"
                class="page-option"
                :class="{ current: maxParallelUploads === v }"
                @click="updateMaxParallelUploads(v)"
            >
                {{ v }}
            </button>
        </div>
        <input ref="hiddenFileInput" type="file" class="file-hidden" name="media-upload" multiple @change="inputFileChanged" />
        <div
            ref="autoFocusElement"
            class="upload-box auto-focus"
            :class="{ dragging: dragging }"
            tabindex="0"
            @click="clickToSelect"
            @dragover="dragOver"
            @dragenter="dragEnter"
            @dragstart="dragEnter"
            @dragend="dragLeave"
            @dragleave="dragLeave"
            @drop="onDrop"
            @keydown="clickOnEnter"
        >
            <div class="upload-box-hint">
                {{ $t("Drop file here or click to open the file selection dialog.") }}
            </div>
        </div>

        <div class="upload-option-container margin-top">
            <DiskUsage></DiskUsage>
        </div>

        <div class="horizontal-filter-menu">
            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Pending')"
                :class="{ selected: selectedState === 'pending' }"
                @click="updateSelectedState('pending')"
                >{{ $t("Pending") }} ({{ entriesPending.length }})</a
            >
            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Ready')"
                :class="{ selected: selectedState === 'ready' }"
                @click="updateSelectedState('ready')"
                >{{ $t("Ready") }} ({{ entriesReady.length }})</a
            >
            <a
                href="javascript:;"
                class="horizontal-filter-menu-item"
                :title="$t('Error')"
                :class="{ selected: selectedState === 'error' }"
                @click="updateSelectedState('error')"
                >{{ $t("Error") }} ({{ entriesError.length }})</a
            >
        </div>

        <div class="upload-list">
            <div v-for="m in filteredEntries" :key="m.id" class="upload-list-item">
                <div class="upload-list-item-top">
                    <div class="upload-list-item-file-name">
                        <span v-if="m.status !== 'ready'" class="bold">{{ m.name }}</span>
                        <a
                            v-if="m.status === 'ready'"
                            class="bold"
                            :href="getMediaURL(m.mid)"
                            target="_blank"
                            rel="noopener noreferrer"
                            @click="goToMedia(m, $event)"
                            >{{ m.name }}</a
                        >
                    </div>
                    <div class="upload-list-item-file-size">
                        <span>{{ renderSize(m.size) }}</span>
                    </div>
                </div>
                <div class="upload-list-item-bottom">
                    <div class="upload-list-item-status">
                        <div class="upload-list-item-status-bar">
                            <div
                                class="upload-list-item-status-bar-current"
                                :class="{ error: m.status === 'error', success: m.status === 'ready' }"
                                :style="{ width: cssProgress(m.status, m.progress) }"
                            ></div>
                            <div class="upload-list-item-status-bar-text">{{ renderStatus(m.status, m.progress, m.error) }}</div>
                        </div>
                    </div>
                    <div class="upload-list-item-right">
                        <button
                            v-if="m.status === 'pending' || m.status === 'uploading' || m.status === 'encrypting' || m.status === 'tag'"
                            type="button"
                            class="table-btn"
                            :title="$t('Cancel upload')"
                            @click="removeFile(m.id)"
                        >
                            <i class="fas fa-times"></i>
                        </button>
                        <button v-if="m.status === 'ready'" type="button" class="table-btn" :title="$t('View media')" @click="goToMedia(m)">
                            <i class="fas fa-eye"></i>
                        </button>
                        <button v-if="m.status === 'ready'" type="button" class="table-btn" :title="$t('Done')" @click="removeFile(m.id)">
                            <i class="fas fa-check"></i>
                        </button>
                        <button v-if="m.status === 'error'" type="button" class="table-btn" :title="$t('Try again')" @click="tryAgain(m)">
                            <i class="fas fa-rotate"></i>
                        </button>
                        <button v-if="m.status === 'error'" type="button" class="table-btn" :title="$t('Remove')" @click="removeFile(m.id)">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="entriesPending.length > 0 || entriesReady.length > 0 || entriesError.length > 0" class="upload-table table-responsive">
            <div v-if="entriesReady.length > 0 || entriesError.length > 0" class="form-group">
                <button type="button" class="btn btn-primary" @click="clearList">
                    <i class="fas fa-broom"></i> {{ $t("Clear list") }}
                </button>
            </div>
            <div v-if="entriesPending.length > 0" class="form-group">
                <button type="button" class="btn btn-primary" @click="cancelAll">
                    <i class="fas fa-times"></i> {{ $t("Cancel all uploads") }}
                </button>
            </div>
        </div>

        <UploadModal
            v-if="displayUploadModal"
            v-model:display="displayUploadModal"
            :in-modal="inModal"
            :fixed-album="fixedAlbum"
            :files="filesToUpload"
            @upload="onUploadConfirmed"
        >
        </UploadModal>
    </div>
</template>

<script setup lang="ts">
import { AppStatus } from "@/control/app-status";
import type { UploadEntryMin } from "@/control/upload";
import { UploadController } from "@/control/upload";
import { getFrontendUrl } from "@/utils/api";
import { defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import { renderSize } from "@/utils/size";
import DiskUsage from "./common/DiskUsage.vue";
import { stringMultiReplace } from "@/utils/string-multi-replace";
import {
    EVENT_NAME_UPLOAD_LIST_CLEAR,
    EVENT_NAME_UPLOAD_LIST_ENTRY_NEW,
    EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS,
    EVENT_NAME_UPLOAD_LIST_ENTRY_READY,
    EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR,
    EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY,
    EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED,
} from "@/control/app-events";
import { useI18n } from "@/composables/use-i18n";
import { onApplicationEvent } from "@/composables/on-app-event";

const UploadModal = defineAsyncComponent({
    loader: () => import("@/components/modals/UploadModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

/**
 * Removes upload entry from upload entry list
 * @param list Upload entry list
 * @param e Element to remove
 */
function removeEntryFromList(list: UploadEntryMin[], e: UploadEntryMin) {
    for (let i = 0; i < list.length; i++) {
        if (list[i].id === e.id) {
            list.splice(i, 1);
            return;
        }
    }
}

// Translation
const { $t } = useI18n();

// Props
defineProps({
    /**
     * Page is being displayed in a modal
     */
    inModal: Boolean,

    /**
     * If set, any uploads will be added into this album
     */
    fixedAlbum: Number,
});

// Emits
const emit = defineEmits<{
    /**
     * Emitted when the user navigates into a media element
     */
    (e: "media-go"): void;
}>();

// True if the user is dragging something into the upload box
const dragging = ref(false);

// List of pending entries
const entriesPending = ref<UploadEntryMin[]>([]);

// List of ready entries
const entriesReady = ref<UploadEntryMin[]>([]);

// List of errored entries
const entriesError = ref<UploadEntryMin[]>([]);

// Allows values for the max parallel uploads option
const MAX_PARALLEL_UPLOADS_OPTIONS = [1, 2, 4, 8, 16, 32, 64];

// Max number of allowed parallel uploads
const maxParallelUploads = ref(UploadController.GetMaxParallelUploads());

// Base statuses of entries in order to classify them
type UploadEntryBaseStatus = "pending" | "ready" | "error";

// Current selected status to display
const selectedState = ref<UploadEntryBaseStatus>("pending");

// List of filtered upload entries to actually display
const filteredEntries = ref<UploadEntryMin[]>([]);

// True to display the upload modal
const displayUploadModal = ref(false);

// List of files to upload
const filesToUpload = ref<File[]>([]);

/**
 * Updates filtered upload entries
 */
const updateFilteredEntries = () => {
    switch (selectedState.value) {
        case "ready":
            filteredEntries.value = entriesReady.value.slice();
            break;
        case "error":
            filteredEntries.value = entriesError.value.slice();
            break;
        default:
            filteredEntries.value = entriesPending.value.slice();
    }

    if (selectedState.value !== "pending") {
        filteredEntries.value = filteredEntries.value.reverse();
    }
};

/**
 * Refreshes uploads entries lists
 */
const refreshUploadList = () => {
    entriesPending.value = UploadController.GetPendingEntries();
    entriesReady.value = UploadController.GetReadyEntries();
    entriesError.value = UploadController.GetErrorEntries();

    updateFilteredEntries();
};

refreshUploadList();

// List cleared: Full refresh
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_CLEAR, refreshUploadList);

// New entry added
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_NEW, (e: UploadEntryMin) => {
    entriesPending.value.push(e);
    if (selectedState.value === "pending") {
        filteredEntries.value.push(e);
    }
});

// A pending entry changed its progress
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, (e: UploadEntryMin) => {
    const entry = entriesPending.value.find((p) => p.id === e.id);

    if (entry) {
        entry.status = e.status;
        entry.progress = e.progress;
        entry.mid = e.mid;
    }
});

// A pending entry is now ready
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_READY, (e: UploadEntryMin) => {
    removeEntryFromList(entriesPending.value, e);

    entriesReady.value.push(e);

    if (selectedState.value === "pending") {
        removeEntryFromList(filteredEntries.value, e);
    } else if (selectedState.value === "ready") {
        filteredEntries.value.unshift(e);
    }
});

// A pending entry failed
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR, (e: UploadEntryMin) => {
    removeEntryFromList(entriesPending.value, e);

    entriesError.value.push(e);

    if (selectedState.value === "pending") {
        removeEntryFromList(filteredEntries.value, e);
    } else if (selectedState.value === "error") {
        filteredEntries.value.unshift(e);
    }
});

// An errored entry is being retried
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY, (e: UploadEntryMin) => {
    removeEntryFromList(entriesError.value, e);

    entriesPending.value.push(e);

    if (selectedState.value === "error") {
        removeEntryFromList(filteredEntries.value, e);
    } else if (selectedState.value === "pending") {
        filteredEntries.value.push(e);
    }
});

// Entry removed
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED, (e: UploadEntryMin) => {
    switch (e.status) {
        case "ready":
            removeEntryFromList(entriesReady.value, e);
            if (selectedState.value === "ready") {
                removeEntryFromList(filteredEntries.value, e);
            }
            break;
        case "error":
            removeEntryFromList(entriesError.value, e);
            if (selectedState.value === "error") {
                removeEntryFromList(filteredEntries.value, e);
            }
            break;
        default:
            removeEntryFromList(entriesPending.value, e);
            if (selectedState.value === "pending") {
                removeEntryFromList(filteredEntries.value, e);
            }
    }
});

// Element to be auto-focused on load
const autoFocusElement = useTemplateRef("autoFocusElement");

/**
 * Automatically focuses the appropriate element on load
 */
const autoFocus = () => {
    nextTick(() => {
        autoFocusElement.value?.focus();
    });
};

onMounted(autoFocus);

// Hidden file input element
const hiddenFileInput = useTemplateRef("hiddenFileInput");

/**
 * User clicked on the upload box.
 * The file input must be triggered.
 */
const clickToSelect = () => {
    if (hiddenFileInput.value) {
        hiddenFileInput.value.value = null;
        hiddenFileInput.value.click();
    }
};

/**
 * Updates the selected state
 * @param s The state
 */
const updateSelectedState = (s: UploadEntryBaseStatus) => {
    selectedState.value = s;
    updateFilteredEntries();
};

/**
 * Updates the max parallel uploads option
 * @param m The max parallel uploads option
 */
const updateMaxParallelUploads = (m: number) => {
    maxParallelUploads.value = m;
    UploadController.SetMaxParallelUploads(m);
};

/**
 * Removes a file
 * @param id The file entry ID
 */
const removeFile = (id: number) => {
    UploadController.RemoveFile(id);
};

/**
 * Clears the list
 */
const clearList = () => {
    UploadController.ClearList();
};

/**
 * Cancels all uploads
 */
const cancelAll = () => {
    UploadController.CancelAll();
};

/**
 * Retries an errored entry
 * @param m The entry
 */
const tryAgain = (m: UploadEntryMin) => {
    UploadController.TryAgain(m.id);
};

/**
 * Navigates into an uploaded media element
 * @param m The upload entry
 * @param e The click event on the link
 */
const goToMedia = (m: UploadEntryMin, e?: MouseEvent) => {
    if (e) {
        e.preventDefault();
    }

    if (m.mid < 0) {
        return;
    }

    AppStatus.ClickOnMedia(m.mid, true);

    emit("media-go");
};

/**
 * Adds files to be uploaded and displays the confirmation modal
 * @param files List of files
 */
const addFiles = (files: File[]) => {
    filesToUpload.value = files;
    displayUploadModal.value = true;
    updateSelectedState("pending");
};

/**
 * Called when the user confirms an upload
 * @param album The album for the media to be added into
 * @param tags The tags for the uploaded media
 */
const onUploadConfirmed = (album: number, tags: string[]) => {
    for (const file of filesToUpload.value) {
        UploadController.AddFile(file, album, tags.slice());
    }
    filesToUpload.value = [];
};

/**
 * Event handler for 'changed' on the file input
 * @param e The event
 */
const inputFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        const files = [];
        for (const file of data) {
            files.push(file);
        }
        addFiles(files);
    }
};

/**
 * Event handler for 'drop' on the upload box
 * @param e The event
 */
const onDrop = (e: DragEvent) => {
    e.preventDefault();

    dragging.value = false;

    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        const files = [];
        for (const file of data) {
            files.push(file);
        }
        addFiles(files);
    }
};

/**
 * Event handler for 'dragover' on the upload box
 * @param e The event
 */
const dragOver = (e: DragEvent) => {
    e.preventDefault();
};

/**
 * Event handler for 'dragenter' on the upload box
 * @param e The event
 */
const dragEnter = (e: DragEvent) => {
    e.preventDefault();
    dragging.value = true;
};

/**
 * Event handler for 'dragleave' on the upload box
 * @param e The event
 */
const dragLeave = (e: DragEvent) => {
    e.preventDefault();
    dragging.value = false;
};

/**
 * Gets the URL for a media element
 * @param mid The media element ID
 * @returns The URL
 */
const getMediaURL = (mid: number): string => {
    return getFrontendUrl({
        media: mid,
    });
};

/**
 * Renders the status of an entry
 * @param status The status
 * @param p The progress (0-1)
 * @param err The error message
 * @returns The rendered status
 */
const renderStatus = (status: string, p: number, err: string) => {
    switch (status) {
        case "ready":
            return $t("Ready");
        case "pending":
            return $t("Pending");
        case "uploading":
            if (p > 0) {
                return $t("Uploading") + "... (" + p + "%)";
            } else {
                return $t("Uploading") + "...";
            }
        case "encrypting":
            if (p > 0) {
                return $t("Encrypting") + "... (" + p + "%)";
            } else {
                return $t("Encrypting") + "...";
            }
        case "tag":
            return $t("Adding tags") + "... (" + stringMultiReplace($t("$N left"), { $N: "" + p }) + ")";
        case "album":
            return $t("Inserting into album") + "...";
        case "error":
            switch (err) {
                case "invalid-media":
                    return $t("Error") + ": " + $t("Invalid media file provided");
                case "access-denied":
                    return $t("Error") + ": " + $t("Access denied");
                case "deleted":
                    return $t("Error") + ": " + $t("The media asset was deleted");
                case "no-internet":
                    return $t("Error") + ": " + $t("Could not connect to the server");
                default:
                    return $t("Error") + ": " + $t("Internal server error");
            }
        default:
            return "-";
    }
};

/**
 * Turns thew progress into percentage for CSS style value
 * @param status The current status
 * @param p The progress (0-1)
 * @returns The progress as CSS-compatible percentage
 */
const cssProgress = (status: string, p: number) => {
    p = Math.min(100, Math.max(0, p));
    switch (status) {
        case "uploading":
            return Math.round((p * 50) / 100) + "%";
        case "encrypting":
            return Math.round(50 + (p * 50) / 100) + "%";
        case "ready":
        case "error":
        case "tag":
            return "100%";
        default:
            return "0";
    }
};
</script>
