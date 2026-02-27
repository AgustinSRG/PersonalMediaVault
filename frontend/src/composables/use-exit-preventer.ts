// Exit preventer composable

"use strict";

import { ExitPreventer } from "@/control/exit-prevent";
import type { Ref } from "vue";
import { onBeforeUnmount, onMounted, ref, shallowRef } from "vue";

/**
 * Exit preventer composable
 */
export type ExitPreventerComposable = {
    /**
     * True to display the exit confirmation
     */
    displayExitConfirmation: Ref<boolean>;

    /**
     * Call after saving
     */
    onSave: () => void;

    /**
     * Called when the user chooses to save the changes and exit
     */
    onExitSaveChanges: () => void;

    /**
     * Called when the user decides to discard changes and exit
     */
    onExitDiscardChanges: () => void;
};

/**
 * Gets the exit preventer composable
 * @param dirty The dirty status (changes made)
 * @param saveChanges A function to save the changes
 * @returns The composable
 */
export function useExitPreventer(dirty: Ref<boolean>, saveChanges: () => void) {
    // Display exit confirmation
    const displayExitConfirmation = ref(false);

    // Exit callback
    const exitCallback = shallowRef<null | (() => void)>(null);

    // Run the exit callback after saving
    const exitOnSave = ref(false);

    onMounted(() => {
        ExitPreventer.SetupExitPrevent(
            () => dirty.value,
            (callback: () => void) => {
                exitCallback.value = callback;
                displayExitConfirmation.value = true;
            },
        );
    });

    onBeforeUnmount(() => {
        ExitPreventer.RemoveExitPrevent();
    });

    /**
     * Called after saving
     */
    const onSave = () => {
        if (exitOnSave.value) {
            exitOnSave.value = false;

            if (exitCallback.value) {
                exitCallback.value();
            }
        }
    };

    /**
     * Called when the user chooses to save the changes and exit
     */
    const onExitSaveChanges = () => {
        if (dirty.value) {
            exitOnSave.value = true;
            saveChanges();
        } else {
            if (exitCallback.value) {
                exitCallback.value();
            }
        }
    };

    /**
     * Called when the user decides to discard changes and exit
     */
    const onExitDiscardChanges = () => {
        if (exitCallback.value) {
            exitCallback.value();
        }
    };

    return {
        displayExitConfirmation,
        onSave,
        onExitSaveChanges,
        onExitDiscardChanges,
    };
}
