// Current media composable

"use strict";

import type { MediaData } from "@/api/models";
import { MediaController } from "@/control/media";
import type { Ref } from "vue";
import { ref } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";

/**
 * Current media composable
 */
export type CurrentMediaComposable = {
    /**
     * Current media ID
     */
    currentMediaId: Ref<number>;

    /**
     * Current media data
     */
    currentMediaData: Ref<MediaData | null>;
};

/**
 * Gets the current media composable
 * @returns The current media composable
 */
export function useCurrentMedia(): CurrentMediaComposable {
    const currentMediaId = ref(MediaController.MediaId);
    const currentMediaData = ref(MediaController.MediaData);

    onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, () => {
        currentMediaId.value = MediaController.MediaId;
        currentMediaData.value = MediaController.MediaData;
    });

    return {
        currentMediaId,
        currentMediaData,
    };
}
