// Size utils

"use strict";

const KB = 1024;
const MB = KB * KB;
const GB = KB * MB;
const TB = KB * GB;
const PB = KB * TB;

interface SizeUnit {
    /**
     * Size in bytes
     */
    size: number;

    /**
     * Unit label
     */
    label: string;
}

const SIZE_UNITS: SizeUnit[] = [
    {
        size: PB,
        label: "PB",
    },
    {
        size: TB,
        label: "TB",
    },
    {
        size: GB,
        label: "GB",
    },
    {
        size: MB,
        label: "MB",
    },
    {
        size: KB,
        label: "KB",
    },
];

/**
 * Renders size in the highest unit possible
 * @param bytes The total number of bytes
 * @returns The rendered size
 */
export function renderSize(bytes: number): string {
    const unit = SIZE_UNITS.find((u) => bytes >= u.size);

    if (unit) {
        return Math.floor((bytes / unit.size) * 100) / 100 + " " + unit.label;
    } else {
        return bytes + " Bytes";
    }
}
