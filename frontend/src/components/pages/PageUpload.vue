<template>
    <div class="page-inner-padded" :class="{ 'page-inner': !inModal, hidden: !display, 'scrollbar-stable': !inModal }" @drop="onDrop">
        <div class="upload-option-container">
            <div class="upload-option-label">{{ $t("Max number of uploads in parallel") }}:</div>
            <button
                v-for="v in parallelOptions"
                :key="v"
                type="button"
                class="page-option"
                :class="{ current: maxParallelUploads === v }"
                @click="updateMaxParallelUploads(v)"
            >
                {{ v }}
            </button>
        </div>
        <input type="file" class="file-hidden" name="media-upload" multiple @change="inputFileChanged" />
        <div
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
            :files="files"
            @upload="onUploadConfirmed"
        >
        </UploadModal>
    </div>
</template>

<script lang="ts">
import { AppStatus } from "@/control/app-status";
import type { UploadEntryMin } from "@/control/upload";
import {
    EVENT_NAME_UPLOAD_LIST_CLEAR,
    EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR,
    EVENT_NAME_UPLOAD_LIST_ENTRY_NEW,
    EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS,
    EVENT_NAME_UPLOAD_LIST_ENTRY_READY,
    EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED,
    EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY,
    UploadController,
} from "@/control/upload";
import { getFrontendUrl } from "@/utils/api";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import { renderSize } from "@/utils/size";
import DiskUsage from "../utils/DiskUsage.vue";
import { stringMultiReplace } from "@/utils/string-multi-replace";

