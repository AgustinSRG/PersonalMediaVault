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
                <div class="form-group media-tags">
                    <div v-for="tag in tags" :key="tag" class="media-tag">
                        <div class="media-tag-name">{{ tag }}</div>
                        <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTag(tag)">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="media-tags-finder">
                        <input
                            v-model="tagToAdd"
                            type="text"
                            autocomplete="off"
                            maxlength="255"
                            class="form-control tag-to-add auto-focus"
                            :placeholder="$t('Add tags') + '...'"
                            @input="onTagAddChanged(false)"
                            @keydown="onTagInputKeyDown"
                        />
                    </div>
                    <div class="media-tags-adder">
                        <button type="button" :disabled="!tagToAdd" class="btn btn-primary btn-xs" @click="addTag">
                            <i class="fas fa-plus"></i> {{ $t("Add tag") }}
                        </button>
                    </div>
                </div>
                <div v-if="matchingTags.length > 0" class="form-group">
                    <button
                        v-for="mt in matchingTags"
                        :key="mt.id"
                        type="button"
                        class="btn btn-primary btn-sm btn-tag-mini btn-add-tag"
                        @click="addTagAndFocusInput(mt.name)"
                        @keydown="onSuggestionKeydown"
                    >
                        <i class="fas fa-plus"></i> {{ mt.name }}
                    </button>
                </div>
            </div>

            <div class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn" @click="doUpload"><i class="fas fa-upload"></i> {{ $t("Upload") }}</button>
            </div>
        </div>

        <AlbumCreateModal v-if="displayAlbumCreate" v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineAsyncComponent, defineComponent, nextTick, PropType } from "vue";
import { useVModel } from "../../utils/v-model";
import { parseTagName } from "@/utils/tags";
import { EVENT_NAME_TAGS_UPDATE, MatchingTag, TagsController } from "@/control/tags";
import { AlbumListItemMinExt, AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import { getLastUsedTags } from "@/control/app-preferences";
import AlbumSelect from "../utils/AlbumSelect.vue";

const AlbumCreateModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumCreateModal.vue"),
});

// Max number of tag suggestions
const TAGS_SUGGESTION_LIMIT = 10;

