<template>
  <div
    class="player-context-menu"
    :class="{
      hidden: !shown,
    }"
    :style="{
      top: top,
      left: left,
      right: right,
      bottom: bottom,
      width: width,
      'max-width': maxWidth,
      'max-height': maxHeight,
    }"
    tabindex="-1"
    @mousedown="stopPropagationEvent"
    @toutchstart.passive="stopPropagationEvent"
    @click="stopPropagationEvent"
    @dblclick="stopPropagationEvent"
  >
    <table class="player-context-menu-table">
      <tr
        v-if="type === 'video' || type === 'audio'"
        class="tr-button"
        tabindex="0"
        @click="toggleLoop"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-repeat icon-config"></i>
          <span class="context-entry-title">{{ $t("Loop") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': !loop }"></i>
        </td>
      </tr>

      <tr
        v-if="(type === 'video' || type === 'audio') && hasSlices"
        class="tr-button"
        tabindex="0"
        @click="toggleSliceLoop"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-repeat icon-config"></i>
          <span class="context-entry-title">{{ $t("Time slice loop") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': !sliceLoop }"></i>
        </td>
      </tr>

      <tr
        v-if="type === 'image'"
        class="tr-button"
        tabindex="0"
        @click="toggleFit"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-magnifying-glass icon-config"></i>
          <span class="context-entry-title">{{ $t("Fit image") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': !fit }"></i>
        </td>
      </tr>

      <tr
        v-if="type === 'image' || type === 'video'"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="toggleControls"
      >
        <td>
          <i class="fas fa-eye-slash icon-config"></i>
          <span class="context-entry-title">{{ $t("Hide controls") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': controls }"></i>
        </td>
      </tr>

      <tr
        v-if="type === 'image' && canWrite"
        class="tr-button"
        tabindex="0"
        @click="toggleNotes"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-pencil-alt icon-config"></i>
          <span class="context-entry-title">{{ $t("Edit image notes") }}</span>
        </td>
        <td class="td-right">
          <i class="fas fa-check" :class="{ 'check-uncheck': !notesEdit }"></i>
        </td>
      </tr>


      <tr
        class="tr-button"
        tabindex="0"
        @click="showTags"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-tag icon-config"></i>
          <span class="context-entry-title">{{ $t("Tags") }}</span>
        </td>
        <td class="td-right">
        </td>
      </tr>

      <tr
        class="tr-button"
        tabindex="0"
        @click="showExtendedDescription"
        @keydown="clickOnEnter"
      >
        <td>
          <i class="fas fa-info icon-config"></i>
          <span class="context-entry-title">{{ $t("Extended description") }}</span>
        </td>
        <td class="td-right">
        </td>
      </tr>

      <tr
        v-if="url"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="download"
      >
        <td>
          <i class="fas fa-download icon-config"></i>
          <span class="context-entry-title">{{ $t("Download") }}</span>
        </td>
        <td class="td-right"></td>
      </tr>

      <tr
        v-if="url"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="openStats"
      >
        <td>
          <i class="fas fa-bars-progress icon-config"></i>
          <span class="context-entry-title">{{ $t("Size Statistics") }}</span>
        </td>
        <td class="td-right"></td>
      </tr>

      <tr
        v-if="url"
        class="tr-button"
        tabindex="0"
        @keydown="clickOnEnter"
        @click="refreshMedia"
      >
        <td>
          <i class="fas fa-sync-alt icon-config"></i>
          <span class="context-entry-title">{{ $t("Refresh") }}</span>
        </td>
        <td class="td-right"></td>
      </tr>
    </table>
  </div>
</template>

<script lang="ts">
import { MediaController } from "@/control/media";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "PlayerContextMenu",
    emits: [
        "update:shown",
        "update:loop",
        "update:controls",
        "update:fit",
        "update:notesEdit",
        "update:sliceLoop",
        "open-tags",
        "open-ext-desc",
        "stats",
        "close",
    ],
    props: {
        shown: Boolean,
        type: String,
        x: Number,
        y: Number,

        url: String,

        loop: Boolean,
        fit: Boolean,
        controls: Boolean,

        sliceLoop: Boolean,
        hasSlices: Boolean,

        notesEdit: Boolean,
        canWrite: Boolean,
    },
    setup(props) {
        return {
            shownState: useVModel(props, "shown"),
            loopState: useVModel(props, "loop"),
            fitState: useVModel(props, "fit"),
            controlsState: useVModel(props, "controls"),
            notesState: useVModel(props, "notesEdit"),
            sliceLoopState: useVModel(props, "sliceLoop"),
        };
    },
    data: function () {
        return {
            top: "",
            left: "",
            right: "",
            bottom: "",

            width: "",

            maxWidth: "",
            maxHeight: "",
        };
    },
    methods: {
        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        toggleLoop: function () {
            this.loopState = !this.loopState;
            this.shownState = false;
            this.$emit("close");
        },

        toggleSliceLoop: function () {
            this.sliceLoopState = !this.sliceLoopState;
            this.shownState = false;
            this.$emit("close");
        },

        toggleFit: function () {
            this.fitState = !this.fitState;
            this.shownState = false;
            this.$emit("close");
        },

        toggleNotes: function () {
            this.notesState = !this.notesState;
            this.shownState = false;
            this.$emit("close");
        },

        toggleControls: function () {
            this.controlsState = !this.controlsState;
            this.shownState = false;
            this.$emit("close");
        },

        refreshMedia: function () {
            MediaController.Load();
            this.shownState = false;
            this.$emit("close");
        },

        showTags: function () {
            this.$emit("open-tags");
            this.shownState = false;
            this.$emit("close");
        },

        showExtendedDescription: function () {
            this.$emit("open-ext-desc");
            this.shownState = false;
            this.$emit("close");
        },

        download: function () {
            this.shownState = false;
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = this.url;
            link.click();
            this.$emit("close");
        },

        hide: function () {
            this.shownState = false;
            this.$emit("close");
        },

        openStats: function () {
            this.$emit("stats");
            this.shownState = false;
            this.$emit("close");
        },

        computeDimensions: function () {
            const pageWidth = window.innerWidth;
            const pageHeight = window.innerHeight;

            const x = this.x;
            const y = this.y;

            let top = y;
            let left = x;

            let maxWidth = pageWidth - left;

            let maxHeight = pageHeight - top;

            this.top = top + "px";
            this.left = left + "px";
            this.right = "auto";
            this.bottom = "auto";
            this.width = "auto";
            this.maxWidth = maxWidth + "px";
            this.maxHeight = maxHeight + "px";
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },
    },
    mounted: function () {
        this.computeDimensions();

        this.$options.hideHandler = this.hide.bind(this);

        document.addEventListener("mousedown", this.$options.hideHandler);
        document.addEventListener("touchstart", this.$options.hideHandler);
    },
    beforeUnmount: function () {
        document.removeEventListener("mousedown", this.$options.hideHandler);
        document.removeEventListener("touchstart", this.$options.hideHandler);
    },
    watch: {
        x: function () {
            this.computeDimensions();
        },
        y: function () {
            this.computeDimensions();
        },
        shown: function () {
            if (this.shown) {
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
    },
});
</script>
