// Image notes controller

"use strict";

import { Request } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { BusyStateController } from "./busy-state";
import { MediaController } from "./media";
import { EditMediaAPI } from "@/api/api-media-edit";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "./auth";
import { ImageNote, parseImageNotes } from "@/utils/notes-format";
import { getUniqueNumericId } from "@/utils/unique-id";
import { getAssetURL } from "@/utils/api";

/**
 * Event triggered when the image notes are updated
 */
export const EVENT_NAME_IMAGE_NOTES_UPDATE = "img-notes-update";

/**
 * Event triggered when the image notes are changed
 */
export const EVENT_NAME_IMAGE_NOTES_CHANGE = "img-notes-change";

/**
 * Event triggered when the image notes are saved
 */
export const EVENT_NAME_IMAGE_NOTES_SAVED = "image-notes-saved";

const BUSY_KEY = "image-notes-save";

const REQUEST_KEY_LOAD = "img-notes-load";
const REQUEST_KEY_SAVE = "notes-save";

/**
 * Management object for image notes
 */
export class ImageNotesController {
    /**
     * Current media ID owner of the image notes
     */
    public static MediaId = -1;

    /**
     * Image width
     */
    public static ImageWidth = 0;

    /**
     * Image height
     */
    public static ImageHeight = 0;

    /**
     * URL to the JSON file to download to get the image notes
     */
    public static NotesFileURL = "";

    /**
     * Array of notes
     */
    public static Notes: ImageNote[] = [];

    /**
     * Initialization logic
     */
    public static Initialize() {
        AuthController.AddChangeEventListener(ImageNotesController.Load);
        AppStatus.AddEventListener(ImageNotesController.OnMediaChanged);
        MediaController.AddUpdateEventListener(ImageNotesController.Load);

        ImageNotesController.MediaId = AppStatus.CurrentMedia;

        ImageNotesController.Load();
    }

    /**
     * Called when the app status changed, in order to check if the current media changed
     */
    public static OnMediaChanged() {
        if (ImageNotesController.MediaId !== AppStatus.CurrentMedia) {
            ImageNotesController.MediaId = AppStatus.CurrentMedia;
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            ImageNotesController.Load();
        }
    }

    /**
     * Gets a copy of the image notes array
     * @returns A copy of the image notes array
     */
    public static GetNotes(): ImageNote[] {
        return ImageNotesController.Notes.map((n) => {
            return {
                id: n.id,
                x: n.x,
                y: n.y,
                w: n.w,
                h: n.h,
                text: n.text,
            };
        });
    }

    /**
     * Loads image notes
     */
    public static Load() {
        ImageNotesController.PendingSave = false;
        Request.Abort(REQUEST_KEY_SAVE);

        if (!MediaController.MediaData) {
            clearNamedTimeout(REQUEST_KEY_LOAD);
            Request.Abort(REQUEST_KEY_LOAD);
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.ImageWidth = 0;
            ImageNotesController.ImageHeight = 0;
            ImageNotesController.Notes = [];
            AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_UPDATE);
            return;
        }

        ImageNotesController.ImageWidth = MediaController.MediaData.width;
        ImageNotesController.ImageHeight = MediaController.MediaData.height;

