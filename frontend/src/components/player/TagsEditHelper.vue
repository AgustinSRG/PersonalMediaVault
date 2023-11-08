<template>
    <div class="tags-edit-helper-container">
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
                            class="btn btn-primary btn-sm btn-tag-mini"
                            :disabled="busy"
                            @click="addMatchingTag(mt.name)"
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
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { AppStatus } from "@/control/app-status";
import { TagsController } from "@/control/tags";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { TagsAPI } from "@/api/api-tags";
import { MediaController } from "@/control/media";

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

            Request.Pending("media-editor-busy", TagsAPI.UntagMedia(mediaId, tag))
                .onSuccess(({ removed }) => {
                    AppEvents.Emit("snack", this.$t("Removed tag") + ": " + tagName);
                    this.busy = false;
                    for (let i = 0; i < this.tags.length; i++) {
                        if (this.tags[i] === tag) {
                            this.tags.splice(i, 1);
                            break;
                        }
                    }
                    if (removed) {
                        TagsController.RemoveTag(tag);
                    }
                    if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                        if (!MediaController.MediaData.tags.includes(tag)) {
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
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid tag name"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        addTag: function (e) {
            if (e) {
                e.preventDefault();
            }
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const tag = this.tagToAdd;

            Request.Pending("tags-editor-busy", TagsAPI.TagMedia(mediaId, tag))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
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
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid tag name"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
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

            Request.Pending("tags-editor-busy", TagsAPI.TagMedia(mediaId, tag))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
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
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid tag name"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        onTagAddChanged: function () {
            if (this._handles.findTagTimeout) {
                return;
            }
            this._handles.findTagTimeout = setTimeout(() => {
                this._handles.findTagTimeout = null;
                this.findTags();
            }, 200);
        },

        findTags: function () {
            const tagFilter = this.tagToAdd
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();
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
                    if (this.matchingTags[0].name !== this.tagToAdd) {
                        e.preventDefault();
                        this.tagToAdd = this.matchingTags[0].name;
                    }
                }
            }
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.tagUpdateH = this.updateTagData.bind(this);
        TagsController.AddEventListener(this._handles.tagUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AuthController.AddChangeEventListener(this._handles.authUpdateH);

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        MediaController.AddUpdateEventListener(this._handles.mediaUpdateH);

        this.updateTagData();
        this.load();

        if (this.display) {
            this.autoFocus();
            TagsController.Load();
        }
    },
    beforeUnmount: function () {
        Request.Abort("tags-editor-busy");
        TagsController.RemoveEventListener(this._handles.tagUpdateH);
        AuthController.RemoveChangeEventListener(this._handles.authUpdateH);
        MediaController.RemoveUpdateEventListener(this._handles.mediaUpdateH);
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

<style>
@import "@/style/player/tags-edit.css";
</style>
