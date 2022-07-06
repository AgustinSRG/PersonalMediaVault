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
                const text = lineBuffer.slice(2).join("\n").trim();

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
