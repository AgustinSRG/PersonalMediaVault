<template>
  <div
    class="player-top-bar"
    :class="{ hidden: !shown, 'with-album': inalbum, 'album-expand': albumexpanded, 'expanded': expanded && !albumexpanded }"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
  >
    <div v-if="!albumexpanded" class="player-title-container">
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

    <PlayerAlbumFullScreen :expanded="albumexpanded" @close="closeAlbum"></PlayerAlbumFullScreen>

  </div>
</template>


<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";
import PlayerAlbumFullScreen from "./PlayerAlbumFullScreen.vue";

export default defineComponent({
  name: "PlayerTopBar",
  components: {
    PlayerAlbumFullScreen,
  },
  emits: ['update:expanded', 'update:albumexpanded'],
  props: {
    mid: Number,
    metadata: Object,

    inalbum: Boolean,

    shown: Boolean,
    fullscreen: Boolean,
    expanded: Boolean,
    albumexpanded: Boolean,
  },
  setup(props) {
    return {
      expandedState: useVModel(props, "expanded"),
      albumexpandedState: useVModel(props, "albumexpanded"),
    };
  },
  data: function () {
    return {
    };
  },
  methods: {
    expandTitle: function () {
      this.albumexpandedState = false;
      this.expandedState = true;
    },

    closeTitle: function () {
      this.expandedState = false;
    },

    expandAlbum: function () {
      this.albumexpandedState = true;
      this.expandedState = false;
    },

    closeAlbum: function () {
      this.albumexpandedState = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },
  },
  watch: {
    fullscreen: function () {
      this.albumexpandedState = false;
    },
  },
  mounted: function () {},
  beforeUnmount: function () {},
});
</script>

<style>
.player-top-bar {
  position: absolute;
  background-color: rgba(0, 0, 0, 0.2);
  transition: opacity 0.3s, height 0.1s, width 0.1s;
  opacity: 1;
  color: white;
  overflow: hidden;
  height: 56px;
  top: 0;
  left: 0;
  width: 100%;
}

.player-min .player-top-bar {
  height: 32px;
}

.player-top-bar.album-expand,
.player-min .player-top-bar.album-expand {
  max-width: 500px;
  width: 100%;
  height: 100%;
}

.player-top-bar.expanded,
.player-min .player-top-bar.expanded {
  width: 100%;
  height: 100%;
}

.player-top-bar.hidden {
  opacity: 0;
  pointer-events: none;
}

.player-title-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 57px;
  display: flex;
}

.player-min .player-title-container {
  height: 32px;
}

.player-title {
  width: calc(100% - 48px);
  height: 100%;
  display: flex;
  align-items: center;
}

.player-min .player-title {
  width: calc(100% - 28px);
}

.player-title div {
  width: 100%;
  padding-right: 8px;
  padding-left: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left;
  font-size: large;
}

.player-min .player-title div {
  font-size: small;
  padding-right: 4px;
  padding-left: 4px;
}

.full-screen .with-album .player-title {
  width: calc(100% - 96px);
}

.player-min.full-screen .with-album .player-title {
  width: calc(100% - 56px);
}

.player-title-left {
  display: none;
}

.full-screen .with-album .player-title-left {
  width: 48px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: right;
}

.player-min.full-screen .with-album .player-title-left {
  width: 28px;
}

.player-title-right {
  width: 48px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: left;
}

.player-min .player-title-right {
  width: 28px;
}
</style>