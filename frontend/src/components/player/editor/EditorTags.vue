<template>
    <div class="player-editor-sub-content">
        <!--- Tags -->

        <div class="form-group">
            <label>{{ $t("Tags") }}:</label>
        </div>
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
        <form v-if="canWrite" @submit="addTag">
            <div class="form-group">
                <label>{{ $t("Tag to add") }}:</label>
                <input
                    v-model="tagToAdd"
                    type="text"
                    autocomplete="off"
                    maxlength="255"
                    :disabled="busy"
                    class="form-control tag-to-add auto-focus"
                    @input="onTagAddChanged"
                    @keydown="onTagAddKeyDown"
                />
            </div>
            <div v-if="matchingTags.length > 0" class="form-group">
                <button
                    v-for="mt in matchingTags"
                    :key="mt.id"
                    type="button"
                    class="btn btn-primary btn-sm btn-tag-mini"
                    :disabled="busy"
                    @click="addMatchingTag(mt.name)"
                >
                    <i class="fas fa-plus"></i> {{ mt.name }}
                </button>
            </div>
            <div class="form-group">
                <button type="submit" class="btn btn-primary" :disabled="busy || !tagToAdd">
                    <i class="fas fa-plus"></i> {{ $t("Add Tag") }}
                </button>
            </div>
            <div v-if="busy" class="form-group loader-delayed-custom">
                <label><i class="fa fa-spinner fa-spin mr-1"></i> {{ $t("Saving changes") }}...</label>
            </div>
        </form>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { getLastUsedTags, setLastUsedTag } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { EVENT_NAME_TAGS_UPDATE, TagsController } from "@/control/tags";
import { clone } from "@/utils/objects";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { getUniqueStringId } from "@/utils/unique-id";
import { defineComponent, nextTick } from "vue";
import { PagesController } from "@/control/pages";
import { apiTagsTagMedia, apiTagsUntagMedia } from "@/api/api-tags";

export default defineComponent({
    name: "EditorTags",
    components: {},
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
            findTagTimeout: null,
        };
    },
    data: function () {
        return {
            type: 0,

            tags: [],
            tagToAdd: "",
            tagVersion: TagsController.TagsVersion,
            matchingTags: [],

            busy: false,

            canWrite: AuthController.CanWrite,
        };
    },

    mounted: function () {
        this.updateMediaData();
        this.updateTagData();

        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        TagsController.Load();

        this.autoFocus();
    },

    beforeUnmount: function () {
        if (this.findTagTimeout) {
            clearTimeout(this.findTagTimeout);
        }

        abortNamedApiRequest(this.requestId);
    },

    methods: {
        autoFocus: function () {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.tags = (MediaController.MediaData.tags || []).slice();
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
                    PagesController.ShowSnackBarRight(this.$t("Removed tag") + ": " + tagName);
                    this.busy = false;
                    for (let i = 0; i < this.tags.length; i++) {
                        if (this.tags[i] === tag) {
                            this.tags.splice(i, 1);
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.tags = clone(this.tags);
                    }
                    if (removed) {
                        TagsController.RemoveTag(tag);
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
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
                    PagesController.ShowSnackBarRight(this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    this.tagToAdd = "";
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    if (MediaController.MediaData) {
                        MediaController.MediaData.tags = clone(this.tags);
                    }
                    this.$emit("changed");
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
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTagName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid tag name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        addMatchingTag: function (tag) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiTagsTagMedia(mediaId, tag))
                .onSuccess((res) => {
                    setLastUsedTag(res.id);
                    PagesController.ShowSnackBarRight(this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    if (MediaController.MediaData) {
                        MediaController.MediaData.tags = clone(this.tags);
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidTagName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid tag name"));
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
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
            if (!tagFilter) {
                const lastUsedTagsIds = getLastUsedTags();
                const lastUsedTags = [];

                for (const tid of lastUsedTagsIds) {
                    if (TagsController.Tags.has(tid) && !this.tags.includes(tid)) {
                        lastUsedTags.push({
                            id: tid,
                            name: TagsController.Tags.get(tid),
                        });
                    }
                }

                this.matchingTags = lastUsedTags;

                return;
            }
            this.matchingTags = Array.from(TagsController.Tags.entries())
                .map((a) => {
                    const i = a[1].indexOf(tagFilter);
                    return {
                        id: a[0],
                        name: a[1],
                        starts: i === 0,
                        contains: i >= 0,
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
                    } else if (a.name < b.name) {
                        return -1;
                    } else {
                        return 1;
                    }
                })
                .slice(0, 10);
        },

        onTagAddKeyDown: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();
                this.addTag();
            } else if (e.key === "Tab") {
                this.findTags();
                if (this.matchingTags.length > 0) {
                    if (this.matchingTags[0].name != this.tagToAdd) {
                        e.preventDefault();
                        this.tagToAdd = this.matchingTags[0].name;
                    }
                }
            }
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },
});
</script>
