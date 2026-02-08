<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal" :static="true">
        <div v-if="display" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Upload files") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Files to upload") }} ({{ files.length }}):</label>
                    <textarea class="form-control form-control-full-width form-textarea" readonly :value="renderFiles(files)"></textarea>
                </div>

                <div class="form-group">
                    <label>{{ $t("Total size") }}: {{ computeTotalSize(files) }}</label>
                </div>

                <div class="form-group">
                    <label>{{ $t("Select an album to add the uploaded media into") }}:</label>
                    <AlbumSelect v-model:album="album" :disabled="inModal"></AlbumSelect>
                </div>
                <div v-if="!inModal" class="form-group">
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

<script lang="ts">
import type { PropType } from "vue";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import AlbumSelect from "../utils/AlbumSelect.vue";
import TagNameList from "../utils/TagNameList.vue";

const AlbumCreateModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumCreateModal.vue"),
});

export default defineComponent({
    name: "UploadModal",
    components: {
        AlbumCreateModal,
        AlbumSelect,
        TagNameList,
    },
    props: {
        display: Boolean,
        inModal: Boolean,
        fixedAlbum: Number,
        files: Array as PropType<File[]>,
    },
    emits: ["update:display", "upload"],
    setup(props) {
        return {
            findTagTimeout: null as ReturnType<typeof setTimeout> | null,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            tags: [] as string[],
            album: -1,

            displayAlbumCreate: false,

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.reset();
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.reset();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        if (this.findTagTimeout) {
            clearTimeout(this.findTagTimeout);
            this.findTagTimeout = null;
        }
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        renderFiles: function (files: File[]) {
            return files
                .map((file) => {
                    return file.name + " (" + this.renderSize(file.size) + ")";
                })
                .join("\n");
        },

        computeTotalSize: function (files: File[]) {
            let size = 0;

            for (const file of files) {
                size += file.size || 0;
            }

            return this.renderSize(size);
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

        reset: function () {
            this.album = -1;
            this.tags = [];
        },

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        doUpload: function () {
            this.$emit("upload", this.album, this.tags);
            this.close();
        },

        createAlbum: function () {
            this.displayAlbumCreate = true;
        },

        onNewAlbum: function (albumId: number) {
            this.album = albumId;
        },

        skipTagSuggestions: function () {
            const el = this.$el.querySelector(".modal-footer-btn");
            if (el) {
                el.focus();
            }
        },

        onTagsSkipKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && event.shiftKey) {
                const inputElem = this.$el.querySelector(".tags-input-search");
                if (inputElem) {
                    event.preventDefault();
                    inputElem.focus();
                }
            }
        },
    },
});
</script>
