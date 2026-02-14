// Focus trap composable

"use strict";

import { FocusTrap } from "@/utils/focus-trap";
import type { Ref, ShallowRef } from "vue";
import { nextTick, onBeforeUnmount, onMounted, watch } from "vue";

/**
 * Sets up a focus trap for the component
 * @param container The container element that should keep the focus. Note: This element should be always available (no v-if)
 * @param displayRef The display status. When true, the trap will activate. When false, it will deactivate.
 * @param onFocusLost A function to call when the focus is lost (you can re-focus o close the element)
 * @param exceptClass The class for elements that will ignore the focus trap (example: buttons that show/hide a dropdown)
 * @param autoFocus True to automatically focus the container element when the traps activates
 */
export function useFocusTrap(
    container: Readonly<ShallowRef<HTMLElement>>,
    displayRef: Ref<boolean>,
    onFocusLost: () => void,
    exceptClass?: string,
    autoFocus?: boolean,
) {
    // Focus trap
    let focusTrap: null | FocusTrap = null;

    onMounted(() => {
        if (!container.value) {
            console.error("Warning: Missing container element for FocusTrap");
            return;
        }

        focusTrap = new FocusTrap(container.value, onFocusLost, exceptClass || undefined);

        if (displayRef.value) {
            focusTrap.activate();

            if (autoFocus) {
                nextTick(() => {
                    container.value?.focus();
                });
            }
        }
    });

    watch(displayRef, () => {
        if (displayRef.value) {
            focusTrap?.activate();

            if (autoFocus) {
                nextTick(() => {
                    container.value?.focus();
                });
            }
        } else {
            focusTrap?.deactivate();
        }
    });

    onBeforeUnmount(() => {
        focusTrap?.destroy();
    });
}
