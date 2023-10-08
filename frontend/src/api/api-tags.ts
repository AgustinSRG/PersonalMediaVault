// Tags API

import { GetApiURL, RequestParams } from "@/utils/request";
import { MediaTag } from "./models";

export class TagsAPI {
    public static GetTags(): RequestParams<MediaTag[]> {
        return {
            method: "GET",
            url: GetApiURL("/api/tags"),
        };
    }

    public static TagMedia(media: number, tagName: string): RequestParams<MediaTag> {
        return {
            method: "POST",
            url: GetApiURL("/api/tags/add"),
            json: {
                media_id: media,
                tag_name: tagName,
            },
        };
    }

    public static UntagMedia(media: number, tagId: number): RequestParams<{ removed: boolean }> {
        return {
            method: "POST",
            url: GetApiURL("/api/tags/remove"),
            json: {
                media_id: media,
                tag_id: tagId,
            },
        };
    }
}
