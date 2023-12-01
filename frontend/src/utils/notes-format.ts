// Image notes formatting

"use strict";

import { getUniqueNumericId } from "./unique-id";

/**
 * Image note
 */
export interface ImageNote {
    /**
     * Unique identifier
     */
    id: number;

    /**
     * X coordinate (css: top)
     */
    x: number;

    /**
     * Y coordinate (css: left)
     */
    y: number;

    /**
     * Width
     */
    w: number;

    /**
     * Height
     */
    h: number;

    /**
     * Text to display
     */
    text: string;
}

/**
 * Parses image notes from a JSON string
 * @param json The JSON string
 * @returns The image notes array
 */
export function parseImageNotes(json: unknown): ImageNote[] {
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
        return o.map((note) => {
            if (note && typeof note === "object") {
                return {
                    id: getUniqueNumericId(),
                    x: parseInt(note.x, 10) || 0,
                    y: parseInt(note.y, 10) || 0,
                    w: parseInt(note.w, 10) || 0,
                    h: parseInt(note.h, 10) || 0,
                    text: (note.text || "") + "",
                };
            } else {
                return {
                    id: getUniqueNumericId(),
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

export const NOTES_TEXT_SEPARATOR = '"""';

/**
 * Turns notes into text format
 * @param notes The notes
 * @returns notes as text
 */
export function imageNotesToText(notes: ImageNote[]): string {
    return notes
        .map((note) => {
            const noteHeader = `[${note.x}, ${note.y}] (${note.w} x ${note.h})`;

            const lines = note.text.split("\n").map((line) => {
                if (/^""["]+$/.test(line.trim())) {
                    line = line.trim() + '"';
                }

                return line;
            });

            return noteHeader + "\n" + NOTES_TEXT_SEPARATOR + "\n" + lines.join("\n") + "\n" + NOTES_TEXT_SEPARATOR + "\n";
        })
        .join("\n");
}

/**
 * Turns text format into image notes
 * @param text The text
 * @returns The image notes list
 */
export function textToImageNotes(text: string): ImageNote[] {
    const lines = text.split("\n");

    const notes: ImageNote[] = [];

    let currentNote: ImageNote | null = null;

    let state = 0;

    for (const line of lines) {
        switch (state) {
            case 0:
                {
                    const trimLine = line.trim();

                    if (!trimLine) {
                        continue;
                    }

                    if (trimLine.startsWith("[")) {
                        const parts = trimLine.substring(1).split("]");

                        const partsCoords = (parts[0] || "").split(",");
                        const x = parseInt(partsCoords[0] || "", 10) || 0;
                        const y = parseInt(partsCoords[1] || "", 10) || 0;

                        const partsSize = (parts[1] || "").replace(/[\(\)]/g, "").split("x");
                        const w = parseInt(partsSize[0] || "", 10) || 0;
                        const h = parseInt(partsSize[1] || "", 10) || 0;

                        currentNote = {
                            id: getUniqueNumericId(),
                            text: "",
                            x: x,
                            y: y,
                            w: w,
                            h: h,
                        };

                        state = 1;
                    }
                }
                break;
            case 1:
                {
                    const trimLine = line.trim();

                    if (!trimLine) {
                        continue;
                    }

                    if (trimLine === NOTES_TEXT_SEPARATOR) {
                        state = 2;
                    } else {
                        currentNote.text = line;
                        state = 2;
                    }
                }
                break;
            case 2:
                {
                    const trimLine = line.trim();

                    if (trimLine === NOTES_TEXT_SEPARATOR) {
                        state = 0;
                        notes.push(currentNote);
                    } else if (/^""["]+$/.test(trimLine)) {
                        if (currentNote.text) {
                            currentNote.text += "\n";
                        }

                        currentNote.text += trimLine.substring(1);
                    } else {
                        if (currentNote.text) {
                            currentNote.text += "\n";
                        }

                        currentNote.text += line;
                    }
                }
                break;
        }
    }

    return notes;
}
