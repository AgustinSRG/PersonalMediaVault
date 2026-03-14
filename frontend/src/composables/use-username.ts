// Username composable

"use strict";

import { EVENT_NAME_AUTH_CHANGED } from "@/global-state/app-events";
import { getAuthStatus } from "@/global-state/auth";
import type { Ref } from "vue";
import { ref } from "vue";
import { onApplicationEvent } from "./on-app-event";

/**
 * Gets a reference to the username.
 * This reference is automatically updated when the auth status changes.
 * @returns The reference to the username
 */
export function useUsername(): Ref<string> {
    // Username
    const username = ref(getAuthStatus().username);

    onApplicationEvent(EVENT_NAME_AUTH_CHANGED, (newAuthStatus) => {
        username.value = newAuthStatus.username;
    });

    return username;
}
