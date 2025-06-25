// Time slices utils

"use strict";

import type { MediaTimeSlice } from "@/api/models";
import { renderTimeSeconds } from "./time";

export function parseTimeSeconds(str: string): number {
    const parts = (str || "").trim().split(":");

    let h = 0;
    let m = 0;
    let s = 0;

    if (parts.length > 2) {
        h = parseInt(parts[0], 10) || 0;
        m = parseInt(parts[1], 10) || 0;
        s = parseInt(parts[2], 10) || 0;
    } else if (parts.length > 1) {
        m = parseInt(parts[0], 10) || 0;
        s = parseInt(parts[1], 10) || 0;
    } else {
        s = parseInt(parts[0], 10) || 0;
    }

    return h * 3600 + m * 60 + s;
}

export function renderTimeSlices(time_slices: MediaTimeSlice[]): string {
    return (time_slices || [])
        .map((t) => {
            return renderTimeSeconds(t.time) + " " + t.name;
        })
        .join("\n");
}

export function parseTimeSlices(text: string): MediaTimeSlice[] {
    const res: MediaTimeSlice[] = [];

    const lines = text.split("\n");

    for (const line of lines) {
        const trimLine = line.trim();
        if (!trimLine) {
            continue;
        }

        const spaceIndex = trimLine.indexOf(" ");

        let timeStr = "";
        let sliceName = "";

        if (spaceIndex >= 0) {
            timeStr = trimLine.substring(0, spaceIndex);
            sliceName = trimLine
                .substring(spaceIndex + 1)
                .substring(0, 80)
                .trim();
            if (sliceName.startsWith("-")) {
                sliceName = sliceName.substring(1).trim();
            }
        } else {
            timeStr = trimLine;
        }

        const timeSeconds = parseTimeSeconds(timeStr);

        if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
            continue;
        }

        res.push({
            time: timeSeconds,
            name: sliceName,
        });
    }

    return res.slice(0, 1024);
}

/**
 * Normalized time slice
 */
export interface NormalizedTimeSlice {
    // Start time (seconds)
    start: number;

    // End time (seconds)
    end: number;

    // Slice name
    name: string;
}

/**
 * Normalizes time slices (resolving the full time range for each slice)
 * @param timeSlices The list of time slices
 * @param duration The media duration
 * @returns The list of normalized time slices
 */
export function normalizeTimeSlices(timeSlices: MediaTimeSlice[], duration: number): NormalizedTimeSlice[] {
    const res: NormalizedTimeSlice[] = [];

    for (let i = 0; i < timeSlices.length; i++) {
        const name = timeSlices[i].name;
        const start = timeSlices[i].time;
        let end = duration;

        if (i < timeSlices.length - 1) {
            end = timeSlices[i + 1].time;
        }

        res.push({
            name: name,
            start: start,
            end: end,
        });
    }

    return res;
}

/**
 * Finds a time slice
 * @param timeSlices The list of time slices
 * @param time The current time
 * @returns The found time slice, or null
 */
export function findTimeSlice(timeSlices: NormalizedTimeSlice[], time: number): NormalizedTimeSlice | null {
    if (timeSlices.length === 0) {
        return null;
    }

    let low = 0;
    let high = timeSlices.length - 1;

    while (low <= high) {
        const m = (low + high) >> 1;
        const v = timeSlices[m].start;

        if (v < time) {
            low = m + 1;
        } else if (v > time) {
            high = m - 1;
        } else {
            low = m;
            high = m - 1;
        }
    }

    if (timeSlices[low] && time >= timeSlices[low].start && time <= timeSlices[low].end) {
        return timeSlices[low];
    } else if (timeSlices[low - 1] && low > 0 && time >= timeSlices[low - 1].start && time <= timeSlices[low - 1].end) {
        return timeSlices[low - 1];
    } else {
        return null;
    }
}
