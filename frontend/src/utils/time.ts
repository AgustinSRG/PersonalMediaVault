// Time utils

"use strict";

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

export function renderDateAndTime(date: Date | string | number, $t: (t: string) => string) {
    if (!date) return "-";
    if (typeof date === "string" || typeof date === "number") {
        date = new Date(date);
    }

    const template = $t("[m] [d], [y] [hh]:[mm]:[ss]");
    const months = [
        $t("January"),
        $t("February"),
        $t("March"),
        $t("April"),
        $t("May"),
        $t("June"),
        $t("July"),
        $t("August"),
        $t("September"),
        $t("October"),
        $t("November"),
        $t("December"),
    ];

    const y = date.getFullYear();
    const d = date.getDate();

    let m = date.getMonth() + 1;

    m = m - 1;
    if (m < 0) {
        m = 0;
    } else if (m > months.length) {
        m = months.length - 1;
    }

    let hh = "" + date.getHours();
    let mm = "" + date.getMinutes();
    let ss = "" + date.getSeconds();

    if (hh.length < 2) {
        hh = "0" + hh;
    }

    if (mm.length < 2) {
        mm = "0" + mm;
    }

    if (ss.length < 2) {
        ss = "0" + ss;
    }

    return ("" + template)
        .replace("[m]", months[m])
        .replace("[d]", d + "")
        .replace("[y]", y + "")
        .replace("[hh]", hh)
        .replace("[mm]", mm)
        .replace("[ss]", ss);
}
