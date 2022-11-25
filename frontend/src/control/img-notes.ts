// Subtitles controller

import { MediaAPI } from "@/api/api-media";
import { GetAssetURL, Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { MediaController } from "./media";

export interface ImageNote {
    id: number;
    x: number;
    y: number;
    w: number;
    h: number;
    text: string;
}

function parseImageNotes(json: string): ImageNote[] {
    let o: any;
    try {
        if (typeof json === "string") {
            o = JSON.parse(json);
        } else {
            o = json;
        }
    } catch (ex) {
        console.error(ex);
        return [];
    }
    if (o && Array.isArray(o)) {
        return o.map(note => {
            if (note && typeof note === "object") {
                return {
                    id: ImageNotesController.GetNewId(),
                    x: parseInt(note.x, 10) || 0,
                    y: parseInt(note.y, 10) || 0,
                    w: parseInt(note.w, 10) || 0,
                    h: parseInt(note.h, 10) || 0,
                    text: (note.text || "") + "",
                };
            } else {
                return {
                    id: ImageNotesController.GetNewId(),
                    x: 0,
                    y: 0,
                    w: 0,
                    h: 0,
                    text: "",
                };
            }
        });
    } else {
        return [];
    }
}

export class ImageNotesController {
    public static MediaId = -1;
    public static ImageWidth = 0;
    public static ImageHeight = 0;
    public static NotesFileURL = "";
    public static Notes: ImageNote[] = [];

    public static NextId = 0;

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", ImageNotesController.Load);
        AppEvents.AddEventListener("app-status-update", ImageNotesController.OnMediaChanged);
        AppEvents.AddEventListener("current-media-update", ImageNotesController.Load);

        ImageNotesController.MediaId = AppStatus.CurrentMedia;

        ImageNotesController.Load();
    }

    public static GetNewId(): number {
        ImageNotesController.NextId++;
        return ImageNotesController.NextId;
    }

    public static OnMediaChanged() {
        if (ImageNotesController.MediaId !== AppStatus.CurrentMedia) {
            ImageNotesController.MediaId = AppStatus.CurrentMedia;
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            ImageNotesController.Load();
        }
    }

    public static Load() {
        ImageNotesController.PendingSave = false;
        Request.Abort("notes-save");

        if (!MediaController.MediaData) {
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.ImageWidth = 0;
            ImageNotesController.ImageHeight = 0;
            ImageNotesController.Notes = [];
            AppEvents.Emit("img-notes-update");
            return;
        }

        ImageNotesController.ImageWidth = MediaController.MediaData.width;
        ImageNotesController.ImageHeight = MediaController.MediaData.height;

        if (!MediaController.MediaData.img_notes || !MediaController.MediaData.img_notes_url) {
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            AppEvents.Emit("img-notes-update");
            return;
        }

        ImageNotesController.NotesFileURL = GetAssetURL(MediaController.MediaData.img_notes_url);
        ImageNotesController.Notes = [];

        Timeouts.Abort("img-notes-load");
        Request.Pending("img-notes-load", {
            method: "GET",
            url: ImageNotesController.NotesFileURL,
        }).onSuccess(jsonNotes => {
            ImageNotesController.Notes = parseImageNotes(jsonNotes);
            AppEvents.Emit("img-notes-update");
            console.log(ImageNotesController.Notes);
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add(404, "*", () => {
                    ImageNotesController.Notes = [];
                    AppEvents.Emit("img-notes-update");
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("img-notes-load", 1500, ImageNotesController.Load);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("img-notes-load", 1500, ImageNotesController.Load);
        });
    }

    public static GetNotes(): ImageNote[] {
        return ImageNotesController.Notes.map(n => {
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

    private static PendingSave = false;
    private static Saving = false;

    public static SaveNotes() {
        if (ImageNotesController.Saving) {
            ImageNotesController.PendingSave = true;
            return;
        }

        ImageNotesController.Saving = true;
        ImageNotesController.PendingSave = false;
        const mediaId = ImageNotesController.MediaId;

        Request.Pending("notes-save", MediaAPI.SetNotes(mediaId, ImageNotesController.Notes))
            .onSuccess(() => {
                ImageNotesController.Saving = false;
                if (ImageNotesController.PendingSave) {
                    ImageNotesController.SaveNotes();
                } else {
                    AppEvents.Emit("image-notes-saved");
                }
            })
            .onCancel(() => {
                ImageNotesController.Saving = false;
                ImageNotesController.PendingSave = false;
            })
            .onRequestError((err) => {
                ImageNotesController.Saving = false;
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit("unauthorized");
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
            });
    }


    public static AddNote(x: number, y: number, w: number, h: number) {
        const note: ImageNote = {
            id: ImageNotesController.GetNewId(),
            x: x,
            y: y,
            w: w,
            h: h,
            text: "",
        };

        ImageNotesController.Notes.push(note);

        AppEvents.Emit("img-notes-push", note);

        ImageNotesController.SaveNotes();
    }

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

        if (actualNote.x === note.x && actualNote.y === note.y && actualNote.w === note.w && actualNote.h === note.h && actualNote.text === note.text) {
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

        AppEvents.Emit("img-notes-change", noteIndex, note);

        ImageNotesController.SaveNotes();
    }


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

        AppEvents.Emit("img-notes-rm", noteIndex);

        ImageNotesController.SaveNotes();
    }
}
