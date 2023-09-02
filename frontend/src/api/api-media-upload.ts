// Upload media API

import { GetApiURL, RequestParams } from "@/utils/request";

export class UploadMediaAPI {
    public static UploadMedia(title: string, file: File, album: number): RequestParams<{ media_id: number }> {
        const form = new FormData();
        form.append("file", file);

        return {
            method: "POST",
            url: GetApiURL("/api/upload?title=" + encodeURIComponent(title) + "&album=" + encodeURIComponent(album + "")),
            form: form,
        };
    }
}