const UploadModal = defineAsyncComponent({
    loader: () => import("@/components/modals/UploadModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

function removeEntryFromList(list: UploadEntryMin[], e: UploadEntryMin) {
    for (let i = 0; i < list.length; i++) {
        if (list[i].id === e.id) {
            list.splice(i, 1);
            return;
        }
    }
}

export default defineComponent({
    name: "PageUpload",
    components: {
        UploadModal,
        DiskUsage,
    },
    props: {
        display: Boolean,
        inModal: Boolean,
        fixedAlbum: Number,
    },
    emits: ["media-go"],
    data: function () {
        return {
            dragging: false,

            entriesPending: [] as UploadEntryMin[],
            entriesReady: [] as UploadEntryMin[],
            entriesError: [] as UploadEntryMin[],

            maxParallelUploads: UploadController.GetMaxParallelUploads(),

            selectedState: "pending" as "pending" | "ready" | "error",

            filteredEntries: [] as UploadEntryMin[],

            displayUploadModal: false,
            files: [] as File[],

            parallelOptions: [1, 2, 4, 8, 16, 32, 64],
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.refreshUploadList();

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_CLEAR, () => {
            this.refreshUploadList();
        });

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_NEW, (e: UploadEntryMin) => {
            this.entriesPending.push(e);
            if (this.selectedState === "pending") {
                this.filteredEntries.push(e);
            }
        });

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, (e: UploadEntryMin) => {
            const entry = this.entriesPending.find((p) => p.id === e.id);

            if (entry) {
                entry.status = e.status;
                entry.progress = e.progress;
                entry.mid = e.mid;
            }
        });

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_READY, (e: UploadEntryMin) => {
            removeEntryFromList(this.entriesPending, e);

            this.entriesReady.push(e);

            if (this.selectedState === "pending") {
                removeEntryFromList(this.filteredEntries, e);
            } else if (this.selectedState === "ready") {
                this.filteredEntries.unshift(e);
            }
        });

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR, (e: UploadEntryMin) => {
            removeEntryFromList(this.entriesPending, e);

            this.entriesError.push(e);

            if (this.selectedState === "pending") {
                removeEntryFromList(this.filteredEntries, e);
            } else if (this.selectedState === "error") {
                this.filteredEntries.unshift(e);
            }
        });

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY, (e: UploadEntryMin) => {
            removeEntryFromList(this.entriesError, e);

            this.entriesError.push(e);

            if (this.selectedState === "error") {
                removeEntryFromList(this.filteredEntries, e);
            } else if (this.selectedState === "pending") {
                this.filteredEntries.push(e);
            }
        });

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED, (e: UploadEntryMin) => {
            switch (e.status) {
                case "ready":
                    removeEntryFromList(this.entriesReady, e);
                    if (this.selectedState === "ready") {
                        removeEntryFromList(this.filteredEntries, e);
                    }
                    break;
                case "error":
                    removeEntryFromList(this.entriesError, e);
                    if (this.selectedState === "error") {
                        removeEntryFromList(this.filteredEntries, e);
                    }
                    break;
                default:
                    removeEntryFromList(this.entriesPending, e);
                    if (this.selectedState === "pending") {
                        removeEntryFromList(this.filteredEntries, e);
                    }
            }
        });

        if (this.display) {
            this.autoFocus();
        }
    },
    methods: {
        clickToSelect: function () {
            const fileElem = this.$el.querySelector(".file-hidden");
            if (fileElem) {
                fileElem.value = null;
                fileElem.click();
            }
        },

        autoFocus: function () {
            nextTick(() => {
                const el = this.$el.querySelector(".auto-focus");
                if (el) {
                    el.focus();
                    if (el.select) {
                        el.select();
                    }
                }
            });
        },

        updateSelectedState: function (s: "pending" | "ready" | "error") {
            this.selectedState = s;
            this.updateFilteredEntries();
        },

        updateFilteredEntries: function () {
            switch (this.selectedState) {
                case "ready":
                    this.filteredEntries = this.entriesReady.slice();
                    break;
                case "error":
                    this.filteredEntries = this.entriesError.slice();
                    break;
                default:
                    this.filteredEntries = this.entriesPending.slice();
            }

            if (this.selectedState !== "pending") {
                this.filteredEntries = this.filteredEntries.reverse();
            }
        },

        updateMaxParallelUploads: function (m: number) {
            this.maxParallelUploads = m;
            UploadController.SetMaxParallelUploads(m);
        },

        inputFileChanged: function (e: InputEvent) {
            const data = (e.target as HTMLInputElement).files;
            if (data && data.length > 0) {
                const files = [];
                for (const file of data) {
                    files.push(file);
                }
                this.addFiles(files);
            }
        },

        onDrop: function (e: DragEvent) {
            e.preventDefault();
            this.dragging = false;
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const files = [];
                for (const file of data) {
                    files.push(file);
                }
                this.addFiles(files);
            }
        },

        dragOver: function (e: DragEvent) {
            e.preventDefault();
        },
        dragEnter: function (e: DragEvent) {
            e.preventDefault();
            this.dragging = true;
        },
        dragLeave: function (e: DragEvent) {
            e.preventDefault();
            this.dragging = false;
        },

        renderSize: renderSize,

        addFiles: function (files: File[]) {
            this.files = files;
            this.displayUploadModal = true;
            this.updateSelectedState("pending");
        },

        onUploadConfirmed: function (album: number, tags: string[]) {
            for (const file of this.files) {
                UploadController.AddFile(file, album, tags.slice());
            }
            this.files = [];
        },

        removeFile: function (id: number) {
            UploadController.RemoveFile(id);
        },

        clearList: function () {
            UploadController.ClearList();
        },

        cancelAll: function () {
            UploadController.CancelAll();
        },

        tryAgain: function (m: UploadEntryMin) {
            UploadController.TryAgain(m.id);
        },

        goToMedia: function (m: UploadEntryMin, e?: MouseEvent) {
            if (e) {
                e.preventDefault();
            }
            if (m.mid < 0) {
                return;
            }
            AppStatus.ClickOnMedia(m.mid, true);
            this.$emit("media-go");
        },

        renderStatus(status: string, p: number, err: string) {
            switch (status) {
                case "ready":
                    return this.$t("Ready");
                case "pending":
                    return this.$t("Pending");
                case "uploading":
                    if (p > 0) {
                        return this.$t("Uploading") + "... (" + p + "%)";
                    } else {
                        return this.$t("Uploading") + "...";
                    }
                case "encrypting":
                    if (p > 0) {
                        return this.$t("Encrypting") + "... (" + p + "%)";
                    } else {
                        return this.$t("Encrypting") + "...";
                    }
                case "tag":
                    return this.$t("Adding tags") + "... (" + stringMultiReplace(this.$t("$N left"), { $N: "" + p }) + ")";
                case "album":
                    return this.$t("Inserting into album") + "...";
                case "error":
                    switch (err) {
                        case "invalid-media":
                            return this.$t("Error") + ": " + this.$t("Invalid media file provided");
                        case "access-denied":
                            return this.$t("Error") + ": " + this.$t("Access denied");
                        case "deleted":
                            return this.$t("Error") + ": " + this.$t("The media asset was deleted");
                        case "no-internet":
                            return this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        default:
                            return this.$t("Error") + ": " + this.$t("Internal server error");
                    }
                default:
                    return "-";
            }
        },

        cssProgress: function (status: string, p: number) {
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
        },

        getMediaURL: function (mid: number): string {
            return getFrontendUrl({
                media: mid,
            });
        },

        refreshUploadList: function () {
            this.entriesPending = UploadController.GetPendingEntries();
            this.entriesReady = UploadController.GetReadyEntries();
            this.entriesError = UploadController.GetErrorEntries();

            this.updateFilteredEntries();
        },
    },
});
</script>
