// Config API

import { GetAPIURL, RequestParams } from "@/utils/request";

export interface VaultUserConfig {
    max_tasks: number,
    encoding_threads: number,
    resolutions: {
        width: number,
        height: number,
        fps: number,
    }[],
    image_resolutions: {
        width: number,
        height: number,
    }[],
}

export class ConfigAPI {
    public static GetConfig(): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/config"),
        };
    }

    public static SetConfig(config: VaultUserConfig): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/config"),
            json: config,
        };
    }
}
