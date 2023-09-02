// Config API

import { GetApiURL, RequestParams } from "@/utils/request";
import { VaultUserConfig } from "./models";

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
