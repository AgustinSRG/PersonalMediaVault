// Composable for user permissions

"use strict";

import type { Ref } from "vue";
import { ref } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_AUTH_CHANGED } from "@/global-state/app-events";
import { getAuthStatus } from "@/global-state/auth";

/**
 * User permissions composable
 */
export type UserPermissionsComposable = {
    /**
     * True if the user is a guest user
     */
    isGuest: Ref<boolean>;

    /**
     * True if the user has write permissions
     */
    canWrite: Ref<boolean>;

    /**
     * True if the user has administrative permissions
     */
    isRoot: Ref<boolean>;
};

/**
 * Gets a composable with references for user permissions variables
 * @returns A composable with references to the user permissions variables
 */
export function useUserPermissions(): UserPermissionsComposable {
    // Initial auth status
    const initialAuthStatus = getAuthStatus();

    // Is guest? (used invite code to login)
    const isGuest = ref(!initialAuthStatus.username);

    // Has write permissions?
    const canWrite = ref(initialAuthStatus.canWrite);

    // Has root permissions?
    const isRoot = ref(initialAuthStatus.isRoot);

    onApplicationEvent(EVENT_NAME_AUTH_CHANGED, (newAuthStatus) => {
        isGuest.value = !newAuthStatus.username;
        canWrite.value = newAuthStatus.canWrite;
        isRoot.value = newAuthStatus.isRoot;
    });

    return {
        isGuest,
        canWrite,
        isRoot,
    };
}