export default defineComponent({
    name: "UploadModal",
    components: {
        AlbumCreateModal,
        AlbumSelect,
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
            tagToAdd: "",
            tagVersion: TagsController.TagsVersion,
            matchingTags: [] as MatchingTag[],

            album: -1,
            albums: [] as AlbumListItemMinExt[],

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
        this.updateTagData();

        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.updateAlbums();

        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, this.updateAlbums.bind(this));

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

            const lastUsedTagsIds = getLastUsedTags();

            if (!tagFilter) {
                const lastUsedTags: MatchingTag[] = [];
                const addedTagIds: number[] = [];

                for (const tid of lastUsedTagsIds) {
                    const tagName = TagsController.Tags.get(tid);

                    if (TagsController.Tags.has(tid) && !this.tags.includes(tagName) && !addedTagIds.includes(tid)) {
                        lastUsedTags.push({
                            id: tid,
                            name: tagName,
                        });

                        addedTagIds.push(tid);

                        if (lastUsedTags.length >= TAGS_SUGGESTION_LIMIT) {
                            break;
                        }
                    }
                }

                if (lastUsedTags.length < TAGS_SUGGESTION_LIMIT) {
                    Array.from(TagsController.Tags.entries())
                        .filter((t) => {
                            return !this.tags.includes(t[1]) && !addedTagIds.includes(t[0]);
                        })
                        .sort((a, b) => {
                            if (a[1] < b[1]) {
                                return -1;
                            } else {
                                return 1;
                            }
                        })
                        .slice(0, TAGS_SUGGESTION_LIMIT - lastUsedTags.length)
                        .forEach((t) => {
                            lastUsedTags.push({
                                id: t[0],
                                name: t[1],
                            });
                        });
                }

                this.matchingTags = lastUsedTags;

                return;
            }
            this.matchingTags = Array.from(TagsController.Tags.entries())
                .map((a) => {
                    const i = a[1].indexOf(tagFilter);
                    const lastUsedIndex = lastUsedTagsIds.indexOf(a[0]);
                    return {
                        id: a[0],
                        name: a[1],
                        starts: i === 0,
                        contains: i >= 0,
                        lastUsed: lastUsedIndex === -1 ? lastUsedTagsIds.length : lastUsedIndex,
                    };
                })
                .filter((a) => {
                    if (this.tags.includes(a.name)) {
                        return false;
                    }
                    return a.starts || a.contains;
                })
                .sort((a, b) => {
                    if (a.starts && !b.starts) {
                        return -1;
                    } else if (b.starts && !a.starts) {
                        return 1;
                    } else if (a.lastUsed < b.lastUsed) {
                        return -1;
                    } else if (a.lastUsed > b.lastUsed) {
                        return 1;
                    } else if (a.name < b.name) {
                        return -1;
                    } else {
                        return 1;
                    }
                })
                .slice(0, TAGS_SUGGESTION_LIMIT);
        },

        updateTagData: function () {
            this.tagVersion = TagsController.TagsVersion;
            this.onTagAddChanged(true);
        },

        onTagInputKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && !event.shiftKey) {
                if (this.tagToAdd) {
                    this.onTagAddChanged(true);
                    if (this.matchingTags.length > 0 && this.matchingTags[0].name !== this.tagToAdd) {
                        this.tagToAdd = this.matchingTags[0].name;
                        this.onTagAddChanged(true);
                        event.preventDefault();
                    }
                } else {
                    const btn = this.$el.querySelector(".modal-footer-btn");

                    if (btn) {
                        event.preventDefault();
                        btn.focus();
                    }
                }
            } else if (event.key === "Enter") {
                if (this.tagToAdd) {
                    this.addTag();
                }
            } else if (event.key === "ArrowDown") {
                const suggestionElem = this.$el.querySelector(".btn-add-tag");
                if (suggestionElem) {
                    suggestionElem.focus();
                }
            }
        },

        onSuggestionKeydown: function (e: KeyboardEvent) {
            if (e.key === "ArrowRight" || e.key === "ArrowDown") {
                e.preventDefault();
                e.stopPropagation();

                const nextElem = (e.target as HTMLElement).nextSibling as HTMLElement;

                if (nextElem && nextElem.focus) {
                    nextElem.focus();
                }
            } else if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
                e.preventDefault();
                e.stopPropagation();

                const prevElem = (e.target as HTMLElement).previousSibling as HTMLElement;

                if (prevElem && prevElem.focus) {
                    prevElem.focus();
                } else {
                    const inputElem = this.$el.querySelector(".tag-to-add");

                    if (inputElem) {
                        inputElem.focus();
                    }
                }
            }
        },

        onTagAddChanged: function (forced: boolean) {
            if (forced) {
                if (this.findTagTimeout) {
                    clearTimeout(this.findTagTimeout);
                    this.findTagTimeout = null;
                }
                this.findTags();
            } else {
                if (this.findTagTimeout) {
                    return;
                }
                this.findTagTimeout = setTimeout(() => {
                    this.findTagTimeout = null;
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

            const inputElem = this.$el.querySelector(".tag-to-add");

            if (inputElem) {
                inputElem.focus();
            }
        },

        addTag: function () {
            this.addTagByName(this.tagToAdd);
            this.tagToAdd = "";
            this.onTagAddChanged(true);

            const inputElement = this.$el.querySelector(".tag-to-add");

            if (inputElement) {
                inputElement.focus();
            }
        },

        addTagAndFocusInput: function (tag: string) {
            this.addTagByName(tag);

            const inputElement = this.$el.querySelector(".tag-to-add");

            if (inputElement) {
                inputElement.focus();
                inputElement.select();
            }
        },

        addTagByName: function (tag: string) {
            tag = parseTagName(tag);
            this.removeTag(tag);
            this.tags.push(tag);
            this.onTagAddChanged(true);
        },

        updateAlbums: function () {
            this.albums = AlbumsController.GetAlbumsListMin().sort((a, b) => {
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

        onNewAlbum: function (albumId: number) {
            this.album = albumId;
        },
    },
});
</script>
