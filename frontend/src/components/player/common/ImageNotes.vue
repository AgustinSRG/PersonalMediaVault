<template>
    <div
        ref="container"
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
            @mousedown="startMovingMouse(note, $event)"
            @touchstart.passive="startMovingTouch(note, $event)"
        >
            <div
                v-if="editing"
                class="resize resize-left"
                @mousedown="startResizingMouse(note, $event, 'l')"
                @touchstart.passive="startResizingTouch(note, $event, 'l')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-top"
                @mousedown="startResizingMouse(note, $event, 't')"
                @touchstart.passive="startResizingTouch(note, $event, 't')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-right"
                @mousedown="startResizingMouse(note, $event, 'r')"
                @touchstart.passive="startResizingTouch(note, $event, 'r')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-bottom"
                @mousedown="startResizingMouse(note, $event, 'b')"
                @touchstart.passive="startResizingTouch(note, $event, 'b')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-top-left"
                @mousedown="startResizingMouse(note, $event, 'tl')"
                @touchstart.passive="startResizingTouch(note, $event, 'tl')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-top-right"
                @mousedown="startResizingMouse(note, $event, 'tr')"
                @touchstart.passive="startResizingTouch(note, $event, 'tr')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-bottom-left"
                @mousedown="startResizingMouse(note, $event, 'bl')"
                @touchstart.passive="startResizingTouch(note, $event, 'bl')"
            ></div>
            <div
                v-if="editing"
                class="resize resize-corner-bottom-right"
                @mousedown="startResizingMouse(note, $event, 'br')"
                @touchstart.passive="startResizingTouch(note, $event, 'br')"
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

<script setup lang="ts">
import { nextTick, onMounted, ref, useTemplateRef, watch } from "vue";
import { ImageNotesController } from "@/control/img-notes";
import { escapeHTML } from "@/utils/html";
import type { ImageNote } from "@/utils/notes-format";
import type { PositionEvent } from "@/utils/position-event";
import { positionEventFromMouseEvent, positionEventFromTouchEvent } from "@/utils/position-event";
import { EVENT_NAME_IMAGE_NOTES_UPDATE, EVENT_NAME_IMAGE_NOTES_CHANGE } from "@/control/app-events";
import { useI18n } from "@/composables/use-i18n";
import { onApplicationEvent } from "@/composables/on-app-event";
import { onDocumentEvent } from "@/composables/on-document-event";
import { stopPropagationEvent } from "@/utils/events";

/**
 * Maps a dimension based in the image dimensions
 * @param dim The dimension
 * @param minDim The min dimension
 * @param maxDim The max dimension
 * @param imgDim The image dimension
 * @returns The mapped dimension
 */
const mapDim = (dim: number, minDim: number, maxDim: number, imgDim: number) => {
    return Math.min(maxDim, Math.max(minDim, Math.round((dim * maxDim) / imgDim))) + "px";
};

/**
 * Escapes image notes text to avoid HTML injection
 * @param txt The original text
 * @returns The escaped text
 */
const escapeText = (txt: string): string => {
    return escapeHTML(txt).replace(/\n/g, "<br>");
};

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
    /**
     * Top coordinate
     */
    top: String,
    /**
     * Left coordinate
     */
    left: String,

    /**
     * Width dimension
     */
    width: String,

    /**
     * Height dimension
     */
    height: String,

    /**
     * True if visible
     */
    visible: Boolean,

    /**
     * True if editing
     */
    editing: Boolean,

    /**
     * True if context menu is opened
     */
    contextOpen: Boolean,
});

// List of image notes
const notes = ref<ImageNote[]>(ImageNotesController.GetNotes());

// Image dimensions (pixels)
const imageWidth = ref(Math.max(1, ImageNotesController.ImageWidth));
const imageHeight = ref(Math.max(1, ImageNotesController.ImageHeight));

