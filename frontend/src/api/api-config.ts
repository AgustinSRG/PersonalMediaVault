// Config API

"use strict";

import { RequestParams } from "./request";
import { VaultUserConfig } from "./models";
import { getApiURL } from "./utils";

export class ConfigAPI {
    public static GetConfig(): RequestParams<VaultUserConfig> {
        return {
            method: "GET",
            url: getApiURL("/api/config"),
        };
    }

    public static SetConfig(config: VaultUserConfig): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/config"),
            json: config,
        };
    }
}
