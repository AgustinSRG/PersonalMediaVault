// Clipboard for tags editor

"use strict";

import { fetchFromLocalStorage, saveIntoLocalStorage } from "@/local-storage/local-storage";
import { parseTagName } from "@/utils/tags";
import { getCurrentMediaData } from "./media";
import { getTagsVersion, resolveTagName } from "./tags";
import { emitAppEvent, EVENT_NAME_TAGS_EDITOR_CLIPBOARD_CHANGED } from "./app-events";

// Local storage key
const LS_KEY_TAGS_CLIPBOARD = "tags-editor-clipboard";

/**
 * Clipboard for tags editor
 */
const TagsEditorClipboard = {
    /**
     * List of tags in the clipboard
     */
    tags: fetchFromLocalStorage(LS_KEY_TAGS_CLIPBOARD, "")
        .split(",")
        .map((t) => parseTagName(t))
        .filter((t) => !!t),
};

// Channel to update the clipboard between tabs
const TagsUpdateChannel = new BroadcastChannel("tags-editor-clipboard-channel");

TagsUpdateChannel.addEventListener("message", (event) => {
    const data = event.data as unknown;

    if (
        typeof data === "object" &&
        data !== null &&
        "tags" in data &&
        Array.isArray(data.tags) &&
        !data.tags.map((t) => typeof t === "string").includes(false)
    ) {
        TagsEditorClipboard.tags = data.tags.slice();
        emitAppEvent(EVENT_NAME_TAGS_EDITOR_CLIPBOARD_CHANGED, data.tags);
    }
});

/**
 * Gets the tags editor clipboard content
 * @returns The list of tags in the clipboard
 */
export function getTagsEditorClipboardContent(): string[] {
    return TagsEditorClipboard.tags.slice();
}

/**
 * Sets the content of the tags editor clipboard
 * @param tags The list of tags
 */
export function setTagsEditorClipboardContent(tags: string[]) {
    TagsEditorClipboard.tags = tags.slice();
    saveIntoLocalStorage(LS_KEY_TAGS_CLIPBOARD, TagsEditorClipboard.tags.join(","));
    TagsUpdateChannel.postMessage({ tags: TagsEditorClipboard.tags });
    emitAppEvent(EVENT_NAME_TAGS_EDITOR_CLIPBOARD_CHANGED, tags);
}

/**
 * Copies the current media tags into tags editor clipboard
 */
export function copyCurrentMediaTagsIntoTagsEditorClipboard() {
    setTagsEditorClipboardContent((getCurrentMediaData()?.tags || []).map((t) => resolveTagName(t, getTagsVersion())));
}