onApplicationEvent(EVENT_NAME_IMAGE_NOTES_UPDATE, () => {
    notes.value = ImageNotesController.GetNotes();
    imageWidth.value = Math.max(1, ImageNotesController.ImageWidth);
    imageHeight.value = Math.max(1, ImageNotesController.ImageHeight);
});

// Id of the selected note
const selectedNotes = ref(-1);

// Data of the selected note
const selectedNotesData = ref<ImageNote | null>(null);

/**
 * Called when a note is pushed
 * @param note The note
 */
const onNotesPush = (note: ImageNote) => {
    notes.value.push({
        id: note.id,
        x: note.x,
        y: note.y,
        w: note.w,
        h: note.h,
        text: note.text,
    });

    selectedNotes.value = note.id;
    selectedNotesData.value = notes.value[notes.value.length - 1];
    autoFocus();
};

/**
 * Called when anote changes
 * @param i The note index
 * @param note The new note data
 */
const onNotesChange = (i: number, note: ImageNote) => {
    notes.value[i].x = note.x;
    notes.value[i].y = note.y;
    notes.value[i].w = note.w;
    notes.value[i].h = note.h;
};

/**
 * Called when a note is removed
 * @param i The note index
 */
const onNotesRemove = (i: number) => {
    notes.value.splice(i, 1);
};

onApplicationEvent(EVENT_NAME_IMAGE_NOTES_CHANGE, (mode: "push" | "rm" | "update", note?: ImageNote, index?: number) => {
    switch (mode) {
        case "push":
            onNotesPush(note);
            break;
        case "rm":
            onNotesRemove(index);
            break;
        case "update":
            onNotesChange(index, note);
            break;
    }
});

// Real dimensions (pixels)
const realWidth = ref(1);
const realHeight = ref(1);

/**
 * Updates the real dimensions of the container
 */
const updateRealDimensions = () => {
    nextTick(() => {
        if (!container.value) {
            return;
        }

        const bounds = container.value.getBoundingClientRect();
        realWidth.value = bounds.width;
        realHeight.value = bounds.height;
    });
};

onMounted(updateRealDimensions);

watch([() => props.top, () => props.left, () => props.width, () => props.height], updateRealDimensions);

// True if adding anote
const adding = ref(false);

// Initial coordinated for the adding note
const addStartX = ref(0);
const addStartY = ref(0);

// Coordinated of the adding note
const addX = ref(0);
const addY = ref(0);
const addW = ref(0);
const addH = ref(0);

// True if moving anote
const moving = ref(false);

// Moving coordinates
const moveOriginalX = ref(0);
const moveOriginalY = ref(0);
const moveStartX = ref(0);
const moveStartY = ref(0);

// True if resizing a note
const resizing = ref(false);

// Resizing coordinates
const resizeOriginalX = ref(0);
const resizeOriginalY = ref(0);
const resizeOriginalW = ref(0);
const resizeOriginalH = ref(0);
const resizeStartX = ref(0);
const resizeStartY = ref(0);

// Resize modes. Depending in from where is being resized
type ResizeMode = "" | "t" | "b" | "l" | "r" | "tl" | "tr" | "bl" | "br";

// Resize mode
const resizeMode = ref<ResizeMode>("");

// Selected note to display
const selectedNote = ref<ImageNote | null>(null);

// Hover coordinates
const hoverRight = ref("");
const hoverLeft = ref("");
const hoverTop = ref("");
const hoverBottom = ref("");

// True if hover is pinned
const hoverPinned = ref(false);

/**
 * Called when the user moves inside a note container,
 * in order to display it.
 * @param e The position event
 */
