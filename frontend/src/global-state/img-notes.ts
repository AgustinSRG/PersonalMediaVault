// Image notes global state

"use strict";

import { RequestErrorHandler, abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import type { ImageNote } from "@/utils/notes-format";
import { parseImageNotes } from "@/utils/notes-format";
import { getUniqueNumericId, getUniqueStringId } from "@/utils/unique-id";
import { getAssetURL } from "@/utils/api";
import { apiMediaSetNotes } from "@/api/api-media-edit";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_NAV_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_IMAGE_NOTES_CHANGE,
    EVENT_NAME_IMAGE_NOTES_SAVED,
    EVENT_NAME_IMAGE_NOTES_UPDATE,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";
import { removeGlobalBusyState, setGlobalBusyState } from "./busy-state";
import { getCurrentMediaData, modifyCurrentMediaData } from "./media";
import { LOAD_RETRY_DELAY } from "@/constants";
import { getNavigationStatus } from "./navigation";

/**
 * The change type for the image nodes
 *  - 'push': New note added
 *  - 'rm': Note removed
 *  - 'update': Note updated
 */
export type ImageNodesChangeType = "push" | "rm" | "update";

/**
 * Image notes state
 */
const ImageNotesState = {
    /**
     * ID of the media owner of the image notes
     */
    mediaId: getNavigationStatus().media,

    /**
     * Image width (px)
     */
    imageWidth: 0,

    /**
     * Image height (px)
     */
    imageHeight: 0,

    /**
     * Notes URL
     */
    url: "",

    /**
     * List of image notes
     */
    notes: [] as ImageNote[],

    /**
     * True if the image notes need to be saved
     */
    pendingSave: false,

    /**
     * True if saving the image notes
     */
    saving: false,
};

/**
 * Gets the ID of the media for which the image notes are being loaded
 * @returns The media ID
 */
export function getImageNotesMediaId(): number {
    return ImageNotesState.mediaId;
}

/**
 * Gets image width for the current image notes
 * @returns The width (px)
 */
export function getImageNotesWidth(): number {
    return ImageNotesState.imageWidth;
}

/**
 * Gets image height for the current image notes
 * @returns The height (px)
 */
export function getImageNotesHeight(): number {
    return ImageNotesState.imageHeight;
}

/**
 * Gets the current image notes array
 * @returns The current image notes array
 */
export function getImageNotes(): ImageNote[] {
    return ImageNotesState.notes;
}

/**
 * Sets image notes
 * @param notes The list of image notes
 */
export function setImageNotes(notes: ImageNote[]) {
    ImageNotesState.notes = notes;
    emitAppEvent(EVENT_NAME_IMAGE_NOTES_UPDATE, ImageNotesState.notes, ImageNotesState.imageWidth, ImageNotesState.imageHeight);
}

// Request ID for loading
const REQUEST_KEY_LOAD = getUniqueStringId();

/**
 * Loads image notes
 */
function loadImageNotes() {
    ImageNotesState.pendingSave = false;
    clearNamedTimeout(REQUEST_KEY_SAVE);
    abortNamedApiRequest(REQUEST_KEY_SAVE);

    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        clearNamedTimeout(REQUEST_KEY_LOAD);
        abortNamedApiRequest(REQUEST_KEY_LOAD);

        ImageNotesState.url = "";
        ImageNotesState.imageWidth = 0;
        ImageNotesState.imageHeight = 0;

        setImageNotes([]);
        return;
    }

    ImageNotesState.imageWidth = mediaData.width;
    ImageNotesState.imageHeight = mediaData.height;

    if (!mediaData.img_notes || !mediaData.img_notes_url) {
        clearNamedTimeout(REQUEST_KEY_LOAD);
        abortNamedApiRequest(REQUEST_KEY_LOAD);

        ImageNotesState.url = "";

        setImageNotes([]);
        return;
    }

    ImageNotesState.url = getAssetURL(mediaData.img_notes_url);
    ImageNotesState.notes = [];

    clearNamedTimeout(REQUEST_KEY_LOAD);
    makeNamedApiRequest(REQUEST_KEY_LOAD, {
        method: "GET",
        url: ImageNotesState.url,
    })
        .onSuccess((jsonNotes) => {
            setImageNotes(parseImageNotes(jsonNotes));
        })
        .onRequestError((err) => {
            new RequestErrorHandler()
                .add(401, "*", () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                })
                .add(404, "*", () => {
                    setImageNotes([]);
                })
                .add("*", "*", () => {
                    // Retry
                    setNamedTimeout(REQUEST_KEY_LOAD, LOAD_RETRY_DELAY, loadImageNotes);
                })
                .handle(err);
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_KEY_LOAD, LOAD_RETRY_DELAY, loadImageNotes);
        });
}

// Key for busy state
const BUSY_KEY = "image-notes-save";

