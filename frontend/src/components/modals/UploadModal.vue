<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
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
                    <select v-model="album" :disabled="inModal" class="form-control form-select form-control-full-width">
                        <option :value="-1">--</option>
                        <option v-for="a in albums" :key="a.id" :value="a.id">
                            {{ a.name }}
                        </option>
                    </select>
                </div>
                <div class="form-group" v-if="!inModal">
                    <button type="button" @click="createAlbum" class="btn btn-primary btn-sm">
                        <i class="fas fa-plus"></i> {{ $t("Create album") }}
                    </button>
                </div>

                <div class="form-group">
                    <label>{{ $t("Tags to automatically add to the uploaded media") }}:</label>
                </div>
                <div class="form-group media-tags">
                    <div v-for="tag in tags" :key="tag" class="media-tag">
                        <div class="media-tag-name">{{ tag }}</div>
                        <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTag(tag)">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="media-tags-finder">
                        <input
                            type="text"
                            autocomplete="off"
                            maxlength="255"
                            v-model="tagToAdd"
                            @input="onTagAddChanged(false)"
                            @keydown="onTagInputKeyDown"
                            class="form-control tag-to-add auto-focus"
                            :placeholder="$t('Add tags') + '...'"
                        />
                    </div>
                    <div class="media-tags-adder">
                        <button type="button" :disabled="!tagToAdd" class="btn btn-primary btn-xs" @click="addTag">
                            <i class="fas fa-plus"></i> {{ $t("Add tag") }}
                        </button>
                    </div>
                </div>
                <div class="form-group" v-if="matchingTags.length > 0">
                    <button
                        v-for="mt in matchingTags"
                        :key="mt.id"
                        type="button"
                        class="btn btn-primary btn-sm btn-tag-mini"
                        @click="addTagByName(mt.name)"
                    >
                        <i class="fas fa-plus"></i> {{ mt.name }}
                    </button>
                </div>
            </div>

            <div class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn" @click="doUpload">
                    <i class="fas fa-upload"></i> {{ $t("Upload") }}
                </button>
            </div>
        </div>

        <AlbumCreateModal v-if="displayAlbumCreate" v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { parseTagName } from "@/utils/tags";
import { clone } from "@/utils/objects";
import { TagsController } from "@/control/tags";
import { AppEvents } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";

const AlbumCreateModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumCreateModal.vue"),
});

export default defineComponent({
    components: {
        AlbumCreateModal,
    },
    name: "UploadModal",
    emits: ["update:display", "upload"],
    props: {
        display: Boolean,
        inModal: Boolean,
        fixedAlbum: Number,
        files: Array,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            tags: [],
            tagToAdd: "",
            tagData: {},
            matchingTags: [],

            album: -1,
            albums: [],

            displayAlbumCreate: false,
        };
    },
    methods: {
        close: function () {
            this.$refs.modalContainer.close();
        },

        renderFiles: function (files) {
            return files.map((file) => {
                return file.name + " (" + this.renderSize(file.size) + ")";
            }).join("\n");
        },

        computeTotalSize: function (files) {
            let size = 0;

            for (let file of files) {
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
            this.tagToAdd = "";
            this.matchingTags = [];

            TagsController.Load();
            AlbumsController.Load();
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

        findTags: function () {
            const tagFilter = this.tagToAdd
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();
            if (!tagFilter) {
                this.matchingTags = [];
                return;
            }
            this.matchingTags = Object.values(this.tagData)
                .map((a: any) => {
                    const i = a.name.indexOf(tagFilter);
                    return {
                        id: a.id,
                        name: a.name,
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    if (this.tags.indexOf(a.name) >= 0) {
                        return false;
                    }
                    return a.starts || a.contains;
                })
                .sort((a, b) => {
                    if (a.starts && !b.starts) {
                        return -1;
                    } else if (b.starts && !a.starts) {
                        return 1;
                    } else if (a.name < b.name) {
                        return -1;
                    } else {
                        return 1;
                    }
                })
                .slice(0, 10);
        },

        updateTagData: function () {
            this.tagData = clone(TagsController.Tags);
            this.onTagAddChanged(true);
        },

        onTagInputKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && this.tagToAdd && !event.shiftKey) {
                this.onTagAddChanged(true);
                if (this.matchingTags.length > 0 && this.matchingTags[0].name !== this.tagToAdd) {
                    this.tagToAdd = this.matchingTags[0].name;
                    this.onTagAddChanged(true);
                    event.preventDefault();
                }
            } else if (event.key === "Enter") {
                if (this.tagToAdd) {
                    this.addTag();
                }
            }
        },

        onTagAddChanged: function (forced: boolean) {
            if (forced) {
                if (this._handles.findTagTimeout) {
                    clearTimeout(this._handles.findTagTimeout);
                    this._handles.findTagTimeout = null;
                }
                this.findTags();
            } else {
                if (this._handles.findTagTimeout) {
                    return;
                }
                this._handles.findTagTimeout = setTimeout(() => {
                    this._handles.findTagTimeout = null;
                    this.findTags();
                }, 200);
            }
        },

        removeTag: function (tag: string) {
            for (let i = 0; i < this.tags.length; i++) {
                if (this.tags[i] === tag) {
                    this.tags.splice(i, 1);
                    this.onTagAddChanged(true);
                    break;
                }
            }
        },

        addTag: function () {
            this.addTagByName(this.tagToAdd);
            this.tagToAdd = "";
            this.onTagAddChanged(true);
        },

        addTagByName: function (tag: string) {
            tag = parseTagName(tag);
            this.removeTag(tag);
            this.tags.push(tag);
            this.onTagAddChanged(true);
        },

        updateAlbums: function () {
            this.albums = AlbumsController.GetAlbumsListCopy().sort((a, b) => {
                if (a.nameLowerCase < b.nameLowerCase) {
                    return -1;
                } else {
                    return 1;
                }
            });

            if (this.inModal) {
                this.album = this.fixedAlbum;
            }
        },

        doUpload: function () {
            this.$emit("upload", this.album, this.tags);
            this.close();
        },

        createAlbum: function () {
            this.displayAlbumCreate = true;
        },

        onNewAlbum: function (albumId) {
            this.album = albumId;
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        
        this.updateTagData();
        this._handles.tagUpdateH = this.updateTagData.bind(this);
        AppEvents.AddEventListener("tags-update", this._handles.tagUpdateH);

        this.updateAlbums();
        this._handles.albumsUpdateH = this.updateAlbums.bind(this);
        AppEvents.AddEventListener("albums-update", this._handles.albumsUpdateH);

        this.reset();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("tags-update", this._handles.tagUpdateH);
        AppEvents.RemoveEventListener("albums-update", this._handles.albumsUpdateH);

        if (this._handles.findTagTimeout) {
            clearTimeout(this._handles.findTagTimeout);
            this._handles.findTagTimeout = null;
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.reset();
                this.autoFocus();
            }
        },
    },
});
</script>