const onMoveInsideNotesContainer = (e: PositionEvent) => {
    if (!container.value) {
        return;
    }

    const x = e.x;
    const y = e.y;

    const bounds = container.value.getBoundingClientRect();

    const realY = y - bounds.top;
    const realX = x - bounds.left;

    const trueX = Math.max(0, Math.min(imageWidth.value - 32, Math.round(((x - bounds.left) * imageWidth.value) / bounds.width)));
    const trueY = Math.max(0, Math.min(imageHeight.value - 32, Math.round(((y - bounds.top) * imageHeight.value) / bounds.height)));

    for (const note of notes.value) {
        if (hoverPinned.value && selectedNote.value === note) {
            continue;
        }
        if (trueX >= note.x && trueX <= note.x + note.w) {
            if (trueY >= note.y && trueY <= note.y + note.h) {
                selectedNote.value = note;

                // Position mouse

                if (realY < bounds.height / 2) {
                    hoverTop.value = realY + 8 + "px";
                    hoverBottom.value = "";
                } else {
                    hoverTop.value = "";
                    hoverBottom.value = bounds.height - realY + 8 + "px";
                }

                if (realX < bounds.width / 2) {
                    hoverLeft.value = realX + 8 + "px";
                    hoverRight.value = "";
                } else {
                    hoverLeft.value = "";
                    hoverRight.value = bounds.width - realX + 8 + "px";
                }

                hoverPinned.value = false;

                return;
            }
        }
    }

    if (!hoverPinned.value) {
        selectedNote.value = null;
    }
};

/**
 * Called when the user click on a note
 * @param e The click event
 */
const onClickFindNotes = (e: PositionEvent) => {
    hoverPinned.value = false;
    onMoveInsideNotesContainer(e);
    hoverPinned.value = !!selectedNote.value;
};

/**
 * Called when the user hovers a note
 * @param e The mouse event
 */
const onMouseMoveInsideContainer = (e: MouseEvent) => {
    onMoveInsideNotesContainer(positionEventFromMouseEvent(e));
};

/**
 * Called when the mouse leaves a container
 */
const onMouseLeave = () => {
    if (hoverPinned.value) {
        return;
    }
    selectedNote.value = null;
};

/**
 * Called when the user clicks in the container,
 * in order to add or, if no edit mode, to select
 * an image note.
 * @param e The position event
 */
const startAdding = (e: PositionEvent) => {
    if (!container.value) {
        return;
    }

    if (!props.editing) {
        onClickFindNotes(e);
        return;
    }

    if (props.contextOpen) {
        return;
    }

    if (selectedNotesData.value) {
        selectedNotesData.value = null;
        selectedNotes.value = -1;
        return;
    }

    e.e.stopPropagation();

    const bounds = container.value.getBoundingClientRect();

    const x = e.x;
    const y = e.y;

    const trueX = Math.max(0, Math.min(imageWidth.value - 32, Math.round(((x - bounds.left) * imageWidth.value) / bounds.width)));
    const trueY = Math.max(0, Math.min(imageHeight.value - 32, Math.round(((y - bounds.top) * imageHeight.value) / bounds.height)));

    adding.value = true;
    addStartX.value = trueX;
    addStartY.value = trueY;

    addX.value = trueX;
    addY.value = trueY;
    addW.value = 32;
    addH.value = 32;
};

/**
 * Starts adding a note with the mouse
 * @param e The mouse event
 */
const startAddingMouse = (e: MouseEvent) => {
    if (e.button !== 0) {
        return; // Not the main button
    }

    startAdding(positionEventFromMouseEvent(e));
};

/**
 * Starts adding a note with the touch screen
 * @param e The touch screen
 */
const startAddingTouch = (e: TouchEvent) => {
    startAdding(positionEventFromTouchEvent(e));
};

/**
 * Start moving an image note
 * @param notes The note
 * @param e The position event
 */
