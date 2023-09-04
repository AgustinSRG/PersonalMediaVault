<template>
    <div class="page-inner-padded" :class="{ 'page-inner': !inModal, hidden: !display }" @drop="onDrop">
        <div class="upload-parallel-options">
            <div class="upload-parallel-label">{{ $t("Max number of uploads in parallel") }}:</div>
            <button
                v-for="v in parallelOptions"
                :key="v"
                type="button"
                class="upload-parallel-option"
                :class="{ current: maxParallelUploads === v }"
                @click="updateMaxParallelUploads(v)"
            >
                {{ v }}
            </button>
        </div>
        <input type="file" class="file-hidden" @change="inputFileChanged" name="media-upload" multiple />
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

        <div class="horizontal-filter-menu">
            <a
                href="javascript:;"
                @click="updateSelectedState('pending')"
                class="horizontal-filter-menu-item"
                :class="{ selected: this.selectedState === 'pending' }"
                >{{ $t("Pending") }} ({{ countPending }})</a
            >
            <a
                href="javascript:;"
                @click="updateSelectedState('ready')"
                class="horizontal-filter-menu-item"
                :class="{ selected: this.selectedState === 'ready' }"
                >{{ $t("Ready") }} ({{ countReady }})</a
            >
            <a
                href="javascript:;"
                @click="updateSelectedState('error')"
                class="horizontal-filter-menu-item"
                :class="{ selected: this.selectedState === 'error' }"
                >{{ $t("Error") }} ({{ countError }})</a
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
                            @click="goToMedia(m, $event)"
                            :href="getMediaURL(m.mid)"
                            target="_blank"
                            rel="noopener noreferrer"
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

        <div class="upload-table table-responsive" v-if="pendingToUpload.length > 0">
            <div class="form-group" v-if="pendingToUpload.length > 0">
                <button type="button" class="btn btn-primary" @click="clearList">
                    <i class="fas fa-broom"></i> {{ $t("Clear list") }}
                </button>
            </div>
            <div class="form-group" v-if="countCancellable > 0">
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
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { UploadController, UploadEntryMin } from "@/control/upload";
import { GenerateURIQuery } from "@/utils/request";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";

import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";