        if (!MediaController.MediaData.img_notes || !MediaController.MediaData.img_notes_url) {
            clearNamedTimeout(REQUEST_KEY_LOAD);
            Request.Abort(REQUEST_KEY_LOAD);
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_UPDATE);
            return;
        }

        ImageNotesController.NotesFileURL = getAssetURL(MediaController.MediaData.img_notes_url);
        ImageNotesController.Notes = [];

        clearNamedTimeout(REQUEST_KEY_LOAD);
        Request.Pending(REQUEST_KEY_LOAD, {
            method: "GET",
            url: ImageNotesController.NotesFileURL,
        })
            .onSuccess((jsonNotes) => {
                ImageNotesController.Notes = parseImageNotes(jsonNotes);
                AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_UPDATE);
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        ImageNotesController.Notes = [];
                        AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_UPDATE);
                    })
                    .add("*", "*", () => {
                        // Retry
                        setNamedTimeout(REQUEST_KEY_LOAD, 1500, ImageNotesController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_KEY_LOAD, 1500, ImageNotesController.Load);
            });
    }

    /**
     * Pending save flag. If true, the notes must save after the current save request ends.
     */
    private static PendingSave = false;

    /**
     * True if there is a pending save request.
     */
    private static Saving = false;

    /**
     * Saves the image notes
     */
    public static SaveNotes() {
        if (ImageNotesController.Saving) {
            ImageNotesController.PendingSave = true;
            return;
        }

        ImageNotesController.Saving = true;
        BusyStateController.SetBusy(BUSY_KEY);
        ImageNotesController.PendingSave = false;
        const mediaId = ImageNotesController.MediaId;

        Request.Pending(REQUEST_KEY_SAVE, EditMediaAPI.SetNotes(mediaId, ImageNotesController.Notes))
            .onSuccess(() => {
                ImageNotesController.Saving = false;
                BusyStateController.RemoveBusy(BUSY_KEY);
                if (ImageNotesController.PendingSave) {
                    ImageNotesController.SaveNotes();
                } else {
                    AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_SAVED);
                }
            })
            .onCancel(() => {
                ImageNotesController.Saving = false;
                ImageNotesController.PendingSave = false;
                BusyStateController.RemoveBusy(BUSY_KEY);
            })
            .onRequestError((err) => {
                ImageNotesController.Saving = false;
                BusyStateController.RemoveBusy(BUSY_KEY);
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(403, "*", () => {
                        ImageNotesController.PendingSave = false;
                    })
                    .add(404, "*", () => {
                        ImageNotesController.PendingSave = false;
                    })
                    .add(500, "*", () => {
                        ImageNotesController.SaveNotes();
                    })
                    .add("*", "*", () => {
                        ImageNotesController.SaveNotes();
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                ImageNotesController.Saving = false;
                ImageNotesController.PendingSave = false;
                BusyStateController.RemoveBusy(BUSY_KEY);
            });
    }

    /**
     * Adds image note
     * @param x X coordinate
     * @param y Y coordinate
     * @param w Width
     * @param h Height
     */
    public static AddNote(x: number, y: number, w: number, h: number) {
        const note: ImageNote = {
            id: getUniqueNumericId(),
            x: x,
            y: y,
            w: w,
            h: h,
            text: "",
        };

        ImageNotesController.Notes.push(note);

        ImageNotesController.Emit("push", note);

        ImageNotesController.SaveNotes();
    }

    /**
     * Modifies a note
     * @param note The modified note
     */
    public static ModifyNote(note: ImageNote) {
        let noteIndex = -1;
        for (let i = 0; i < ImageNotesController.Notes.length; i++) {
            if (ImageNotesController.Notes[i].id === note.id) {
                noteIndex = i;
                break;
            }
        }

        if (noteIndex === -1) {
            return;
        }

        const actualNote = ImageNotesController.Notes[noteIndex];

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

        ImageNotesController.Notes[noteIndex] = {
            id: note.id,
            x: note.x,
            y: note.y,
            w: note.w,
            h: note.h,
            text: note.text,
        };

        ImageNotesController.Emit("update", note, noteIndex);

        ImageNotesController.SaveNotes();
    }

    /**
     * Removes a note
     * @param note The note to remove
     */
    public static RemoveNote(note: ImageNote) {
        let noteIndex = -1;
        for (let i = 0; i < ImageNotesController.Notes.length; i++) {
            if (ImageNotesController.Notes[i].id === note.id) {
                noteIndex = i;
                break;
            }
        }

        if (noteIndex === -1) {
            return;
        }

        ImageNotesController.Notes.splice(noteIndex, 1);

        ImageNotesController.Emit("rm", null, noteIndex);

        ImageNotesController.SaveNotes();
    }

    /**
     * Emits list update event
     * @param mode Mode
     * @param note Note
     * @param index Index
     */
    private static Emit(mode: "push" | "rm" | "update", note?: ImageNote, index?: number) {
        AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_CHANGE, mode, note, index);
    }

    /**
     * Adds event listener to check for updates
     * @param handler Event handler
     */
    public static AddEventListener(handler: (mode: "push" | "rm" | "update", note?: ImageNote, index?: number) => void) {
        AppEvents.AddEventListener(EVENT_NAME_IMAGE_NOTES_CHANGE, handler);
    }

    /**
     * Removes event listener
     * @param handler Event handler
     */
    public static RemoveEventListener(handler: (mode: "push" | "rm" | "update", note?: ImageNote, index?: number) => void) {
        AppEvents.RemoveEventListener(EVENT_NAME_IMAGE_NOTES_CHANGE, handler);
    }
}

ImageNotesController.Initialize();
