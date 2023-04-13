// Config API

import { GetApiURL, RequestParams } from "@/utils/request";

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
    public static GetConfig(): RequestParams<VaultUserConfig> {
        return {
            method: "GET",
            url: GetApiURL("/api/config"),
        };
    }

    public static SetConfig(config: VaultUserConfig): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/config"),
            json: config,
        };
    }
}
