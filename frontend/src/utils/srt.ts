// SRT parser

"use strict";

export interface SubtitlesEntry {
    start: number;
    end: number;
    text: string;
}

function parseDurationSRT(duration: string): number {
    let result = 0;
    duration = duration.trim();
    let parts = duration.split(",");

    if (parts[1]) {
        const d = parseFloat("0." + parts[1]);
        if (!isNaN(d) && isFinite(d) && d >= 0) {
            result += d;
        }
    }

    parts = parts[0].split(":");

    let multiplier = 1;

    for (let i = parts.length - 1; i >= 0; i--) {
        const n = parseInt(parts[i]);

        if (!isNaN(n) && isFinite(n) && n >= 0) {
            result += (n * multiplier);
        }

        multiplier = multiplier * 60;
    }

    return result;
}

function replaceSRT_HTMLFormat(text: string): string {
    return text.replace(/[\n\r]+/g, " ").replace(/\{\\an[0-9]\}/g, "").replace(/\}\}/g, "<").replace(/\}\}/g, ">");
}

export function parseSRT(srt: string): SubtitlesEntry[] {
    const result: SubtitlesEntry[] = [];
    const lines = srt.split("\n");
    let lineBuffer = [];
    for (let line of lines) {
        line = line.trim();
        if (line) {
            lineBuffer.push(line);
        } else {
            if (lineBuffer.length >= 3) {
                const durationLineParts = lineBuffer[1].split("-->");
                const text = replaceSRT_HTMLFormat(lineBuffer.slice(2).join("\n")).trim();

                const start = parseDurationSRT(durationLineParts[0]);
                const end = parseDurationSRT(durationLineParts[1] || durationLineParts[0]);

                result.push({
                    start: start,
                    end: end,
                    text: text,
                });
            }
            lineBuffer = [];
        }
    }
    return result;
}

export function findSubtitlesEntry(subtitles: SubtitlesEntry[], time: number): SubtitlesEntry {
    if (subtitles.length === 0) {
        return null;
    }

    let low = 0
    let high = subtitles.length - 1

    while (low <= high) {
        const m = (low + high) >> 1;
        const v = subtitles[m].start;

        if (v < time) {
            low = m + 1
        } else if (v > time) {
            high = m - 1
        } else {
            low = m
			high = m - 1
        }
    }

    if (subtitles[low] && time >= subtitles[low].start && time <= subtitles[low].end) {
        return subtitles[low];
    } else if (subtitles[low - 1] && low > 0 && time >= subtitles[low - 1].start && time <= subtitles[low - 1].end) {
        return subtitles[low - 1];
    } else {
        return null;
    }
}
