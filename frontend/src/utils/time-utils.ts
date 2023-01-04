

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
    if (s > 60) {
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

    return (h * 3600) + (m * 60) + s;
}

export function renderTimeSlices(time_slices: { time: number, name: string, }[]): string {
    return (time_slices || []).map(t => {
        return renderTimeSeconds(t.time) + " " + t.name;
    }).join("\n");
}

export function parseTimeSlices(text: string): { time: number, name: string, }[] {
    const res: { time: number, name: string, }[] = [];

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
            sliceName = trimLine.substring(spaceIndex + 1).substring(0, 80).trim();
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
        })
    }

    return res.slice(0, 1024);
}
