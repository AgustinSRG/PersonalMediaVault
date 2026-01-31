<template>
    <div
        class="image-notes-container"
        :class="{ 'edit-active': editing, 'add-active': adding, 'always-visible': visible && !editing && !adding }"
        :style="{ top: top, left: left, width: width, height: height }"
        @mousedown="startAddingMouse"
        @touchstart.passive="startAddingTouch"
        @mousemove="onMouseMoveInsideContainer"
        @mouseleave="onMouseLeave"
    >
        <div
            v-for="note in notes"
            :key="note.id"
            class="image-notes"
            tabindex="0"
            :class="{ selected: selectedNotes === note.id }"
            :style="{
                top: mapDim(note.y, 0, realHeight, imageHeight),
                left: mapDim(note.x, 0, realWidth, imageWidth),
                width: mapDim(note.w, 0, realWidth, imageWidth),
                height: mapDim(note.h, 0, realHeight, imageHeight),
            }"
            @mousedown="clickOnNotes(note, $event)"
            @touchstart.passive="clickOnNotes(note, $event)"
        >
            <div
                v-if="editing"
                class="resize resize-left"
                @mousedown="startResizeNotes(note, $event, 'l')"
                @touchstart.passive="startResizeNotes(note, $event, 'l')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-top"
                @mousedown="startResizeNotes(note, $event, 't')"
                @touchstart.passive="startResizeNotes(note, $event, 't')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-right"
                @mousedown="startResizeNotes(note, $event, 'r')"
                @touchstart.passive="startResizeNotes(note, $event, 'r')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-bottom"
                @mousedown="startResizeNotes(note, $event, 'b')"
                @touchstart.passive="startResizeNotes(note, $event, 'b')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-top-left"
                @mousedown="startResizeNotes(note, $event, 'tl')"
                @touchstart.passive="startResizeNotes(note, $event, 'tl')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-top-right"
                @mousedown="startResizeNotes(note, $event, 'tr')"
                @touchstart.passive="startResizeNotes(note, $event, 'tr')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-bottom-left"
                @mousedown="startResizeNotes(note, $event, 'bl')"
                @touchstart.passive="startResizeNotes(note, $event, 'bl')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-bottom-right"
                @mousedown="startResizeNotes(note, $event, 'br')"
                @touchstart.passive="startResizeNotes(note, $event, 'br')"
            ></div>

            <div
                v-if="editing && !moving && !resizing && selectedNotes === note.id"
                class="image-notes-text-edit"
                :class="{
                    top: note.y + note.h / 2 < imageHeight / 2,
                    left: note.x + note.w / 2 < imageWidth / 2,
                    bottom: note.y + note.h / 2 >= imageHeight / 2,
                    right: note.x + note.w / 2 >= imageWidth / 2,
                }"
                tabindex="-1"
                @dblclick="stopPropagationEvent"
                @keydown="stopPropagationEvent"
                @click="stopPropagationEvent"
                @mousedown="stopPropagationEvent"
                @touchstart.passive="stopPropagationEvent"
                @contextmenu="stopPropagationEvent"
            >
                <div class="form-group">
                    <textarea
                        v-model="note.text"
                        class="form-control form-textarea form-control-full-width auto-focus"
                        :placeholder="$t('Type the notes text') + '...'"
                        @change="saveNote(note)"
                    ></textarea>
                </div>
                <div class="form-group">
                    <button type="button" class="btn btn-primary btn-xs btn-mr" @click="saveNote(note)">
                        <i class="fas fa-check"></i> {{ $t("Save") }}
                    </button>
                    <button type="button" class="btn btn-danger btn-xs" @click="deleteNote(note)">
                        <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                    </button>
                </div>
            </div>
        </div>
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

        <div
            v-if="!editing && selectedNote"
            class="image-notes-hover"
            :style="{
                top: hoverTop,
                bottom: hoverBottom,
                left: hoverLeft,
                right: hoverRight,
            }"
            v-html="escapeText(selectedNote.text)"
        ></div>
    </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { ImageNotesController } from "@/control/img-notes";
import { escapeHTML } from "@/utils/html";
import type { ImageNote } from "@/utils/notes-format";
import type { PositionEvent } from "@/utils/position-event";
import { positionEventFromMouseEvent, positionEventFromTouchEvent } from "@/utils/position-event";
import { EVENT_NAME_IMAGE_NOTES_UPDATE, EVENT_NAME_IMAGE_NOTES_CHANGE } from "@/control/app-events";

