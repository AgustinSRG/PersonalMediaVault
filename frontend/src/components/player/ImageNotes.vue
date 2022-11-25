<template>
  <div
    class="image-notes-container"
    :class="{ 'edit-active': editing, 'add-active': adding }"
    :style="{ top: top, left: left, width: width, height: height }"
    @mousedown="startAdding"
    @touchstart.passive="startAdding"
  >
    <div
      v-for="note in notes"
      :key="note.id"
      class="image-notes"
      :class="{ selected: selectedNotes === note.id }"
      :style="{
        top: mapDim(note.y, 0, realHeight, imageHeight),
        left: mapDim(note.x, 0, realWidth, imageWidth),
        width: mapDim(note.w, 0, realWidth, imageWidth),
        height: mapDim(note.h, 0, realHeight, imageHeight),
      }"
      @mousedown="clickOnNotes(note, $event)"
      @touchstart.passive="clickOnNotes(note, $event)"
    ></div>
    <div
      v-if="adding"
      class="image-notes creating"
      :style="{
        top: mapDim(addY, 0, realHeight, imageHeight),
        left: mapDim(addX, 0, realWidth, imageWidth),
        width: mapDim(addW, 0, realWidth, imageWidth),
        height: mapDim(addH, 0, realHeight, imageHeight),
      }"
    ></div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { isTouchDevice } from "../../utils/touch";
import { ImageNote, ImageNotesController } from "@/control/img-notes";
import { AppEvents } from "@/control/app-events";

