// Time slices utils

"use strict";

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

export function renderTimeSlices(time_slices: { time: number; name: string }[]): string {
    return (time_slices || [])
        .map((t) => {
            return renderTimeSeconds(t.time) + " " + t.name;
        })
        .join("\n");
}

export function parseTimeSlices(text: string): { time: number; name: string }[] {
    const res: { time: number; name: string }[] = [];

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

export function normalizeTimeSlices(
    time_slices: { time: number; name: string }[],
    duration: number,
): { start: number; end: number; name: string }[] {
    const res: { start: number; end: number; name: string }[] = [];

    for (let i = 0; i < time_slices.length; i++) {
        const name = time_slices[i].name;
        const start = time_slices[i].time;
        let end = duration;

        if (i < time_slices.length - 1) {
            end = time_slices[i + 1].time;
        }

        res.push({
            name: name,
            start: start,
            end: end,
        });
    }

    return res;
}

export function findTimeSlice(
    time_slices: { start: number; end: number; name: string }[],
    time: number,
): { start: number; end: number; name: string } {
    if (time_slices.length === 0) {
        return null;
    }

    let low = 0;
    let high = time_slices.length - 1;

    while (low <= high) {
        const m = (low + high) >> 1;
        const v = time_slices[m].start;

        if (v < time) {
            low = m + 1;
        } else if (v > time) {
            high = m - 1;
        } else {
            low = m;
            high = m - 1;
        }
    }

    if (time_slices[low] && time >= time_slices[low].start && time <= time_slices[low].end) {
        return time_slices[low];
    } else if (time_slices[low - 1] && low > 0 && time >= time_slices[low - 1].start && time <= time_slices[low - 1].end) {
        return time_slices[low - 1];
    } else {
        return null;
    }
}
