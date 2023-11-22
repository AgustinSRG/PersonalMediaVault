// Upload media API

"use strict";

import { RequestParams } from "@asanrom/request-browser";
import { getApiURL } from "@/utils/api";

export class UploadMediaAPI {
    public static UploadMedia(title: string, file: File, album: number): RequestParams<{ media_id: number }> {
        const form = new FormData();
        form.append("file", file);

        return {
            method: "POST",
            url: getApiURL("/api/upload?title=" + encodeURIComponent(title) + "&album=" + encodeURIComponent(album + "")),
            form: form,
        };
    }
}
