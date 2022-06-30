<template>
  <div class="player-media-editor">
    <form @submit="changeTitle">
      <div class="form-group">
        <label>{{ $t("Title") }}:</label>
        <input
          type="text"
          autocomplete="off"
          maxlength="255"
          :disabled="busy"
          v-model="title"
          class="form-control form-control-full-width"
        />
      </div>
      <div class="form-group">
        <button
          type="submit"
          class="btn btn-primary btn-sm"
          :disabled="busy || !title || originalTitle === title"
        >
          <i class="fas fa-pencil-alt"></i> {{ $t("Change title") }}
        </button>
      </div>
    </form>
    <div class="form-group border-top">
      <label>{{ $t("Description") }}:</label>
      <textarea
        v-model="desc"
        class="form-control form-control-full-width form-textarea"
        rows="3"
        :disabled="busy"
      ></textarea>
    </div>
    <div class="form-group">
      <button
        type="button"
        class="btn btn-primary btn-sm"
        :disabled="busy || originalDesc === desc"
        @click="changeDescription"
      >
        <i class="fas fa-pencil-alt"></i> {{ $t("Change description") }}
      </button>
    </div>
    <div class="form-group border-top">
      <label>{{ $t("Tags") }}:</label>
    </div>
    <div class="form-group media-tags">
      <label v-if="tags.length === 0">{{
        $t("There are no tags yet for this media.")
      }}</label>
      <div v-for="tag in tags" :key="tag" class="media-tag">
        <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
        <button
          type="button"
          :title="$t('Remove tag')"
          class="media-tag-btn"
          @click="removeTag(tag)"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
    </div>
    <form @submit="addTag">
      <div class="form-group">
        <label>{{ $t("Tag to add") }}:</label>
        <input
          type="text"
          autocomplete="off"
          maxlength="255"
          v-model="tagToAdd"
          class="form-control"
        />
      </div>
      <div class="form-group">
        <button
          type="submit"
          class="btn btn-primary btn-sm"
          :disabled="busy || !tagToAdd"
        >
          <i class="fas fa-plus"></i> {{ $t("Add Tag") }}
        </button>
      </div>
    </form>
    <div class="form-group border-top">
      <label>{{
        $t(
          "If the media resource did not encode properly, try using the button below. If it still does not work, try re-uploading the media."
        )
      }}</label>
    </div>
    <div class="form-group">
      <button
        type="button"
        class="btn btn-primary btn-sm"
        :disabled="busy"
        @click="encodeMedia"
      >
        <i class="fas fa-sync-alt"></i> {{ $t("Re-Encode") }}
      </button>
    </div>
    <div class="form-group border-top">
      <label>{{
        $t("If you want to delete this media resource, click the button below.")
      }}</label>
    </div>
    <div class="form-group">
      <button
        type="button"
        class="btn btn-danger btn-sm"
        :disabled="busy"
        @click="deleteMedia"
      >
        <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { MediaAPI } from "@/api/api-media";
import { TagsAPI } from "@/api/api-tags";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { MediaController } from "@/control/media";
import { TagsController } from "@/control/tags";
import { copyObject } from "@/utils/objects";
import { Request } from "@/utils/request";
import { defineComponent } from "vue";

