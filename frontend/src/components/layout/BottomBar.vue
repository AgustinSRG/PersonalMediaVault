<template>
    <div class="bottom-bar">
        <div
            class="bottom-bar-option bottom-bar-option-list"
            :class="{ disabled: !prev && !hasPagePrev }"
            tabindex="0"
            @click="goPrev"
            @keydown="clickOnEnter"
        >
            <i class="fas fa-backward-step"></i><span> {{ $t("Previous") }}</span>
        </div>
        <div
            class="bottom-bar-option bottom-bar-option-media"
            :class="{ selected: focus === 'left' }"
            tabindex="0"
            @click="clickLeft"
            @keydown="clickOnEnter"
        >
            <i class="fas fa-photo-film"></i><span> {{ $t("Media") }}</span>
        </div>
        <div
            class="bottom-bar-option bottom-bar-option-list"
            :class="{ selected: focus === 'right' }"
            tabindex="0"
            @click="clickRight"
            @keydown="clickOnEnter"
        >
            <i class="fas fa-list"></i><span> {{ $t("List") }}</span>
        </div>
        <div
            class="bottom-bar-option bottom-bar-option-list"
            :class="{ disabled: !next && !hasPageNext }"
            tabindex="0"
            @click="goNext"
            @keydown="clickOnEnter"
        >
            <i class="fas fa-forward-step"></i><span> {{ $t("Next") }}</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";
import { AlbumsController } from "@/control/albums";
import {
    emitAppEvent,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_GO_NEXT,
    EVENT_NAME_GO_PREV,
    EVENT_NAME_PAGE_MEDIA_NAV_UPDATE,
} from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { PagesController } from "@/control/pages";
import { clickOnEnter } from "@/utils/events";
import { ref } from "vue";

// Translation function
const { $t } = useI18n();

// Current focus
const focus = ref(AppStatus.CurrentFocus);

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    focus.value = AppStatus.CurrentFocus;
});

// Previous and next element in album
const prev = ref(AlbumsController.CurrentPrev);
const next = ref(AlbumsController.CurrentNext);

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, () => {
    prev.value = AlbumsController.CurrentPrev;
    next.value = AlbumsController.CurrentNext;
});

// Previous and next element in page
const hasPagePrev = ref(PagesController.HasPagePrev);
const hasPageNext = ref(PagesController.HasPageNext);

onApplicationEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, (p, n) => {
    hasPagePrev.value = p;
    hasPageNext.value = n;
});

/**
 * Called when the user clicked on the 'left' button
 */
const clickLeft = () => {
    AppStatus.FocusLeft();
};

/**
 * Called when the user clicked on the 'right' button
 */
const clickRight = () => {
    AppStatus.FocusRight();
};

/**
 * Called when the user clicked on the 'next' button
 */
const goNext = () => {
    emitAppEvent(EVENT_NAME_GO_NEXT);
};

/**
 * Called when the user clicked on the 'previous' button
 */
const goPrev = () => {
    emitAppEvent(EVENT_NAME_GO_PREV);
};
</script>
