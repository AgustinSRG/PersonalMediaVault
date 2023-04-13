<template>
  <div
    class="player-top-bar"
    :class="{
      hidden: !shown,
      'with-album': inAlbum,
      'album-expand': albumExpanded,
      expanded: expanded && !albumExpanded,
    }"
    tabindex="-1"
    @click="clickTopBar"
    @dblclick="stopPropagationEvent"
    @contextmenu="stopPropagationEvent"
    @keydown="onKeyDown"
  >
    <div v-if="!albumExpanded" class="player-title-container">
      <div class="player-title-left">
        <button
          type="button"
          :title="$t('View Album')"
          class="player-btn"
          @click="expandAlbum"
        >
          <i class="fas fa-list-ol"></i>
        </button>
      </div>
      <div class="player-title">
        <div v-if="metadata">{{ metadata.title }}</div>
      </div>
      <div class="player-title-right">
        <button
          v-if="metadata && !expanded"
          type="button"
          :title="$t('Expand')"
          class="player-btn"
          @click="expandTitle"
        >
          <i class="fas fa-chevron-down"></i>
        </button>

        <button
          v-if="metadata && expanded"
          type="button"
          :title="$t('Close')"
          class="player-btn"
          @click="closeTitle"
        >
          <i class="fas fa-chevron-up"></i>
        </button>
      </div>
    </div>

    <PlayerAlbumFullScreen
      :expanded="albumExpanded"
      @close="closeAlbum"
    ></PlayerAlbumFullScreen>
    <PlayerMediaEditor
      v-if="expanded"
      @changed="onEditDone"
    ></PlayerMediaEditor>
  </div>
</template>


<script lang="ts">
import { MediaController } from "@/control/media";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import PlayerAlbumFullScreen from "./PlayerAlbumFullScreen.vue";
import PlayerMediaEditor from "./PlayerMediaEditor.vue";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";

export default defineComponent({
  name: "PlayerTopBar",
  components: {
    PlayerAlbumFullScreen,
    PlayerMediaEditor,
  },
  emits: ["update:expanded", "update:albumExpanded", "click-player"],
  props: {
    mid: Number,
    metadata: Object,

    inAlbum: Boolean,

    shown: Boolean,
    fullscreen: Boolean,
    expanded: Boolean,
    albumExpanded: Boolean,
  },
  setup(props) {
    return {
      expandedState: useVModel(props, "expanded"),
      albumExpandedState: useVModel(props, "albumExpanded"),
    };
  },
  data: function () {
    return {
      dirty: false,
    };
  },
  methods: {
    clickTopBar: function (e) {
      e.stopPropagation();
      this.$emit("click-player");
    },

    expandTitle: function () {
      this.albumExpandedState = false;
      this.expandedState = true;
    },

    onEditDone: function () {
      this.dirty = true;
    },

    closeTitle: function () {
      this.expandedState = false;
    },

    expandAlbum: function () {
      this.albumExpandedState = true;
      this.expandedState = false;
    },

    closeAlbum: function () {
      this.albumExpandedState = false;
    },

    close: function () {
      this.closeTitle();
      this.closeAlbum();
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    onKeyDown: function (e) {
      if (!this.expanded && !this.albumExpanded) {
        return;
      }
      e.stopPropagation();
      if (e.key === "Escape") {
        e.preventDefault();
        this.close();
      }
    },

    handleGlobalKey: function (event: KeyboardEvent): boolean {
      if (AuthController.Locked || !event.key || event.ctrlKey) {
        return false;
      }

      if (event.key.toUpperCase() === "E") {
        this.expandTitle();
        return true;
      }

      return false;
    },
  },
  watch: {
    fullscreen: function () {
      this.albumExpandedState = false;
    },

    expanded: function () {
      if (this.expanded) {
        nextTick(() => {
            const el = this.$el.querySelector(".player-media-editor");
            if (el) {
              el.focus();
            }
          });
      }
      if (this.dirty) {
        this.dirty = false;
        setTimeout(() => {
          MediaController.Load();
        }, 100);
      }
    },

    albumExpanded: function () {
      if (this.albumExpanded) {
         nextTick(() => {
            const el = this.$el.querySelector(".player-album-container");
            if (el) {
              el.focus();
            }
          });
      }
    },
  },
  mounted: function () {
    this.$options.handleGlobalKeyH = this.handleGlobalKey.bind(this);
    KeyboardManager.AddHandler(this.$options.handleGlobalKeyH);
  },
  beforeUnmount: function () {
    KeyboardManager.RemoveHandler(this.$options.handleGlobalKeyH);
  },
});
</script>
