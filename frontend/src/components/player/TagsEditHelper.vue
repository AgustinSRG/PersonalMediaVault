<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            :title="$t('Tags')"
            v-model:display="displayStatus"
            :contextOpen="contextOpen"
            :position-key="'tags-edit-helper-pos'"
            @clicked="propagateClick"
        >
            <div class="tags-editor-body">
                <div class="form-group media-tags">
                    <label v-if="tags.length === 0">{{ $t("There are no tags yet for this media.") }}</label>
                    <div v-for="tag in tags" :key="tag" class="media-tag">
                        <div class="media-tag-name">{{ getTagName(tag, tagVersion) }}</div>
                        <button
                            v-if="canWrite"
                            type="button"
                            :title="$t('Remove tag')"
                            class="media-tag-btn"
                            :disabled="busy"
                            @click="removeTag(tag)"
                        >
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                </div>
                <div v-if="canWrite">
                    <div class="form-group media-tags">
                        <div class="media-tags-finder">
                            <input
                                type="text"
                                autocomplete="off"
                                maxlength="255"
                                v-model="tagToAdd"
                                :disabled="busy"
                                @input="onTagAddChanged"
                                @keydown="onTagAddKeyDown"
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
                            class="btn btn-primary btn-sm btn-tag-mini btn-add-tag"
                            :disabled="busy"
                            @click="addMatchingTag(mt.name)"
                            @keydown="onSuggestionKeydown"
                        >
                            <i class="fas fa-plus"></i> {{ mt.name }}
                        </button>
                    </div>
                </div>
            </div>
        </ResizableWidget>
    </div>
</template>

<script lang="ts">
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

import ResizableWidget from "@/components/player/ResizableWidget.vue";
import { nextTick } from "vue";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { AppStatus } from "@/control/app-status";
import { EVENT_NAME_TAGS_UPDATE, TagsController } from "@/control/tags";
import { AppEvents } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getLastUsedTags, setLastUsedTag } from "@/control/app-preferences";
import { getUniqueStringId } from "@/utils/unique-id";
import { EVENT_NAME_GO_NEXT, EVENT_NAME_GO_PREV, PagesController } from "@/control/pages";
import { apiTagsTagMedia, apiTagsUntagMedia } from "@/api/api-tags";