const startMoving = (notes: ImageNote, e: PositionEvent) => {
    if (!container.value) {
        return;
    }

    if (props.contextOpen) {
        return;
    }

    if (!props.editing) {
        return;
    }

    if (moving.value || resizing.value) {
        return;
    }

    e.e.stopPropagation();

    selectedNotes.value = notes.id;
    selectedNotesData.value = notes;

    const bounds = container.value.getBoundingClientRect();

    const x = e.x;
    const y = e.y;

    const trueX = Math.max(0, Math.min(imageWidth.value, Math.round(((x - bounds.left) * imageWidth.value) / bounds.width)));
    const trueY = Math.max(0, Math.min(imageHeight.value, Math.round(((y - bounds.top) * imageHeight.value) / bounds.height)));

    moving.value = true;
    moveStartX.value = trueX;
    moveStartY.value = trueY;
    moveOriginalX.value = notes.x;
    moveOriginalY.value = notes.y;
};

/**
 * Starts moving a note with the mouse
 * @param notes The note
 * @param e The mouse event
 */
const startMovingMouse = (notes: ImageNote, e: MouseEvent) => {
    if (e.button !== 0) {
        return; // Not the main button
    }

    startMoving(notes, positionEventFromMouseEvent(e));
};

/**
 * Starts moving a note with the touch screen
 * @param notes The note
 * @param e The touch screen
 */
const startMovingTouch = (notes: ImageNote, e: TouchEvent) => {
    startMoving(notes, positionEventFromTouchEvent(e));
};

/**
 * Starts resizing a note
 * @param notes The note
 * @param e The position event
 * @param rm The resize mode
 */
const startResizing = (notes: ImageNote, e: PositionEvent, rm: ResizeMode) => {
    if (!container.value) {
        return;
    }

    if (props.contextOpen) {
        return;
    }

    if (!props.editing) {
        return;
    }

    if (moving.value || resizing.value) {
        return;
    }

    e.e.stopPropagation();

    selectedNotes.value = notes.id;
    selectedNotesData.value = notes;

    const bounds = container.value.getBoundingClientRect();

    const x = e.x;
    const y = e.y;

    const trueX = Math.max(0, Math.min(imageWidth.value, Math.round(((x - bounds.left) * imageWidth.value) / bounds.width)));
    const trueY = Math.max(0, Math.min(imageHeight.value, Math.round(((y - bounds.top) * imageHeight.value) / bounds.height)));

    resizing.value = true;
    resizeMode.value = rm;
    resizeOriginalX.value = notes.x;
    resizeOriginalY.value = notes.y;
    resizeOriginalW.value = notes.w;
    resizeOriginalH.value = notes.h;

    resizeStartX.value = trueX;
    resizeStartY.value = trueY;
};

/**
 * Starts resizing a note with the mouse
 * @param notes The note
 * @param e The mouse event
 * @param resizeMode The resize mode
 */
const startResizingMouse = (notes: ImageNote, e: MouseEvent, resizeMode: ResizeMode) => {
    if (e.button !== 0) {
        return; // Not the main button
    }

    startResizing(notes, positionEventFromMouseEvent(e), resizeMode);
};

/**
 * Starts resizing a note with the touch screen
 * @param notes The note
 * @param e The touch screen
 * @param resizeMode The resize mode
 */
const startResizingTouch = (notes: ImageNote, e: TouchEvent, resizeMode: ResizeMode) => {
    startResizing(notes, positionEventFromTouchEvent(e), resizeMode);
};

/**
 * Called when the user moved the position.
 * Depending on the mode, it will do one thing or another.
 * @param e The position event
 */