// Request ID for saving
const REQUEST_KEY_SAVE = getUniqueStringId();

/**
 * Saves the image notes
 */
function saveImageNotes() {
    if (ImageNotesState.saving) {
        ImageNotesState.pendingSave = true;
        return;
    }

    ImageNotesState.saving = true;

    setGlobalBusyState(BUSY_KEY);

    ImageNotesState.pendingSave = false;

    const mediaId = ImageNotesState.mediaId;

    clearNamedTimeout(REQUEST_KEY_SAVE);

    makeNamedApiRequest(REQUEST_KEY_SAVE, apiMediaSetNotes(mediaId, ImageNotesState.notes))
        .onSuccess((res) => {
            ImageNotesState.saving = false;

            removeGlobalBusyState(BUSY_KEY);

            if (ImageNotesState.mediaId === mediaId) {
                ImageNotesState.url = res.url || "";
            }

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.img_notes_url = res.url || "";
                metadata.img_notes = !!res.url;
            });

            if (ImageNotesState.pendingSave) {
                saveImageNotes();
            } else {
                emitAppEvent(EVENT_NAME_IMAGE_NOTES_SAVED);
            }
        })
        .onCancel(() => {
            ImageNotesState.saving = false;
            ImageNotesState.pendingSave = false;
            removeGlobalBusyState(BUSY_KEY);
        })
        .onRequestError((err, handleErr) => {
            ImageNotesState.saving = false;

            removeGlobalBusyState(BUSY_KEY);

            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                badRequest: () => {
                    ImageNotesState.pendingSave = false;
                },
                accessDenied: () => {
                    ImageNotesState.pendingSave = false;
                },
                notFound: () => {
                    ImageNotesState.pendingSave = false;
                },
                temporalError: () => {
                    setNamedTimeout(REQUEST_KEY_SAVE, LOAD_RETRY_DELAY, saveImageNotes);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            ImageNotesState.saving = false;
            ImageNotesState.pendingSave = false;
            removeGlobalBusyState(BUSY_KEY);
        });
}

/**
 * Emits list update event
 * @param type Change type
 * @param note Note
 * @param index Index
 */
function emit(type: ImageNodesChangeType, note: ImageNote | null, index?: number) {
    emitAppEvent(EVENT_NAME_IMAGE_NOTES_CHANGE, type, note, index);
}

/**
 * Adds image note
 * @param x X coordinate
 * @param y Y coordinate
 * @param w Width
 * @param h Height
 */
export function addImageNote(x: number, y: number, w: number, h: number) {
    const note: ImageNote = {
        id: getUniqueNumericId(),
        x: x,
        y: y,
        w: w,
        h: h,
        text: "",
    };

    ImageNotesState.notes.push(note);

    emit("push", note);

    saveImageNotes();
}

/**
 * Modifies a note
 * @param note The modified note
 */
export function modifyImageNote(note: ImageNote) {
    let noteIndex = -1;
    for (let i = 0; i < ImageNotesState.notes.length; i++) {
        if (ImageNotesState.notes[i].id === note.id) {
            noteIndex = i;
            break;
        }
    }

    if (noteIndex === -1) {
        return;
    }

    const actualNote = ImageNotesState.notes[noteIndex];

    if (note.id !== actualNote.id) {
        return;
    }

    if (
        actualNote.x === note.x &&
        actualNote.y === note.y &&
        actualNote.w === note.w &&
        actualNote.h === note.h &&
        actualNote.text === note.text
    ) {
        return; // Nothing changed
    }

    ImageNotesState.notes[noteIndex] = {
        id: note.id,
        x: note.x,
        y: note.y,
        w: note.w,
        h: note.h,
        text: note.text,
    };

    emit("update", note, noteIndex);

    saveImageNotes();
}

/**
 * Removes a note
 * @param note The note to remove
 */
export function removeImageNote(note: ImageNote) {
    let noteIndex = -1;
    for (let i = 0; i < ImageNotesState.notes.length; i++) {
        if (ImageNotesState.notes[i].id === note.id) {
            noteIndex = i;
            break;
        }
    }

    if (noteIndex === -1) {
        return;
    }

    ImageNotesState.notes.splice(noteIndex, 1);

    emit("rm", null, noteIndex);

    saveImageNotes();
}

addAppEventListener(EVENT_NAME_AUTH_CHANGED, loadImageNotes);
addAppEventListener(EVENT_NAME_MEDIA_UPDATE, loadImageNotes);

addAppEventListener(EVENT_NAME_NAV_STATUS_CHANGED, (navStatus) => {
    if (ImageNotesState.mediaId !== navStatus.media) {
        ImageNotesState.mediaId = navStatus.media;
        ImageNotesState.url = "";
        ImageNotesState.notes = [];

        loadImageNotes();
    }
});

// Initially load image notes
loadImageNotes();
