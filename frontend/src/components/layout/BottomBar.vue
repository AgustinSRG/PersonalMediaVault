<template>
  <div class="bottom-bar">
    <div class="bottom-bar-option bottom-bar-option-media" :class="{selected: focus === 'left'}" tabindex="0" @click="clickLeft">
      <i class="fas fa-photo-film"></i> {{ $t('Media') }}
    </div>
    <div class="bottom-bar-option bottom-bar-option-list" :class="{selected: focus === 'right'}" tabindex="0" @click="clickRight">
      <i class="fas fa-list"></i> {{ $t('List') }}
    </div>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";

export default defineComponent({
  name: "BottomBar",
  data: function () {
    return {
      focus: AppStatus.CurrentFocus,
    };
  },
  methods: {
    onStatusUpdate: function () {
      this.focus = AppStatus.CurrentFocus;
    },

    clickLeft: function () {
      AppStatus.FocusLeft();
    },

    clickRight: function () {
      AppStatus.FocusRight();
    },
  },
  mounted: function () {
    this.$options.updateStatusH = this.onStatusUpdate.bind(this);
    AppEvents.AddEventListener("app-status-update", this.$options.updateStatusH);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("app-status-update", this.$options.updateStatusH);
  },
});
</script>

<style>
.bottom-bar {
  display: none;
  flex-direction: row;
  justify-content: center;
  align-items: center;


  position: fixed;
  bottom: 0;
  height: 40px;
  left: 0;
  width: 100%;
  background: #212121;
}

.layout-media-split .bottom-bar,
.layout-album .bottom-bar {
  display: flex;
}

@media (min-width: 1000px) {
  .bottom-bar {
    display: none;
  }

  .layout-media-split .bottom-bar,
  .layout-album .bottom-bar {
    display: none;
  }
}

.bottom-bar-option {
  width: 50%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  white-space: nowrap;
  font-size: 1.1rem;
  cursor: pointer;
}

.bottom-bar-option i {
  margin-right: 1rem;
}

.bottom-bar-option:hover {
  background: rgba(255, 255, 255, 0.1);
}

.bottom-bar-option.selected,
.bottom-bar-option.selected:hover {
  background: rgba(255, 255, 255, 0.2);
}

</style>