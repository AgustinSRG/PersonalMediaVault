// Time utils

"use strict";

/**
 * Renders time
 * @param s The time (seconds)
 * @returns The rendered time
 */
export function renderTimeSeconds(s: number): string {
    if (isNaN(s) || !isFinite(s) || s < 0) {
        s = 0;
    }
    s = Math.floor(s);
    let hours = 0;
    let minutes = 0;
    if (s >= 3600) {
        hours = Math.floor(s / 3600);
        s = s % 3600;
    }
    if (s >= 60) {
        minutes = Math.floor(s / 60);
        s = s % 60;
    }
    let r = "";

    if (s > 9) {
        r = "" + s + r;
    } else {
        r = "0" + s + r;
    }

    if (minutes > 9) {
        r = "" + minutes + ":" + r;
    } else {
        r = "0" + minutes + ":" + r;
    }

    if (hours > 0) {
        if (hours > 9) {
            r = "" + hours + ":" + r;
        } else {
            r = "0" + hours + ":" + r;
        }
    }

    return r;
}

/**
 * Renders date and time
 * @param date The date
 * @param locale The locale
 * @returns The rendered date and time
 */
export function renderDateAndTime(date: Date | string | number, locale: string) {
    if (!date) return "-";

    if (typeof date === "string" || typeof date === "number") {
        date = new Date(date);
    }

    return date.toLocaleTimeString(locale, {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}

/**
 * Renders date for the user
 * @param ts The timestamp or date
 * @param locale The locale
 * @returns The rendered date
 */
export function renderDate(date: Date | string | number, locale: string): string {
    return new Date(date).toLocaleString(locale, {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}
