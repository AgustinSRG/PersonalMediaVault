// Utils related to files

"use strict";

/**
 * Removes extension if present, keeping only the file name
 * @param fileName The file name with extension
 * @returns The file name without extension
 */
export function removeExtensionFromFileName(fileName: string): string {
    const parts = (fileName + "").split(".");
    if (parts.length > 1) {
        parts.pop();
        return parts.join(".");
    } else {
        return fileName;
    }
}
