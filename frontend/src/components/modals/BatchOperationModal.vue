<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <div v-if="display" class="modal-dialog modal-xl" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Batch operation") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="batch-op-group-search">
                    <div class="form-group">
                        <label>{{ $t("Title or description must contain") }}:</label>
                        <input
                            v-model="textSearch"
                            type="text"
                            name="title-search"
                            autocomplete="off"
                            maxlength="255"
                            class="form-control form-control-full-width auto-focus"
                        />
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Media type") }}:</label>
                        <select v-model="typeSearch" class="form-control form-select form-control-full-width">
                            <option :value="0">{{ $t("Any media") }}</option>
                            <option :value="1">{{ $t("Images") }}</option>
                            <option :value="2">{{ $t("Videos") }}</option>
                            <option :value="3">{{ $t("Audios") }}</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Album") }}:</label>
                        <AlbumSelect v-model:album="albumSearch" :disabled="displayProgress"></AlbumSelect>
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Tags") }}:</label>
                        <select v-model="tagModeSearch" class="form-control form-select form-control-full-width">
                            <option :value="'all'">
                                {{ $t("Media must contain ALL of the selected tags") }}
                            </option>
                            <option :value="'any'">
                                {{ $t("Media must contain ANY of the selected tags") }}
                            </option>
                            <option :value="'none'">
                                {{ $t("Media must contain NONE of the selected tags") }}
                            </option>
                            <option :value="'untagged'">
                                {{ $t("Media must be untagged") }}
                            </option>
                        </select>
                    </div>
                    <div v-if="tagModeSearch !== 'untagged'" class="form-group media-tags">
                        <div v-for="tag in tagsSearch" :key="tag" class="media-tag">
                            <div class="media-tag-name">{{ getTagName(tag, tagVersion) }}</div>
                            <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTagSearch(tag)">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                        <div class="media-tags-finder">
                            <input
                                v-model="tagToAddSearch"
                                type="text"
                                autocomplete="off"
                                maxlength="255"
                                class="form-control tags-input-search"
                                :placeholder="$t('Search for tags') + '...'"
                                @input="onTagSearchChanged(false)"
                                @keydown="onTagSearchKeyDown"
                            />
                        </div>
                    </div>
                    <div v-if="tagModeSearch !== 'untagged' && matchingTagsSearch.length > 0" class="form-group">
                        <button
                            v-for="mt in matchingTagsSearch"
                            :key="mt.id"
                            type="button"
                            class="btn btn-primary btn-sm btn-tag-mini btn-add-tag"
                            @click="addMatchingTagSearch(mt)"
                            @keydown="onSuggestionKeydown"
                        >
                            <i class="fas fa-plus"></i> {{ mt.name }}
                        </button>
                    </div>
                </div>

                <hr />

                <div class="batch-op-group-action">
                    <div class="form-group"></div>
                    <div class="form-group">
                        <label>{{ $t("Select and action to apply") }}:</label>
                        <select
                            v-model="action"
                            class="form-control form-select form-control-full-width tags-focus-skip"
                            @keydown="onTagsSkipKeyDown"
                        >
                            <option :value="''">
                                {{ $t("--- Select an action ---") }}
                            </option>
                            <option :value="'tag-add'">
                                {{ $t("Add tags to the media assets") }}
                            </option>
                            <option :value="'tag-remove'">
                                {{ $t("Remove tags from the media assets") }}
                            </option>
                            <option :value="'album-add'">
                                {{ $t("Add media assets to album") }}
                            </option>
                            <option :value="'album-remove'">
                                {{ $t("Remove media assets from album") }}
                            </option>
                            <option :value="'delete'">
                                {{ $t("Delete media assets") }}
                            </option>
                        </select>
                    </div>

                    <div v-if="action === 'tag-add' || action === 'tag-remove'">
                        <div class="form-group media-tags">
                            <div v-for="tag in tagsAction" :key="tag" class="media-tag">
                                <div class="media-tag-name">{{ tag }}</div>
                                <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTagAction(tag)">
                                    <i class="fas fa-times"></i>
                                </button>
                            </div>
                            <div class="media-tags-finder">
                                <input
                                    v-model="tagToAddAction"
                                    type="text"
                                    autocomplete="off"
                                    maxlength="255"
                                    class="form-control tags-adder tags-input-search-action"
                                    :placeholder="$t('Add tags') + '...'"
                                    @input="onTagActionChanged(false)"
                                    @keydown="onTagActionKeyDown"
                                />
                            </div>
                            <div class="media-tags-adder">
                                <button
                                    type="button"
                                    :disabled="!tagToAddAction"
                                    class="btn btn-primary btn-xs"
                                    @click="addTagAction"
                                    @keydown="onAddTagActionKeyDown"
                                >
                                    <i class="fas fa-plus"></i> {{ $t("Add tag") }}
                                </button>
                            </div>
                        </div>
                        <div v-if="matchingTagsAction.length > 0" class="form-group">
                            <button
                                v-for="mt in matchingTagsAction"
                                :key="mt.id"
                                type="button"
                                class="btn btn-primary btn-sm btn-tag-mini btn-add-tag-action"
                                @click="addMatchingTagAction(mt)"
                                @keydown="onSuggestionActionKeydown"
                            >
                                <i class="fas fa-plus"></i> {{ mt.name }}
                            </button>
                        </div>
                    </div>

                    <div v-if="action === 'album-add' || action === 'album-remove'">
                        <div class="form-group">
                            <label>{{ $t("Album") }}:</label>
                            <AlbumSelect v-model:album="albumToAdd" :disabled="displayProgress"></AlbumSelect>
                        </div>
                    </div>
                </div>
            </div>
            <div class="modal-footer no-padding">
                <button
                    type="button"
                    class="modal-footer-btn tags-focus-skip-action"
                    :disabled="displayProgress"
                    @click="start"
                    @keydown="onTagsSkipActionKeyDown"
                >
                    <i class="fas fa-check"></i> {{ $t("Apply") }}
                </button>
            </div>
        </div>

        <BatchOperationProgressModal
            v-if="displayProgress"
            v-model:display="displayProgress"
            :status="status"
            :progress="progress"
            :error="error"
            :action-count="actionCount"
            @update:display="afterSubModalClosed"
            @cancel="cancel"
            @confirm="confirm"
        ></BatchOperationProgressModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import BatchOperationProgressModal from "./BatchOperationProgressModal.vue";
