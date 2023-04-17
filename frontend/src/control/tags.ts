// Tags data controller

import { TagsAPI } from "@/api/api-tags";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AuthController } from "./auth";

export interface TagEntry {
    id: number;
    name: string;
}

export class TagsController {
    public static Tags: { [id: string]: TagEntry } = Object.create(null);

    public static Loading = true;

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", TagsController.Load);
        TagsController.Load();
    }

    public static Load() {
        TagsController.Loading = true;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        AppEvents.Emit("tags-loading", true);
        Timeouts.Abort("tags-load");
        Request.Pending("tags-load", TagsAPI.GetTags()).onSuccess(tags => {
            TagsController.Tags = Object.create(null);

            for (const tag of tags) {
                TagsController.Tags[tag.id + ""] = tag;
            }

            AppEvents.Emit("tags-update", TagsController.Tags);
            TagsController.Loading = false;
            AppEvents.Emit("tags-loading", false);
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("tags-load", 1500, TagsController.Load);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("tags-load", 1500, TagsController.Load);
        });
    }

    public static FindTag(name: string): number {
        const tags = Object.values(TagsController.Tags);

        for (const tag of tags) {
            if (tag.name === name) {
                return tag.id;
            }
        }

        return -1;
    }

    public static AddTag(id: number, name: string) {
        // Remove any other tag with that name
        const tags = Object.values(TagsController.Tags);

        for (const tag of tags) {
            if (tag.name === name) {
                delete TagsController.Tags[tag.id + ""];
                break;
            }
        }

        TagsController.Tags[id + ""] = {
            id: id,
            name: name,
        };

        AppEvents.Emit("tags-update", TagsController.Tags);
    }
}
