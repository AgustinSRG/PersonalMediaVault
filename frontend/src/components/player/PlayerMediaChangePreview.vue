<template>
    <div class="player-media-change-preview">
        <div class="player-media-change-preview-thumbnail">
            <img v-if="thumbnail" :src="thumbnail" :alt="title">
            <div v-if="!thumbnail" class="player-media-change-preview-no-thumbnail">
                <i v-if="type === 0" class="fas fa-ban"></i>
                <i v-if="type === 1" class="fas fa-image"></i>
                <i v-if="type === 2" class="fas fa-video"></i>
                <i v-if="type === 3" class="fas fa-headphones"></i>
            </div>
        </div>
        <div class="player-media-change-preview-details">
            <div class="player-media-change-preview-title">{{next ? $t('Next') : $t('Previous')}}: {{ title || $t('Untitled') }}</div>
            <div class="player-media-change-preview-type">
                <span v-if="type === 0"><i class="fas fa-ban"></i> {{ $t('Deleted media') }}</span>
                <span v-if="type === 1"><i class="fas fa-image"></i> {{ $t('Image') }}</span>
                <span v-if="type === 2"><i class="fas fa-video"></i> {{ $t('Video') }}</span>
                <span v-if="type === 3"><i class="fas fa-headphones"></i> {{ $t('Audio') }}</span>
            </div>
            <div class="player-media-change-preview-type" v-if="type === 2 || type === 3">{{ renderDuration(duration) }}</div>
            <div class="player-media-change-preview-type" v-if="type === 1">{{width}}x{{height}}</div>
            <div class="player-media-change-preview-type" v-if="type === 2">{{width}}x{{height}}, {{fps}} fps</div>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { renderTimeSeconds } from "../../utils/time-utils";

export default defineComponent({
  name: "PlayerMediaChangePreview",
  emits: [],
  props: {
    media: Object,
    next: Boolean,
  },
  data: function () {
    return {
        type: 0,
        thumbnail: "",
        title: "",
        duration: 0,
        width: 0,
        height: 0,
        fps: 0,
    };
  },
  methods: {
   upodateData: function () {
       if (this.media) {
           this.type = this.media.type;
           this.thumbnail = this.media.thumbnail;
           this.title = this.media.title;
           this.width = this.media.width;
           this.height = this.media.height;
           this.fps = this.media.fps;
           this.duration = this.media.duration;
       }
   },
   renderDuration: function (s) {
       return renderTimeSeconds(s);
   },
  },
  mounted: function () {
      this.upodateData();
  },
  beforeUnmount: function () {},
  watch: {
      media: function() {
          this.upodateData();
      },
  },
});
</script>

<style>

.player-media-change-preview {
    display: flex;
}

.player-min .player-media-change-preview {
    font-size: small;
}

.player-media-change-preview-thumbnail {
    width: 96px;
    height: 96px;
    display: flex;
    justify-content: center;
    align-items: center;
}

.player-min .player-media-change-preview-thumbnail {
    width: 64px;
    height: 64px;
}

.player-media-change-preview-thumbnail img {
    width: 100%;
    height: 100%;
}

.player-media-change-preview-details {
    padding-left: 0.5rem;
    text-align: left;
}

.player-media-change-preview-no-thumbnail {
    font-size: 24px;
    color: rgba(255, 255, 255, 0.5);
}

.player-min .player-media-change-preview-no-thumbnail {
    font-size: 16px;
}

.player-media-change-preview-title {
    font-weight: bold;
}

.player-media-change-preview-type {
    padding-top: 0.5rem;
    font-size: small;
}

.player-min .player-media-change-preview-type {
    font-size: x-small;
}

</style>