export default defineComponent({
  name: "PlayerMediaEditor",
  emits: ["changed"],
  data: function () {
    return {
      title: "",
      originalTitle: "",

      desc: "",
      originalDesc: "",

      tags: [],
      tagToAdd: "",
      tagData: {},
      matchingTags: [],

      busy: false,
    };
  },

  methods: {
    updateMediaData: function () {
      if (!MediaController.MediaData) {
        return;
      }

      this.originalTitle = MediaController.MediaData.title;
      this.title = this.originalTitle;

      this.originalDesc = MediaController.MediaData.description;
      this.desc = this.originalDesc;

      this.tags = MediaController.MediaData.tags.slice();
    },

    changeTitle: function (e) {
      if (e) {
        e.preventDefault();
      }

      if (this.busy) {
        return;
      }

      this.busy = true;

      const mediaId = AppStatus.CurrentMedia;

      Request.Do(MediaAPI.ChangeMediaTitle(mediaId, this.title))
        .onSuccess(() => {
          AppEvents.Emit("snack", this.$t("Successfully changed title"));
          this.busy = false;
          this.originalTitle = this.title;
          this.$emit("changed");
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              AppEvents.Emit("snack", this.$t("Bad request"));
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

    changeDescription: function () {
      if (this.busy) {
        return;
      }

      this.busy = true;

      const mediaId = AppStatus.CurrentMedia;

      Request.Do(MediaAPI.ChangeMediaDescription(mediaId, this.desc))
        .onSuccess(() => {
          AppEvents.Emit("snack", this.$t("Successfully changed description"));
          this.busy = false;
          this.originalDesc = this.desc;
          this.$emit("changed");
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              AppEvents.Emit("snack", this.$t("Bad request"));
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

    encodeMedia: function () {
      if (this.busy) {
        return;
      }

      this.busy = true;

      const mediaId = AppStatus.CurrentMedia;

      Request.Do(MediaAPI.EncodeMedia(mediaId))
        .onSuccess(() => {
          AppEvents.Emit(
            "snack",
            this.$t("Successfully requested pending encoding tasks")
          );
          this.busy = false;
          MediaController.Load();
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
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

    deleteMedia: function () {
      AppEvents.Emit("media-delete-request");
    },

    updateTagData: function () {
      this.tagData = copyObject(TagsController.Tags);
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

      Request.Do(TagsAPI.UntagMedia(mediaId, tag))
        .onSuccess(() => {
          AppEvents.Emit("snack", this.$t("Removed tag") + ": " + tagName);
          this.busy = false;
          for (let i = 0; i < this.tags.length; i++) {
            if (this.tags[i] === tag) {
              this.tags.splice(i, 1);
              break;
            }
          }
          this.$emit("changed");
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

      Request.Do(TagsAPI.TagMedia(mediaId, tag))
        .onSuccess((res) => {
          AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
          this.busy = false;
          this.tagToAdd = "";
          if (this.tags.indexOf(res.id) === -1) {
            this.tags.push(res.id);
          }
          TagsController.AddTag(res.id, res.name);
          this.$emit("changed");
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
  },

  mounted: function () {
    this.updateMediaData();
    this.updateTagData();

    this.$options.mediaUpdateH = this.updateMediaData.bind(this);

    AppEvents.AddEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    this.$options.tagUpdateH = this.updateTagData.bind(this);

    AppEvents.AddEventListener("tags-update", this.$options.tagUpdateH);
  },

  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    AppEvents.RemoveEventListener("tags-update", this.$options.tagUpdateH);
  },
});
</script>

<style>
.player-media-editor {
  position: absolute;
  top: 57px;
  left: 0;
  width: 100%;
  height: calc(100% - 57px);
  overflow: auto;
  background: rgba(0, 0, 0, 0.9);
  color: white;
  padding: 1rem;
}

.player-min .player-media-editor {
  top: 32px;
  height: calc(100% - 32px);
}

.form-group.border-top {
  border-top: solid 1px rgba(255, 255, 255, 0.1);
  padding-top: 1rem;
}

.form-control.form-textarea {
  resize: vertical;
  height: auto;
}

.media-tags {
  display: flex;
  flex-wrap: wrap;
}

.media-tag {
  border: solid 1px rgba(255, 255, 255, 0.1);
  border-radius: 100vw;
  display: flex;
  align-items: center;
  margin: 0.5rem;
  padding: 0.25rem;
}

.media-tag-name {
  padding: 0.5rem;
}

.media-tag-btn {
  display: block;
  width: 32px;
  height: 32px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
  outline: none;
}

.media-tag-btn:hover {
  color: white;
}

.media-tag-btn:focus {
  outline: none;
}
</style>