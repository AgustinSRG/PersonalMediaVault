<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @click="close" @keydown="keyDownHandle">
    <div v-if="display" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Tags") }}</div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group media-tags">
          <label v-if="tags.length === 0">{{
            $t("There are no tags yet for this media.")
          }}</label>
          <div v-for="tag in tags" :key="tag" class="media-tag">
            <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
            <button v-if="canWrite" type="button" :title="$t('Remove tag')" class="media-tag-btn" :disabled="busy" @click="removeTag(tag)">
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
        <div v-if="canWrite">
          <div class="form-group">
            <label>{{ $t("Tag to add") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" v-model="tagToAdd" :disabled="busy" @input="onTagAddChanged" @keydown="onTagAddKeyDown" class="form-control tag-to-add auto-focus" />
          </div>
          <div class="form-group" v-if="matchingTags.length > 0">
            <button v-for="mt in matchingTags" :key="mt.id" type="button" class="btn btn-primary btn-sm btn-tag-mini" :disabled="busy" @click="addMatchingTag(mt.name)">
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
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import { MediaController } from "@/control/media";
import { TagsController } from "@/control/tags";
import { TagsAPI } from "@/api/api-tags";
import { clone } from "@/utils/objects";

export default defineComponent({
  components: {
  },
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
      tagData: {},
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
      if (this.busy) {
        return;
      }
      this.displayStatus = false;
      if (this.changed) {
        MediaController.Load();
      }
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    updateMediaData: function () {
      this.mid = AppStatus.CurrentMedia;
      this.load();
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },

    updateTagData: function () {
      this.tagData = clone(TagsController.Tags);
    },

    getTagName: function (tag, data) {
      if (data[tag + ""]) {
        return data[tag + ""].name;
      } else {
        return "???";
      }
    },

    removeTag: function (tag) {
      if (this.busy) {
        return;
      }

      this.busy = true;

      const mediaId = AppStatus.CurrentMedia;
      const tagName = this.getTagName(tag, this.tagData);

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
              AppEvents.Emit(
                "snack",
                this.$t("Could not connect to the server")
              );
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
              AppEvents.Emit(
                "snack",
                this.$t("Could not connect to the server")
              );
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
              AppEvents.Emit(
                "snack",
                this.$t("Could not connect to the server")
              );
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
      if (this.$options.findTagTimeout) {
        return;
      }
      this.$options.findTagTimeout = setTimeout(() => {
        this.$options.findTagTimeout = null;
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
        this.matchingTags = Object.values(this.tagData)
          .map((a: any) => {
            return {
              id: a.id,
              name: a.name,
            };
          })
          .filter((a) => {
            if (this.tags.indexOf(a.id) >= 0) {
              return false;
            }
            return true
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
        e.preventDefault();
        this.findTags();
        if (this.matchingTags.length > 0) {
          this.tagToAdd = this.matchingTags[0].name;
        }
      }
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

    this.$options.tagUpdateH = this.updateTagData.bind(this);
    AppEvents.AddEventListener("tags-update", this.$options.tagUpdateH);

    this.$options.authUpdateH = this.updateAuthInfo.bind(this);

    AppEvents.AddEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );

    this.$options.mediaUpdateH = this.updateMediaData.bind(this);

    AppEvents.AddEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    this.updateTagData();
    this.load();

    if (this.display) {
      this.$options.focusTrap.activate();
      this.autoFocus();
      TagsController.Load();
    }
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("tags-update", this.$options.tagUpdateH);
    AppEvents.RemoveEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );
    AppEvents.RemoveEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        this.autoFocus();
        this.load();
        TagsController.Load();
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },
  },
});
</script>