const UploadModal = defineAsyncComponent({
    loader: () => import("@/components/modals/UploadModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const STATE_FILTER_PENDING = ["pending", "uploading", "encrypting", "tag"];
const STATE_FILTER_READY = ["ready"];
const STATE_FILTER_ERROR = ["error"];

export default defineComponent({
    components: {
        UploadModal,
    },
    name: "PageUpload",
    emits: ["media-go"],
    props: {
        display: Boolean,
        inModal: Boolean,
        fixedAlbum: Number,
    },
    data: function () {
        return {
            dragging: false,
            pendingToUpload: [],
            countCancellable: 0,

            optionsShown: false,

            maxParallelUploads: UploadController.MaxParallelUploads,

            countPending: 0,
            countReady: 0,
            countError: 0,

            stateFilter: STATE_FILTER_PENDING.slice(),
            selectedState: "pending",

            filteredEntries: [],

            displayUploadModal: false,
            files: [],

            parallelOptions: [1, 2, 4, 8, 16, 32, 64],
        };
    },
    methods: {
        clickToSelect: function () {
            this.$el.querySelector(".file-hidden").click();
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

        updateSelectedState: function (s: string) {
            this.selectedState = s;
            switch (s) {
                case "ready":
                    this.stateFilter = STATE_FILTER_READY.slice();
                    break;
                case "error":
                    this.stateFilter = STATE_FILTER_ERROR.slice();
                    break;
                default:
                    this.stateFilter = STATE_FILTER_PENDING.slice();
            }
            this.updateFilteredEntries();
        },

        updateFilteredEntries: function () {
            this.filteredEntries = this.pendingToUpload.filter((e: UploadEntryMin) => {
                return this.stateFilter.includes(e.status);
            });

            if (this.selectedState !== "pending") {
                this.filteredEntries = this.filteredEntries.reverse();
            }
        },

        updateMaxParallelUploads: function (m: number) {
            this.maxParallelUploads = m;
            UploadController.MaxParallelUploads = m;
        },

        inputFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                let files = [];
                for (let file of data) {
                    files.push(file);
                }
                this.addFiles(files);
            }
        },

        onDrop: function (e) {
            e.preventDefault();
            this.dragging = false;
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                let files = [];
                for (let file of data) {
                    files.push(file);
                }
                this.addFiles(files);
            }
        },

        dragOver: function (e) {
            e.preventDefault();
        },
        dragEnter: function (e) {
            e.preventDefault();
            this.dragging = true;
        },
        dragLeave: function (e) {
            e.preventDefault();
            this.dragging = false;
        },

        renderSize: function (bytes: number): string {
            if (bytes > 1024 * 1024 * 1024) {
                let gb = bytes / (1024 * 1024 * 1024);
                gb = Math.floor(gb * 100) / 100;
                return gb + " GB";
            } else if (bytes > 1024 * 1024) {
                let mb = bytes / (1024 * 1024);
                mb = Math.floor(mb * 100) / 100;
                return mb + " MB";
            } else if (bytes > 1024) {
                let kb = bytes / 1024;
                kb = Math.floor(kb * 100) / 100;
                return kb + " KB";
            } else {
                return bytes + " Bytes";
            }
        },

        addFiles: function (files: File[]) {
            this.files = files;
            this.displayUploadModal = true;
            this.updateSelectedState("pending");
        },

        onUploadConfirmed: function (album: number, tags: string[]) {
            for (let file of this.files) {
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
                    return this.$t("Adding tags") + "... (" + this.$t("$N left").replace("$N", "" + p) + ")";
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

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },

        showOptions: function (b: boolean) {
            this.optionsShown = b;
        },

        onPendingPush: function (m: UploadEntryMin) {
            this.pendingToUpload.push(m);
            this.updateCountCancellable(this.pendingToUpload);

            if (this.stateFilter.includes(m.status)) {
                if (this.selectedState === "pending") {
                    this.filteredEntries.push(m);
                } else {
                    this.filteredEntries.unshift(m);
                }
            }
        },

        onPendingRemove: function (i: number) {
            const removed = this.pendingToUpload.splice(i, 1)[0];
            this.updateCountCancellable(this.pendingToUpload);

            if (removed) {
                for (let j = 0; j < this.filteredEntries.length; j++) {
                    if (this.filteredEntries[j].id === removed.id) {
                        this.filteredEntries.splice(j, 1);
                        break;
                    }
                }
            }
        },

        onPendingClear: function () {
            this.pendingToUpload = UploadController.GetEntries();
            this.updateCountCancellable(this.pendingToUpload);
            this.updateFilteredEntries();
        },

        onPendingUpdate: function (i: number, m: UploadEntryMin) {
            let mustUpdate = this.pendingToUpload[i].status !== m.status;
            this.pendingToUpload[i].status = m.status;
            this.pendingToUpload[i].error = m.error;
            this.pendingToUpload[i].progress = m.progress;
            this.pendingToUpload[i].mid = m.mid;
            if (mustUpdate) {
                this.updateCountCancellable(this.pendingToUpload);
            }

            let found = false;

            for (let j = 0; j < this.filteredEntries.length; j++) {
                if (this.filteredEntries[j].id === m.id) {
                    found = true;
                    if (mustUpdate && !this.stateFilter.includes(m.status)) {
                        this.filteredEntries.splice(j, 1);
                    } else {
                        this.filteredEntries[j].status = m.status;
                        this.filteredEntries[j].error = m.error;
                        this.filteredEntries[j].progress = m.progress;
                        this.filteredEntries[j].mid = m.mid;
                    }
                }
            }

            if (!found && mustUpdate && this.stateFilter.includes(m.status)) {
                if (this.selectedState === "pending") {
                    this.filteredEntries.push(m);
                } else {
                    this.filteredEntries.unshift(m);
                }
            }
        },

        updateCountCancellable: function (list: UploadEntryMin[]) {
            let count = 0;
            let countPending = 0;
            let countReady = 0;
            let countError = 0;
            for (let l of list) {
                if (l.status !== "ready" && l.status !== "error") {
                    count++;
                }

                if (STATE_FILTER_PENDING.includes(l.status)) {
                    countPending++;
                } else if (STATE_FILTER_READY.includes(l.status)) {
                    countReady++;
                } else if (STATE_FILTER_ERROR.includes(l.status)) {
                    countError++;
                }
            }
            this.countCancellable = count;
            this.countPending = countPending;
            this.countReady = countReady;
            this.countError = countError;
        },

        getMediaURL: function (mid: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                GenerateURIQuery({
                    media: mid + "",
                })
            );
        },


    },
    mounted: function () {
        this._handles = Object.create(null);
        this.pendingToUpload = UploadController.GetEntries();
        this.updateCountCancellable(this.pendingToUpload);
        this.updateFilteredEntries();

        this._handles.onPendingPushH = this.onPendingPush.bind(this);
        this._handles.onPendingRemoveH = this.onPendingRemove.bind(this);
        this._handles.onPendingClearH = this.onPendingClear.bind(this);
        this._handles.onPendingUpdateH = this.onPendingUpdate.bind(this);

        AppEvents.AddEventListener("upload-list-push", this._handles.onPendingPushH);
        AppEvents.AddEventListener("upload-list-rm", this._handles.onPendingRemoveH);
        AppEvents.AddEventListener("upload-list-clear", this._handles.onPendingClearH);
        AppEvents.AddEventListener("upload-list-update", this._handles.onPendingUpdateH);

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("upload-list-push", this._handles.onPendingPushH);
        AppEvents.RemoveEventListener("upload-list-rm", this._handles.onPendingRemoveH);
        AppEvents.RemoveEventListener("upload-list-clear", this._handles.onPendingClearH);
        AppEvents.RemoveEventListener("upload-list-update", this._handles.onPendingUpdateH);

        if (this._handles.findTagTimeout) {
            clearTimeout(this._handles.findTagTimeout);
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
            }
        },
    }
});
</script>