const move = (e: PositionEvent) => {
    if (!container.value) {
        return;
    }

    if (!adding.value && !moving.value && !resizing.value) {
        return;
    }
    const bounds = container.value.getBoundingClientRect();

    const x = e.x;
    const y = e.y;

    const trueX = Math.min(imageWidth.value, Math.max(0, Math.round(((x - bounds.left) * imageWidth.value) / bounds.width)));
    const trueY = Math.min(imageHeight.value, Math.max(0, Math.round(((y - bounds.top) * imageHeight.value) / bounds.height)));

    if (adding.value) {
        if (trueX - addStartX.value > 0) {
            addX.value = addStartX.value;
            addW.value = Math.max(32, trueX - addStartX.value);
        } else {
            addX.value = Math.max(0, trueX);
            addW.value = Math.max(32, addStartX.value - trueX);
        }

        if (trueY - addStartY.value > 0) {
            addY.value = addStartY.value;
            addH.value = trueY - addStartY.value;
        } else {
            addY.value = Math.max(0, trueY);
            addH.value = Math.max(32, addStartY.value - trueY);
        }
    }

    if (moving.value && selectedNotesData.value) {
        const diffX = moveStartX.value - trueX;

        selectedNotesData.value.x = Math.max(0, moveOriginalX.value - diffX);

        if (selectedNotesData.value.x + selectedNotesData.value.w > imageWidth.value) {
            selectedNotesData.value.x = Math.max(0, imageWidth.value - selectedNotesData.value.w);
        }

        const diffY = moveStartY.value - trueY;

        selectedNotesData.value.y = Math.max(0, moveOriginalY.value - diffY);

        if (selectedNotesData.value.y + selectedNotesData.value.h > imageHeight.value) {
            selectedNotesData.value.y = Math.max(0, imageHeight.value - selectedNotesData.value.h);
        }
    }

    if (resizing.value && selectedNotesData.value) {
        const diffX = resizeStartX.value - trueX;
        const diffY = resizeStartY.value - trueY;

        let x1 = resizeOriginalX.value;
        let y1 = resizeOriginalY.value;
        let x2 = x1 + resizeOriginalW.value;
        let y2 = y1 + resizeOriginalH.value;

        switch (resizeMode.value) {
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

        x1 = Math.min(imageWidth.value, Math.max(0, x1));
        x2 = Math.min(imageWidth.value, Math.max(0, x2));

        y1 = Math.min(imageHeight.value, Math.max(0, y1));
        y2 = Math.min(imageHeight.value, Math.max(0, y2));

        selectedNotesData.value.x = Math.min(x1, x2);
        selectedNotesData.value.y = Math.min(y1, y2);

        selectedNotesData.value.w = Math.max(32, Math.abs(x1 - x2));
        selectedNotesData.value.h = Math.max(32, Math.abs(y1 - y2));
    }
};

onDocumentEvent("mousemove", (e: MouseEvent) => {
    move(positionEventFromMouseEvent(e));
});

onDocumentEvent("touchmove", (e: TouchEvent) => {
    move(positionEventFromTouchEvent(e));
});

/**
 * Called when the mouse or touch is dropped.
 * This will stop moving, adding or resizing.
 */
const drop = () => {
    if (!adding.value && !moving.value && !resizing.value) {
        return;
    }

    if (adding.value) {
        adding.value = false;

        ImageNotesController.AddNote(addX.value, addY.value, addW.value, addH.value);
    }

    if (moving.value) {
        moving.value = false;

        if (selectedNotesData.value) {
            ImageNotesController.ModifyNote(selectedNotesData.value);
        }

        autoFocus();
    }

    if (resizing.value) {
        resizing.value = false;

        if (selectedNotesData.value) {
            ImageNotesController.ModifyNote(selectedNotesData.value);
        }

        autoFocus();
    }
};

onDocumentEvent("mouseup", drop);
onDocumentEvent("touchend", drop);

/**
 * Saves a note
 * @param note The note
 */
const saveNote = (note: ImageNote) => {
    ImageNotesController.ModifyNote(note);

    selectedNotes.value = -1;
    selectedNotesData.value = null;
};

/**
 * Deletes a note
 * @param note The note
 */
const deleteNote = (note: ImageNote) => {
    ImageNotesController.RemoveNote(note);

    selectedNotes.value = -1;
    selectedNotesData.value = null;
};

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    nextTick(() => {
        const autoFocusElement = container.value?.querySelector(".auto-focus") as HTMLElement;

        if (autoFocusElement) {
            autoFocusElement.focus();
        }
    });
};
</script>
