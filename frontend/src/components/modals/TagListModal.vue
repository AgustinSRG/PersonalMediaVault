<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" :lock-close="busy" @close="onClose">
        <div v-if="display" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Tags") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body">
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

            <div class="modal-footer no-padding">
                <button type="button" @click="close" :disabled="busy" class="modal-footer-btn">
                    <i class="fas fa-check"></i> {{ $t("Done") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { MediaController } from "@/control/media";
import { TagsController } from "@/control/tags";
import { TagsAPI } from "@/api/api-tags";

export default defineComponent({
    components: {},
    name: "TagListModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
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

            changed: false,
        };
    },
    methods: {
        load: function () {
            if (!MediaController.MediaData) {
                return;
            }
            this.tags = (MediaController.MediaData.tags || []).slice();
            this.onTagAddChanged();
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

        close: function () {
            this.$refs.modalContainer.close();
        },

        onClose: function () {
            if (this.changed) {
                MediaController.Load();
            }
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

        removeTag: function (tag) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const tagName = this.getTagName(tag, this.tagVersion);

            Request.Pending("media-editor-busy", TagsAPI.UntagMedia(mediaId, tag))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Removed tag") + ": " + tagName);
                    this.busy = false;
                    for (let i = 0; i < this.tags.length; i++) {
                        if (this.tags[i] === tag) {
                            this.tags.splice(i, 1);
                            break;
                        }
                    }
                    this.changed = true;
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
                            AppEvents.Emit("unauthorized");
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

            Request.Pending("media-editor-busy", TagsAPI.TagMedia(mediaId, tag))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    this.tagToAdd = "";
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    this.changed = true;
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
                            AppEvents.Emit("unauthorized");
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

            Request.Pending("media-editor-busy", TagsAPI.TagMedia(mediaId, tag))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    this.changed = true;
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
                            AppEvents.Emit("unauthorized");
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
            if (!tagFilter) {
                this.matchingTags = Array.from(TagsController.Tags.entries())
                    .map(a => {
                        return {
                            id: a[0],
                            name: a[1],
                        };
                    })
                    .filter((a) => {
                        if (this.tags.indexOf(a.id) >= 0) {
                            return false;
                        }
                        return true;
                    })
                    .sort((a, b) => {
                        if (a.name < b.name) {
                            return -1;
                        } else {
                            return 1;
                        }
                    })
                    .slice(0, 10);
            }
            this.matchingTags = Array.from(TagsController.Tags.entries())
                .map(a => {
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
        AppEvents.AddEventListener("tags-update", this._handles.tagUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this._handles.mediaUpdateH);

        this.updateTagData();
        this.load();

        if (this.display) {
            this.autoFocus();
            TagsController.Load();
        }
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("tags-update", this._handles.tagUpdateH);
        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);
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