import { EVENT_NAME_TAGS_UPDATE, MatchingTag, TagsController } from "@/control/tags";
import { AppEvents } from "@/control/app-events";
import { AlbumListItemMinExt, AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { MediaController } from "@/control/media";
import { normalizeString, filterToWords, matchSearchFilter } from "@/utils/normalize";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { MediaListItem } from "@/api/models";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiAlbumsAddMediaToAlbum, apiAlbumsGetAlbum, apiAlbumsRemoveMediaFromAlbum } from "@/api/api-albums";
import { apiMediaDeleteMedia } from "@/api/api-media-edit";
import { apiTagsTagMedia, apiTagsUntagMedia } from "@/api/api-tags";
import { apiAdvancedSearch } from "@/api/api-search";
import AlbumSelect from "../utils/AlbumSelect.vue";

const PAGE_SIZE = 50;

export default defineComponent({
    name: "BatchOperationModal",
    components: {
        BatchOperationProgressModal,
        AlbumSelect,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            findTagSearchTimeout: null as ReturnType<typeof setTimeout> | null,
            findTagActionTimeout: null as ReturnType<typeof setTimeout> | null,
            batchRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            busy: false,

            tagVersion: TagsController.TagsVersion,
            albums: [] as AlbumListItemMinExt[],

            matchingTagsSearch: [] as MatchingTag[],
            tagToAddSearch: "",

            matchingTagsAction: [] as MatchingTag[],
            tagToAddAction: "",

            textSearch: "",
            typeSearch: 0,
            albumSearch: -1,
            tagsSearch: [] as number[],

            tagModeSearch: "all",

            action: "",

            tagsAction: [] as string[],

            albumToAdd: -1,

            displayProgress: false,

            progress: 0,
            status: "",
            actionCount: 0,
            actionItems: [] as number[],
            error: "",
            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.autoFocus();
                TagsController.Load();
                AlbumsController.Load();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.updateTagData();

        this.updateAlbums();

        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, this.updateAlbums.bind(this));

        if (this.display) {
            this.error = "";
            this.autoFocus();
            TagsController.Load();
            AlbumsController.Load();
        }
    },
    beforeUnmount: function () {
        abortNamedApiRequest(this.batchRequestId);

        if (this.findTagSearchTimeout) {
            clearTimeout(this.findTagSearchTimeout);
            this.findTagSearchTimeout = null;
        }

        if (this.findTagActionTimeout) {
            clearTimeout(this.findTagActionTimeout);
            this.findTagActionTimeout = null;
        }
    },
    methods: {
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

        afterSubModalClosed: function (display: boolean) {
            if (!display && this.display) {
                this.autoFocus();
            }
        },

        close: function () {
            this.closeSignal++;
        },

        getTagName: function (tag: number, v: number) {
            return TagsController.GetTagName(tag, v);
        },

        onTagSearchChanged: function (forced?: boolean) {
            if (forced) {
                if (this.findTagSearchTimeout) {
                    clearTimeout(this.findTagSearchTimeout);
                    this.findTagSearchTimeout = null;
                }
                this.findTagsSearch();
            } else {
                if (this.findTagSearchTimeout) {
                    return;
                }
                this.findTagSearchTimeout = setTimeout(() => {
                    this.findTagSearchTimeout = null;
                    this.findTagsSearch();
                }, 200);
            }
        },

        findTagsSearch: function () {
            const tagFilter = this.tagToAddSearch
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();
            this.matchingTagsSearch = Array.from(TagsController.Tags.entries())
                .map((a) => {
                    if (!tagFilter) {
                        return {
                            id: a[0],
                            name: a[1],
                            starts: true,
                            contains: true,
                        };
                    }
                    const i = a[1].indexOf(tagFilter);
                    return {
                        id: a[0],
                        name: a[1],
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    if (this.tagsSearch.indexOf(a.id) >= 0) {
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

        onTagActionChanged: function (forced?: boolean) {
            if (forced) {
                if (this.findTagActionTimeout) {
                    clearTimeout(this.findTagActionTimeout);
                    this.findTagActionTimeout = null;
                }
                this.findTagsAction();
            } else {
                if (this.findTagActionTimeout) {
                    return;
                }
                this.findTagActionTimeout = setTimeout(() => {
                    this.findTagActionTimeout = null;
                    this.findTagsAction();
                }, 200);
            }
        },

        findTagsAction: function () {
            const tagFilter = this.tagToAddAction
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();
            this.matchingTagsAction = Array.from(TagsController.Tags.entries())
                .map((a) => {
                    if (!tagFilter) {
                        return {
                            id: a[0],
                            name: a[1],
                            starts: true,
                            contains: true,
                        };
                    }
                    const i = a[1].indexOf(tagFilter);
                    return {
                        id: a[0],
                        name: a[1],
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    if (this.tagsAction.indexOf(a.name) >= 0) {
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

        removeTagSearch: function (tag: number) {
            this.tagsSearch = this.tagsSearch.filter((t) => {
                return tag !== t;
            });
            this.onTagSearchChanged(true);
            const inputElem = this.$el.querySelector(".tags-input-search");
            if (inputElem) {
                inputElem.focus();
            }
        },

        addMatchingTagSearch: function (tag) {
            if (this.tagsSearch.indexOf(tag.id) >= 0) {
                return;
            }
            this.tagsSearch.push(tag.id);
            this.onTagSearchChanged(true);
            const inputElem = this.$el.querySelector(".tags-input-search");
            if (inputElem) {
                inputElem.focus();
            }
        },

        removeTagAction: function (tag: string) {
            this.tagsAction = this.tagsAction.filter((t) => {
                return tag !== t;
            });
            this.onTagActionChanged(true);
            const inputElem = this.$el.querySelector(".tags-input-search-action");
            if (inputElem) {
                inputElem.focus();
            }
        },

        addMatchingTagAction: function (tag) {
            if (this.tagsAction.indexOf(tag.name) >= 0) {
                return;
            }
            this.tagsAction.push(tag.name);
            this.onTagActionChanged(true);
            const inputElem = this.$el.querySelector(".tags-input-search-action");
            if (inputElem) {
                inputElem.focus();
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

        onTagsSkipActionKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && event.shiftKey) {
                const inputElem = this.$el.querySelector(".tags-input-search-action");
                if (inputElem) {
                    event.preventDefault();
                    inputElem.focus();
                }
            }
        },

        onAddTagActionKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab") {
                const inputElem = this.$el.querySelector(".tags-focus-skip-action");
                if (inputElem) {
                    event.preventDefault();
                    inputElem.focus();
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
                    const inputElem = this.$el.querySelector(".tags-input-search");

                    if (inputElem) {
                        inputElem.focus();
                    }
                }
            }
        },

        onSuggestionActionKeydown: function (e: KeyboardEvent) {
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
                    const inputElem = this.$el.querySelector(".tags-input-search-action");

                    if (inputElem) {
                        inputElem.focus();
                    }
                }
            }
        },

        onTagSearchKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();

                this.onTagSearchChanged(true);

                if (this.matchingTagsSearch.length > 0) {
                    this.addMatchingTagSearch(this.matchingTagsSearch[0]);
                    this.tagToAddSearch = "";
                    this.onTagSearchChanged(true);
                }
            } else if (event.key === "Tab" && this.tagToAddSearch && !event.shiftKey) {
                if (this.matchingTagsSearch.length > 0) {
                    if (this.matchingTagsSearch[0].name !== this.tagToAddSearch) {
                        event.preventDefault();
                        this.tagToAddSearch = this.matchingTagsSearch[0].name;
                        this.onTagSearchChanged(true);
                    }
                }
            } else if (event.key === "Tab" && !event.shiftKey) {
                const toFocus = this.$el.querySelector(".tags-focus-skip");
                if (toFocus) {
                    event.preventDefault();
                    toFocus.focus();
                }
            } else if (event.key === "ArrowDown") {
                const suggestionElem = this.$el.querySelector(".btn-add-tag");
                if (suggestionElem) {
                    event.preventDefault();
                    suggestionElem.focus();
                }
            }

            event.stopPropagation();
        },

        addTagAction: function () {
            if (!this.tagToAddAction) {
                return;
            }
            this.addMatchingTagAction({ name: this.tagToAddAction, id: -1 });
            this.tagToAddAction = "";
            this.onTagActionChanged(true);
        },

        onTagActionKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();

                this.addTagAction();
            } else if (event.key === "Tab" && this.tagToAddSearch && !event.shiftKey) {
                this.onTagActionChanged(true);

                if (this.matchingTagsAction.length > 0) {
                    if (this.tagToAddAction !== this.matchingTagsAction[0].name) {
                        event.preventDefault();
                        this.tagToAddAction = this.matchingTagsAction[0].name;
                        this.onTagActionChanged(true);
                    }
                }
            } else if (event.key === "Tab" && !this.tagToAddAction && !event.shiftKey) {
                const toFocus = this.$el.querySelector(".tags-focus-skip-action");
                if (toFocus) {
                    event.preventDefault();
                    toFocus.focus();
                }
            } else if (event.key === "ArrowDown") {
                const suggestionElem = this.$el.querySelector(".btn-add-tag-action");
                if (suggestionElem) {
                    event.preventDefault();
                    suggestionElem.focus();
                }
            }
        },

        updateTagData: function () {
            this.tagVersion = TagsController.TagsVersion;
            this.onTagSearchChanged(false);
            this.onTagActionChanged(false);
        },

        updateAlbums: function () {
            this.albums = AlbumsController.GetAlbumsListMin().sort((a, b) => {
                if (a.nameLowerCase < b.nameLowerCase) {
                    return -1;
                } else {
                    return 1;
                }
            });
        },

        start: function () {
            if (!this.action) {
                return;
            }

            switch (this.action) {
                case "tag-add":
                case "tag-remove":
                    if (this.tagsAction.length === 0) {
                        return;
                    }
                    break;
                case "album-add":
                case "album-remove":
                    if (this.albumToAdd < 0) {
                        return;
                    }
                    break;
            }

            this.displayProgress = true;
            this.status = "search";
            this.progress = 0;
            this.actionItems = [];

            if (this.albumSearch >= 0) {
                this.loadAlbumSearch();
            } else {
                this.searchNext(null);
            }
        },

        loadAlbumSearch: function () {
            abortNamedApiRequest(this.batchRequestId);

            makeNamedApiRequest(this.batchRequestId, apiAlbumsGetAlbum(this.albumSearch))
                .onSuccess((result) => {
                    this.filterElements(result.list);
                    this.finishSearch();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            this.status = "error";
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        notFound: () => {
                            this.status = "error";
                            this.error = this.$t("The selected album was not found");
                        },
                        serverError: () => {
                            this.status = "error";
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.status = "error";
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    this.status = "error";
                    this.error = err.message;
                });
        },

        getTagMode: function (): "allof" | "anyof" | "noneof" {
            switch (this.tagModeSearch) {
                case "any":
                    if (this.tagsSearch.length > 16) {
                        return "allof";
                    }
                    return "anyof";
                case "none":
                    return "noneof";
                default:
                    return "allof";
            }
        },

        getTagList: function (): string[] {
            if (this.tagModeSearch === "untagged") {
                return [];
            }
            if (this.tagModeSearch === "any" && this.tagsSearch.length > 16) {
                return [];
            }
            return this.tagsSearch
                .map((tag) => {
                    return this.getTagName(tag, this.tagVersion);
                })
                .slice(0, 16);
        },

        searchNext: function (continueRef: number | null) {
            abortNamedApiRequest(this.batchRequestId);

            makeNamedApiRequest(this.batchRequestId, apiAdvancedSearch(this.getTagMode(), this.getTagList(), "asc", continueRef, PAGE_SIZE))
                .onSuccess((result) => {
                    this.filterElements(result.items);

                    this.progress = (Math.max(0, result.scanned) / Math.max(1, result.total_count)) * 100;

                    if (result.scanned >= result.total_count) {
                        // Finished
                        this.finishSearch();
                    } else {
                        this.searchNext(result.continue);
                    }
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            this.status = "error";
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        serverError: () => {
                            this.status = "error";
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.status = "error";
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    this.status = "error";
                    this.error = err.message;
                });
        },

        finishSearch: function () {
            if (this.actionItems.length > 0) {
                this.status = this.action === "delete" ? "confirmation-delete" : "confirmation";
                this.actionCount = this.actionItems.length;
            } else {
                this.status = "error";
                this.error = this.$t("No items found matching the specified criteria");
            }
        },

        filterElements: function (results: MediaListItem[]) {
            const filterText = normalizeString(this.textSearch).trim().toLowerCase();
            const filterTextWords = filterToWords(filterText);
            const filterType = this.typeSearch;
            const filterTags = this.tagsSearch.slice();
            const filterTagMode = this.tagModeSearch;

            for (const e of results) {
                if (filterText) {
                    if (
                        matchSearchFilter(e.title, filterText, filterTextWords) < 0 &&
                        matchSearchFilter(e.description, filterText, filterTextWords) < 0
                    ) {
                        continue;
                    }
                }

                if (filterType) {
                    if (e.type !== filterType) {
                        continue;
                    }
                }

                if (filterTagMode === "all") {
                    if (filterTags.length > 0) {
                        let passesTags = true;
                        for (const tag of filterTags) {
                            if (!e.tags || !e.tags.includes(tag)) {
                                passesTags = false;
                                break;
                            }
                        }

                        if (!passesTags) {
                            continue;
                        }
                    }
                } else if (filterTagMode === "none") {
                    if (filterTags.length > 0) {
                        let passesTags = true;
                        for (const tag of filterTags) {
                            if (e.tags && e.tags.includes(tag)) {
                                passesTags = false;
                                break;
                            }
                        }

                        if (!passesTags) {
                            continue;
                        }
                    }
                } else if (filterTagMode === "any") {
                    if (filterTags.length > 0) {
                        let passesTags = false;
                        for (const tag of filterTags) {
                            if (!e.tags || e.tags.includes(tag)) {
                                passesTags = true;
                                break;
                            }
                        }

                        if (!passesTags) {
                            continue;
                        }
                    }
                } else if (filterTagMode === "untagged") {
                    if (e.tags && e.tags.length > 0) {
                        continue;
                    }
                }

                this.actionItems.push(e.id);
            }
        },

        cancel: function () {
            abortNamedApiRequest(this.batchRequestId);
        },

        confirm: function () {
            this.status = "action";
            this.progress = 0;
            this.actionNext(0);
        },

        actionNext: function (i: number) {
            abortNamedApiRequest(this.batchRequestId);

            if (i >= this.actionItems.length) {
                // Finish
                this.status = "success";

                AlbumsController.LoadCurrentAlbum();
                MediaController.Load();
                TagsController.Load();
                return;
            }

            const mediaId = this.actionItems[i];

            switch (this.action) {
                case "tag-add":
                    this.actionAddTag(mediaId, this.tagsAction.slice(), i + 1);
                    break;
                case "tag-remove":
                    this.actionRemoveTag(mediaId, this.tagsAction.slice(), i + 1);
                    break;
                case "album-add":
                    this.actionAddAlbum(mediaId, i + 1);
                    break;
                case "album-remove":
                    this.actionRemoveAlbum(mediaId, i + 1);
                    break;
                case "delete":
                    this.actionDelete(mediaId, i + 1);
                    break;
            }

            this.progress = ((i + 1) * 100) / (this.actionItems.length || 1);
        },

        actionDelete: function (mid: number, next: number) {
            makeNamedApiRequest(this.batchRequestId, apiMediaDeleteMedia(mid))
                .onSuccess(() => {
                    this.actionNext(next);
                })
                .onRequestError((err, handleErr) => {
                    this.status = "error";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        notFound: () => {
                            this.error = this.$t("Not found");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.status = "error";
                });
        },

        actionAddAlbum: function (mid: number, next: number) {
            makeNamedApiRequest(this.batchRequestId, apiAlbumsAddMediaToAlbum(this.albumToAdd, mid))
                .onSuccess(() => {
                    this.actionNext(next);
                })
                .onRequestError((err, handleErr) => {
                    this.status = "error";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        maxSizeReached: () => {
                            this.error = this.$t("The album reached the limit of 1024 elements. Please, consider creating another album.");
                        },
                        badRequest: () => {
                            this.error = this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        notFound: () => {
                            this.error = this.$t("Not found");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.status = "error";
                });
        },

        actionRemoveAlbum: function (mid: number, next: number) {
            makeNamedApiRequest(this.batchRequestId, apiAlbumsRemoveMediaFromAlbum(this.albumToAdd, mid))
                .onSuccess(() => {
                    this.actionNext(next);
                })
                .onRequestError((err, handleErr) => {
                    this.status = "error";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        notFound: () => {
                            this.error = this.$t("Not found");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.status = "error";
                });
        },

        actionAddTag: function (mid: number, tags: string[], next: number) {
            if (tags.length === 0) {
                this.actionNext(next);
                return;
            }

            makeNamedApiRequest(this.batchRequestId, apiTagsTagMedia(mid, tags[0]))
                .onSuccess(() => {
                    this.actionAddTag(mid, tags.slice(1), next);
                })
                .onRequestError((err, handleErr) => {
                    this.status = "error";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTagName: () => {
                            this.error = this.$t("Invalid tag name");
                        },
                        badRequest: () => {
                            this.error = this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.status = "error";
                });
        },

        actionRemoveTag: function (mid: number, tags: string[], next: number) {
            if (tags.length === 0) {
                this.actionNext(next);
                return;
            }

            const tagId = TagsController.FindTag(tags[0]);

            if (tagId < 0) {
                // Tag not found
                this.actionRemoveTag(mid, tags.slice(1), next);
                return;
            }

            makeNamedApiRequest(this.batchRequestId, apiTagsUntagMedia(mid, tagId))
                .onSuccess(() => {
                    this.actionRemoveTag(mid, tags.slice(1), next);
                })
                .onRequestError((err, handleErr) => {
                    this.status = "error";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.status = "error";
                });
        },
    },
});
</script>
