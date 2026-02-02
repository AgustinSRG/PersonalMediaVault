// Composable for modals

"use strict";

import type { ModelRef, ShallowRef } from "vue";
import { nextTick, onMounted, watch } from "vue";

// Interface for ModalDialogContainer
interface IModalDialogContainer {
    close(forced?: boolean): void;
    focus(): void;
}

/**
 * Composable for modals
 */
export type ModalComposable = {
    /**
     * Closes the modal
     */
    close: () => void;

    /**
     * Forcefully closes the modal
     */
    forceClose: () => void;

    /**
     * Focuses the element with 'auto-focus' class
     * or the modal container
     */
    focus: () => void;
};

/**
 * Creates a composable for the modal
 * @param display The display model ref
 * @param container The container ref
 * @returns Any methods or references useful for the modal
 */
export function useModal(display: ModelRef<boolean>, container: Readonly<ShallowRef<IModalDialogContainer, IModalDialogContainer>>) {
    const close = () => {
        container.value?.close();
    };

    const forceClose = () => {
        container.value?.close(true);
    };

    const focus = () => {
        nextTick(() => {
            container.value?.focus();
        });
    };

    onMounted(() => {
        if (display.value) {
            focus();
        }
    });

    watch(display, () => {
        if (display.value) {
            focus();
        }
    });

    return { close, forceClose, focus };
}
