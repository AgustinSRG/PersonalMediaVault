// Composable for user permissions

"use strict";

import { AuthController } from "@/control/auth";
import type { Ref } from "vue";
import { ref } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_AUTH_CHANGED } from "@/control/app-events";

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
    const isGuest = ref(!AuthController.Username);
    const canWrite = ref(AuthController.CanWrite);
    const isRoot = ref(AuthController.IsRoot);

    onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
        isGuest.value = !AuthController.Username;
        canWrite.value = AuthController.CanWrite;
        isRoot.value = AuthController.IsRoot;
    });

    return {
        isGuest,
        canWrite,
        isRoot,
    };
}
