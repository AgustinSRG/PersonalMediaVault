<template>
  <div
    class="player-top-bar"
    :class="{ hidden: !shown, 'with-album': !!album }"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
  >
    <div v-if="!expanded && !albumexpanded" class="player-title-container">
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
        <div>{{ metadata.title }}</div>
      </div>
      <div class="player-title-right">
        <button
          type="button"
          :title="$t('Expand')"
          class="player-btn"
          @click="expandTitle"
        >
          <i class="fas fa-chevron-down"></i>
        </button>
      </div>
    </div>
  </div>
</template>


<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "PlayerTopBar",
  emits: [],
  props: {
    mid: Number,
    metadata: Object,
    album: Object,

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
    return {};
  },
  methods: {
    expandTitle: function () {},

    expandAlbum: function () {},

    stopPropagationEvent: function (e) {
      e.stopPropagation();
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
  transition: opacity 0.3s;
  opacity: 1;
  color: white;
  overflow: hidden;
  height: 56px;
  top: 0;
  left: 0;
  width: 100%;
  display: flex;
}

.player-min .player-top-bar {
  height: 32px;
}

.player-top-bar.hidden {
  opacity: 0;
  pointer-events: none;
}

.player-title-container {
  width: 100%;
  height: 100%;
  display: flex;
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

.with-album .player-title-left {
  width: 48px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: right;
}

.player-min .with-album .player-title-left {
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