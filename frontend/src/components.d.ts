// Global components

import ModalDialogContainer from "@/components/utils/ModalDialogContainer.vue";

declare module '@vue/runtime-core' {
    export interface GlobalComponents {
        ModalDialogContainer: typeof ModalDialogContainer
    }
}