export default defineComponent({
    components: {
        ResizableWidget,
    },
    name: "TagsEditHelper",
    emits: ["update:display", "tags-update", "clicked"],
    props: {
        display: Boolean,
        contextOpen: Boolean,
        currentTime: Number,
    },
    setup(props) {
        return {
            requestId: getUniqueStringId(),
            findTagTimeout: null,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            mid: AppStatus.CurrentMedia,

            tags: [],
            tagToAdd: "",
            tagVersion: TagsController.TagsVersion,
            matchingTags: [],

            loading: true,
            busy: false,

            canWrite: AuthController.CanWrite,
        };
    },
    methods: {
        propagateClick: function () {
            this.$emit("clicked");
        },

        close: function () {
            this.displayStatus = false;
        },

        load: function () {
            if (!MediaController.MediaData) {
                return;
            }
            this.tags = (MediaController.MediaData.tags || []).slice();
            this.onTagAddChanged();
            this.autoFocus();
        },

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                } else {
                    this.$el.focus();
                }
            });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        updateMediaData: function () {
            this.mid = AppStatus.CurrentMedia;
            this.load();
        },

        updateTagData: function () {
            this.tagVersion = TagsController.TagsVersion;
            this.findTags();
        },

        getTagName: function (tag: number, v: number) {
            return TagsController.GetTagName(tag, v);
        },

        removeTag: function (tag: number) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const tagName = this.getTagName(tag, this.tagVersion);

            makeNamedApiRequest(this.requestId, apiTagsUntagMedia(mediaId, tag))
                .onSuccess(({ removed }) => {
                    PagesController.ShowSnackBar(this.$t("Removed tag") + ": " + tagName);
                    this.busy = false;
                    for (let i = 0; i < this.tags.length; i++) {
                        if (this.tags[i] === tag) {
                            this.tags.splice(i, 1);
                            break;
                        }
                    }
                    if (removed) {
                        TagsController.RemoveTag(tag);
                    } else {
                        this.findTags();
                    }
                    if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                        if (MediaController.MediaData.tags.includes(tag)) {
                            MediaController.MediaData.tags = MediaController.MediaData.tags.filter((t) => {
                                return t !== tag;
                            });
                        }
                    }
                    this.$emit("tags-update");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBar(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        addTag: function (e?: Event) {
            if (e) {
                e.preventDefault();
            }
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const tag = this.tagToAdd;

            makeNamedApiRequest(this.requestId, apiTagsTagMedia(mediaId, tag))
                .onSuccess((res) => {
                    setLastUsedTag(res.id);
                    PagesController.ShowSnackBar(this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    this.tagToAdd = "";
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                        if (!MediaController.MediaData.tags.includes(res.id)) {
                            MediaController.MediaData.tags.push(res.id);
                        }
                    }
                    this.$emit("tags-update");
                    nextTick(() => {
                        const elemFocus = this.$el.querySelector(".tag-to-add");

                        if (elemFocus) {
                            elemFocus.focus();
                        }
                    });
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTagName: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid tag name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBar(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        addMatchingTag: function (tag: string) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiTagsTagMedia(mediaId, tag))
                .onSuccess((res) => {
                    setLastUsedTag(res.id);
                    PagesController.ShowSnackBar(this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                        if (!MediaController.MediaData.tags.includes(res.id)) {
                            MediaController.MediaData.tags.push(res.id);
                        }
                    }
                    this.$emit("tags-update");
                    this.autoFocus();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTagName: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid tag name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBar(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        onTagAddChanged: function () {
            if (this.findTagTimeout) {
                return;
            }
            this.findTagTimeout = setTimeout(() => {
                this.findTagTimeout = null;
                this.findTags();
            }, 200);
        },

        findTags: function () {
            const tagFilter = this.tagToAdd
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();

            const lastUsedTagsIds = getLastUsedTags();

            if (!tagFilter) {
                const lastUsedTags = [];

                for (const tid of lastUsedTagsIds) {
                    if (TagsController.Tags.has(tid) && !this.tags.includes(tid)) {
                        lastUsedTags.push({
                            id: tid,
                            name: TagsController.Tags.get(tid),
                        });
                    }
                }

                if (lastUsedTags.length < 10) {
                    Array.from(TagsController.Tags.entries())
                        .filter((t) => {
                            return !this.tags.includes(t[0]) && !lastUsedTags.includes(t[0]);
                        })
                        .slice(0, 10 - lastUsedTags.length)
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
                    if (this.tags.indexOf(a.id) >= 0) {
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
                .slice(0, 10);
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

        onTagAddKeyDown: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();
                this.addTag();
            } else if (e.key === "Tab" && !e.shiftKey) {
                this.findTags();
                if (this.matchingTags.length > 0) {
                    if (this.matchingTags[0].name !== this.tagToAdd) {
                        e.preventDefault();
                        this.tagToAdd = this.matchingTags[0].name;
                    }
                }
            } else if (e.key === "ArrowRight") {
                if (!this.tagToAdd) {
                    AppEvents.Emit(EVENT_NAME_GO_NEXT);
                }
            } else if (e.key === "PageDown") {
                AppEvents.Emit(EVENT_NAME_GO_NEXT);
            } else if (e.key === "ArrowLeft") {
                if (!this.tagToAdd) {
                    AppEvents.Emit(EVENT_NAME_GO_PREV);
                }
            } else if (e.key === "PageUp") {
                AppEvents.Emit(EVENT_NAME_GO_PREV);
            } else if (e.key === "ArrowDown") {
                const suggestionElem = this.$el.querySelector(".btn-add-tag");
                if (suggestionElem) {
                    suggestionElem.focus();
                }
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));

        this.load();

        if (this.display) {
            this.autoFocus();
            TagsController.Load();
        }
    },
    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
                this.load();
                TagsController.Load();
            }
        },
    },
});
</script>