export default defineComponent({
  name: "ImageNotes",
  emits: [],
  props: {
    top: String,
    left: String,
    width: String,
    height: String,

    editing: Boolean,

    contextopen: Boolean,
  },

  data: function () {
    return {
      notes: [],

      selectedNotes: -1,
      selectedNotesData: null,

      realWidth: 1,
      realHeight: 1,

      imageWidth: 1,
      imageHeight: 1,

      adding: false,

      addStartX: 0,
      addStartY: 0,

      addX: 0,
      addY: 0,
      addW: 0,
      addH: 0,

      moving: false,
      moveOriginalX: 0,
      moveOriginalY: 0,
      moveStartX: 0,
      moveStartY: 0,
    };
  },

  methods: {
    mapDim: function (
      dim: number,
      minDim: number,
      maxDim: number,
      imgDim: number
    ) {
      return (
        Math.min(
          maxDim,
          Math.max(minDim, Math.floor((dim * maxDim) / imgDim))
        ) + "px"
      );
    },

    startAdding: function (e) {
      if (this.contextopen) {
        return;
      }
      if ((e.which || e.button) !== 1) {
        return;
      }
      if (!this.editing) {
        return;
      }
      if (this.selectedNotesData) {
        this.selectedNotesData = null;
        this.selectedNotes = -1;
        return;
      }
      e.stopPropagation();
      const bounds = this.$el.getBoundingClientRect();
      let x: number;
      let y: number;
      if (e.touches && e.touches.length > 0) {
        x = e.touches[0].pageX;
        y = e.touches[0].pageY;
      } else {
        x = e.pageX;
        y = e.pageY;
      }

      console.log(bounds);

      const trueX = Math.max(
        0,
        Math.min(
          this.imageWidth - 32,
          Math.floor(((x - bounds.left) * this.imageWidth) / bounds.width)
        )
      );
      const trueY = Math.max(
        0,
        Math.min(
          this.imageHeight - 32,
          Math.floor(((y - bounds.top) * this.imageHeight) / bounds.height)
        )
      );

      this.adding = true;
      this.addStartX = trueX;
      this.addStartY = trueY;

      this.addX = trueX;
      this.addY = trueY;
      this.addW = 32;
      this.addH = 32;
    },

    mouseDrop: function () {
      if (!this.adding && !this.moving) {
        return;
      }
      if (this.adding) {
        this.adding = false;
        ImageNotesController.AddNote(
          this.addX,
          this.addY,
          this.addW,
          this.addH
        );
      }

      if (this.moving) {
        this.moving = false;
        if (this.selectedNotesData) {
          ImageNotesController.ModifyNote(this.selectedNotesData);
        }
      }
    },

    mouseMove: function (e) {
      if (!this.adding && !this.moving) {
        return;
      }
      const bounds = this.$el.getBoundingClientRect();
      let x: number;
      let y: number;
      if (e.touches && e.touches.length > 0) {
        x = e.touches[0].pageX;
        y = e.touches[0].pageY;
      } else {
        x = e.pageX;
        y = e.pageY;
      }
      if (this.adding) {
        const trueX = Math.min(
          this.imageWidth,
          Math.max(
            0,
            Math.floor(((x - bounds.left) * this.imageWidth) / bounds.width)
          )
        );
        const trueY = Math.min(
          this.imageHeight,
          Math.max(
            0,
            Math.floor(((y - bounds.top) * this.imageHeight) / bounds.height)
          )
        );

        if (trueX - this.addStartX > 0) {
          this.addX = this.addStartX;
          this.addW = Math.max(32, trueX - this.addStartX);
        } else {
          this.addX = Math.max(0, trueX);
          this.addW = Math.max(32, this.addStartX - trueX);
        }

        if (trueY - this.addStartY > 0) {
          this.addY = this.addStartY;
          this.addH = trueY - this.addStartY;
        } else {
          this.addY = Math.max(0, trueY);
          this.addH = Math.max(32, this.addStartY - trueY);
        }
      }
      if (this.moving && this.selectedNotesData) {
        const trueX = Math.min(
          this.imageWidth,
          Math.max(
            0,
            Math.floor(((x - bounds.left) * this.imageWidth) / bounds.width)
          )
        );
        const trueY = Math.min(
          this.imageHeight,
          Math.max(
            0,
            Math.floor(((y - bounds.top) * this.imageHeight) / bounds.height)
          )
        );

        const diffX = this.moveStartX - trueX;
        this.selectedNotesData.x = Math.max(0, this.moveOriginalX - diffX);

        if (this.selectedNotesData.x + this.selectedNotesData.w > this.imageWidth) {
            this.selectedNotesData.x = Math.max(0, this.imageWidth - this.selectedNotesData.w);
        }

        const diffY = this.moveStartY - trueY;
        this.selectedNotesData.y = Math.max(0, this.moveOriginalY - diffY);

        if (this.selectedNotesData.y + this.selectedNotesData.h > this.imageHeight) {
            this.selectedNotesData.y = Math.max(0, this.imageHeight - this.selectedNotesData.h);
        }
      }
    },

    clickOnNotes: function (notes, e) {
      if (this.contextopen) {
        return;
      }
      if (!this.editing) {
        return;
      }
      if ((e.which || e.button) !== 1) {
        return;
      }
      e.stopPropagation();
      this.selectedNotes = notes.id;
      this.selectedNotesData = notes;
      const bounds = this.$el.getBoundingClientRect();
      let x: number;
      let y: number;
      if (e.touches && e.touches.length > 0) {
        x = e.touches[0].pageX;
        y = e.touches[0].pageY;
      } else {
        x = e.pageX;
        y = e.pageY;
      }
      const trueX = Math.max(
        0,
        Math.min(
          this.imageWidth,
          Math.floor(((x - bounds.left) * this.imageWidth) / bounds.width)
        )
      );
      const trueY = Math.max(
        0,
        Math.min(
          this.imageHeight,
          Math.floor(((y - bounds.top) * this.imageHeight) / bounds.height)
        )
      );

      this.moving = true;
      this.moveStartX = trueX;
      this.moveStartY = trueY;
      this.moveOriginalX = notes.x;
      this.moveOriginalY = notes.y;
    },

    onNotesUpdate: function () {
      this.notes = ImageNotesController.GetNotes();
      this.imageWidth = Math.max(1, ImageNotesController.ImageWidth);
      this.imageHeight = Math.max(1, ImageNotesController.ImageHeight);
    },

    onNotesPush: function (note: ImageNote) {
      this.notes.push({
        id: note.id,
        x: note.x,
        y: note.y,
        w: note.w,
        h: note.h,
        text: note.text,
      });
    },

    onNotesChange: function (i: number, note: ImageNote) {
      this.notes[i] = {
        x: note.x,
        y: note.y,
        w: note.w,
        h: note.h,
        text: note.text,
      };
    },

    onNotesRemove: function (i: number) {
      this.notes.splice(i, 1);
    },

    updateRealDimensions: function () {
      nextTick(() => {
        const bounds = this.$el.getBoundingClientRect();
        this.realWidth = bounds.width;
        this.realHeight = bounds.height;
      });
    },
  },

  mounted: function () {
    this.updateRealDimensions();

    this.$options.onNotesUpdateH = this.onNotesUpdate.bind(this);
    AppEvents.AddEventListener(
      "img-notes-update",
      this.$options.onNotesUpdateH
    );

    this.$options.onNotesPushH = this.onNotesPush.bind(this);
    AppEvents.AddEventListener("img-notes-push", this.$options.onNotesPushH);

    this.$options.onNotesChangeH = this.onNotesChange.bind(this);
    AppEvents.AddEventListener(
      "img-notes-change",
      this.$options.onNotesChangeH
    );

    this.$options.onNotesRemoveH = this.onNotesRemove.bind(this);
    AppEvents.AddEventListener("img-notes-rm", this.$options.onNotesRemoveH);

    this.$options.mouseDropH = this.mouseDrop.bind(this);
    document.addEventListener("mouseup", this.$options.mouseDropH);
    document.addEventListener("touchend", this.$options.mouseDropH);

    this.$options.mouseMoveH = this.mouseMove.bind(this);

    document.addEventListener("mousemove", this.$options.mouseMoveH);
    document.addEventListener("touchmove", this.$options.mouseMoveH);

    this.onNotesUpdate();
  },

  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "img-notes-update",
      this.$options.onNotesUpdateH
    );
    AppEvents.RemoveEventListener("img-notes-push", this.$options.onNotesPushH);
    AppEvents.RemoveEventListener(
      "img-notes-change",
      this.$options.onNotesChangeH
    );
    AppEvents.RemoveEventListener("img-notes-rm", this.$options.onNotesRemoveH);

    document.removeEventListener("mouseup", this.$options.mouseDropH);
    document.removeEventListener("touchend", this.$options.mouseDropH);
    document.removeEventListener("mousemove", this.$options.mouseMoveH);
    document.removeEventListener("touchmove", this.$options.mouseMoveH);
  },

  watch: {
    top: function () {
      this.updateRealDimensions();
    },

    left: function () {
      this.updateRealDimensions();
    },

    width: function () {
      this.updateRealDimensions();
    },

    height: function () {
      this.updateRealDimensions();
    },
  },
});
</script>