export default defineComponent({
    name: "ImageNotes",
    props: {
        top: String,
        left: String,
        width: String,
        height: String,

        visible: Boolean,

        editing: Boolean,

        contextOpen: Boolean,
    },
    data: function () {
        return {
            notes: [] as ImageNote[],

            selectedNotes: -1,
            selectedNotesData: null as ImageNote | null,

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

            resizing: false,
            resizeOriginalX: 0,
            resizeOriginalY: 0,
            resizeOriginalW: 0,
            resizeOriginalH: 0,
            resizeStartX: 0,
            resizeStartY: 0,
            resizeMode: "",

            selectedNote: null as ImageNote | null,
            hoverRight: "",
            hoverLeft: "",
            hoverTop: "",
            hoverBottom: "",
            hoverPinned: false,
        };
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

    mounted: function () {
        this.updateRealDimensions();

        this.$listenOnAppEvent(EVENT_NAME_IMAGE_NOTES_UPDATE, this.onNotesUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_IMAGE_NOTES_CHANGE, (mode: "push" | "rm" | "update", note?: ImageNote, index?: number) => {
            switch (mode) {
                case "push":
                    this.onNotesPush(note);
                    break;
                case "rm":
                    this.onNotesRemove(index);
                    break;
                case "update":
                    this.onNotesChange(index, note);
                    break;
            }
        });

        this.$listenOnDocumentEvent("mouseup", this.mouseDrop.bind(this));
        this.$listenOnDocumentEvent("touchend", this.mouseDrop.bind(this));

        this.$listenOnDocumentEvent("mousemove", this.mouseMove.bind(this));
        this.$listenOnDocumentEvent("touchmove", this.touchMove.bind(this));

        this.onNotesUpdate();
    },

    methods: {
        mapDim: function (dim: number, minDim: number, maxDim: number, imgDim: number) {
            return Math.min(maxDim, Math.max(minDim, Math.round((dim * maxDim) / imgDim))) + "px";
        },

        escapeText: function (txt: string): string {
            return escapeHTML(txt).replace(/\n/g, "<br>");
        },

        startAddingMouse: function (e: MouseEvent) {
            if (e.button !== 0) {
                return; // Not the main button
            }

            this.startAdding(positionEventFromMouseEvent(e));
        },

        startAddingTouch: function (e: TouchEvent) {
            this.startAdding(positionEventFromTouchEvent(e));
        },

        startAdding: function (e: PositionEvent) {
            if (!this.editing) {
                this.onClickFindNotes(e);
                return;
            }
            if (this.contextOpen) {
                return;
            }
            if (this.selectedNotesData) {
                this.selectedNotesData = null;
                this.selectedNotes = -1;
                return;
            }

            e.e.stopPropagation();

            const bounds = this.$el.getBoundingClientRect();
            const x = e.x;
            const y = e.y;

            const trueX = Math.max(0, Math.min(this.imageWidth - 32, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
            const trueY = Math.max(0, Math.min(this.imageHeight - 32, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

            this.adding = true;
            this.addStartX = trueX;
            this.addStartY = trueY;

            this.addX = trueX;
            this.addY = trueY;
            this.addW = 32;
            this.addH = 32;
        },

        onClickFindNotes: function (e: PositionEvent) {
            this.hoverPinned = false;
            this.onMoveInsideNotesContainer(e);

            this.hoverPinned = !!this.selectedNote;
        },

        onMouseLeave: function () {
            if (this.hoverPinned) {
                return;
            }
            this.selectedNote = null;
        },

        onMouseMoveInsideContainer: function (e: MouseEvent) {
            this.onMoveInsideNotesContainer(positionEventFromMouseEvent(e));
        },

        onMoveInsideNotesContainer: function (e: PositionEvent) {
            const x = e.x;
            const y = e.y;

            const bounds = this.$el.getBoundingClientRect();

            const realY = y - bounds.top;
            const realX = x - bounds.left;

            const trueX = Math.max(0, Math.min(this.imageWidth - 32, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
            const trueY = Math.max(0, Math.min(this.imageHeight - 32, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

            for (const note of this.notes) {
                if (this.hoverPinned && this.selectedNote === note) {
                    continue;
                }
                if (trueX >= note.x && trueX <= note.x + note.w) {
                    if (trueY >= note.y && trueY <= note.y + note.h) {
                        this.selectedNote = note;

                        // Position mouse

                        if (realY < bounds.height / 2) {
                            this.hoverTop = realY + 8 + "px";
                            this.hoverBottom = "";
                        } else {
                            this.hoverTop = "";
                            this.hoverBottom = bounds.height - realY + 8 + "px";
                        }

                        if (realX < bounds.width / 2) {
                            this.hoverLeft = realX + 8 + "px";
                            this.hoverRight = "";
                        } else {
                            this.hoverLeft = "";
                            this.hoverRight = bounds.width - realX + 8 + "px";
                        }

                        this.hoverPinned = false;

                        return;
                    }
                }
            }

            if (!this.hoverPinned) {
                this.selectedNote = null;
            }
        },

        autoFocus: function () {
            nextTick(() => {
                const editElement = this.$el.querySelector(".auto-focus");
                if (editElement) {
                    editElement.focus();
                }
            });
        },

        mouseDrop: function () {
            if (!this.adding && !this.moving && !this.resizing) {
                return;
            }
            if (this.adding) {
                this.adding = false;
                ImageNotesController.AddNote(this.addX, this.addY, this.addW, this.addH);
            }

            if (this.moving) {
                this.moving = false;
                if (this.selectedNotesData) {
                    ImageNotesController.ModifyNote(this.selectedNotesData);
                }
                this.autoFocus();
            }
            if (this.resizing) {
                this.resizing = false;
                if (this.selectedNotesData) {
                    ImageNotesController.ModifyNote(this.selectedNotesData);
                }
                this.autoFocus();
            }
        },

        touchMove: function (e: TouchEvent) {
            this.move(positionEventFromTouchEvent(e));
        },

        mouseMove(e: MouseEvent) {
            this.move(positionEventFromMouseEvent(e));
        },

        move: function (e: PositionEvent) {
            if (!this.adding && !this.moving && !this.resizing) {
                return;
            }
            const bounds = this.$el.getBoundingClientRect();
            const x = e.x;
            const y = e.y;

            if (this.adding) {
                const trueX = Math.min(this.imageWidth, Math.max(0, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
                const trueY = Math.min(this.imageHeight, Math.max(0, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

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
                const trueX = Math.min(this.imageWidth, Math.max(0, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
                const trueY = Math.min(this.imageHeight, Math.max(0, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

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
            if (this.resizing && this.selectedNotesData) {
                const trueX = Math.min(this.imageWidth, Math.max(0, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
                const trueY = Math.min(this.imageHeight, Math.max(0, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

                const diffX = this.resizeStartX - trueX;
                const diffY = this.resizeStartY - trueY;

                let x1 = this.resizeOriginalX;
                let y1 = this.resizeOriginalY;
                let x2 = x1 + this.resizeOriginalW;
                let y2 = y1 + this.resizeOriginalH;

                switch (this.resizeMode) {
                    case "t":
                        y1 -= diffY;
                        break;
                    case "b":
                        y2 -= diffY;
                        break;
                    case "l":
                        x1 -= diffX;
                        break;
                    case "r":
                        x2 -= diffX;
                        break;
                    case "tl":
                        y1 -= diffY;
                        x1 -= diffX;
                        break;
                    case "tr":
                        y1 -= diffY;
                        x2 -= diffX;
                        break;
                    case "bl":
                        y2 -= diffY;
                        x1 -= diffX;
                        break;
                    case "br":
                        y2 -= diffY;
                        x2 -= diffX;
                        break;
                }

                x1 = Math.min(this.imageWidth, Math.max(0, x1));
                x2 = Math.min(this.imageWidth, Math.max(0, x2));

                y1 = Math.min(this.imageHeight, Math.max(0, y1));
                y2 = Math.min(this.imageHeight, Math.max(0, y2));

                this.selectedNotesData.x = Math.min(x1, x2);
                this.selectedNotesData.y = Math.min(y1, y2);

                this.selectedNotesData.w = Math.max(32, Math.abs(x1 - x2));
                this.selectedNotesData.h = Math.max(32, Math.abs(y1 - y2));
            }
        },

        clickOnNotes: function (notes, e) {
            if (this.contextOpen) {
                return;
            }
            if (!this.editing) {
                return;
            }
            if (this.moving || this.resizing) {
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
            const trueX = Math.max(0, Math.min(this.imageWidth, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
            const trueY = Math.max(0, Math.min(this.imageHeight, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

            this.moving = true;
            this.moveStartX = trueX;
            this.moveStartY = trueY;
            this.moveOriginalX = notes.x;
            this.moveOriginalY = notes.y;
        },

        startResizeNotes: function (notes, e, resizeMode) {
            if (this.contextOpen) {
                return;
            }
            if (!this.editing) {
                return;
            }
            if (this.moving || this.resizing) {
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
            const trueX = Math.max(0, Math.min(this.imageWidth, Math.round(((x - bounds.left) * this.imageWidth) / bounds.width)));
            const trueY = Math.max(0, Math.min(this.imageHeight, Math.round(((y - bounds.top) * this.imageHeight) / bounds.height)));

            this.resizing = true;
            this.resizeMode = resizeMode;
            this.resizeOriginalX = notes.x;
            this.resizeOriginalY = notes.y;
            this.resizeOriginalW = notes.w;
            this.resizeOriginalH = notes.h;

            this.resizeStartX = trueX;
            this.resizeStartY = trueY;
        },

        saveNote: function (note: ImageNote) {
            ImageNotesController.ModifyNote(note);
            this.selectedNotes = -1;
            this.selectedNotesData = null;
        },

        deleteNote: function (note: ImageNote) {
            ImageNotesController.RemoveNote(note);
            this.selectedNotes = -1;
            this.selectedNotesData = null;
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

            this.selectedNotes = note.id;
            this.selectedNotesData = this.notes[this.notes.length - 1];
            this.autoFocus();
        },

        onNotesChange: function (i: number, note: ImageNote) {
            this.notes[i].x = note.x;
            this.notes[i].y = note.y;
            this.notes[i].w = note.w;
            this.notes[i].h = note.h;
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
});
</script>
