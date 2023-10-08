// Tags data controller

import { TagsAPI } from "@/api/api-tags";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AuthController } from "./auth";
import { MediaListItem } from "@/api/models";
import { MediaEntry } from "./media";

export class TagsController {
    public static TagsVersion = 0;
    public static Tags: Map<number, string> = new Map();
    public static LastTagId = -1;

    public static Loading = true;
    public static InitiallyLoaded = false;

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", TagsController.Load);
        TagsController.Load();
    }

    public static GetTagName(id: number, v: number) {
        if (import.meta.env.DEV) {
            if (TagsController.TagsVersion !== v) {
                console.warn("Tag version not properly updated. Current: " + TagsController.TagsVersion + " | Given: " + v);
            }
        }
        return TagsController.Tags.get(id) || "???";
    }

    public static Load() {
        TagsController.Loading = true;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        AppEvents.Emit("tags-loading", true);
        Timeouts.Abort("tags-load");
        Request.Pending("tags-load", TagsAPI.GetTags())
            .onSuccess((tags) => {
                TagsController.Tags = new Map();

                for (const tag of tags) {
                    if (tag.id > TagsController.LastTagId) {
                        TagsController.LastTagId = tag.id;
                    }
                    TagsController.Tags.set(tag.id, tag.name);
                }

                TagsController.TagsVersion++;
                AppEvents.Emit("tags-update", TagsController.TagsVersion);
                TagsController.Loading = false;
                AppEvents.Emit("tags-loading", false);
                TagsController.InitiallyLoaded = true;
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit("unauthorized", false);
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("tags-load", 1500, TagsController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set("tags-load", 1500, TagsController.Load);
            });
    }

    public static FindTag(name: string): number {
        for (const [tagId, tagName] of TagsController.Tags) {
            if (tagName === name) {
                return tagId;
            }
        }

        return -1;
    }

    public static AddTag(id: number, name: string) {
        // Remove any other tag with that name
        for (const [tagId, tagName] of TagsController.Tags) {
            if (tagName === name) {
                TagsController.Tags.delete(tagId);
                break;
            }
        }

        TagsController.Tags.set(id, name);

        TagsController.TagsVersion++;
        AppEvents.Emit("tags-update", TagsController.TagsVersion);

        if (TagsController.LastTagId < id) {
            if (TagsController.LastTagId === id - 1) {
                // Next tag added, no de-sync
                TagsController.LastTagId = id;
            } else if (!TagsController.Loading) {
                // De-sync, reload
                TagsController.Load();
            }
        }
    }

    public static RemoveTag(id: number) {
        TagsController.Tags.delete(id);

        TagsController.TagsVersion++;
        AppEvents.Emit("tags-update", TagsController.TagsVersion);
    }

    public static OnMediaListReceived(list: (MediaListItem | MediaEntry)[]) {
        if (TagsController.Loading) {
            return; // Already loading
        }
        for (const m of list) {
            if (!m.tags) {
                continue;
            }

            for (const t of m.tags) {
                if (t > TagsController.LastTagId) {
                    TagsController.Load();
                    return;
                }
            }
        }
    }
}
