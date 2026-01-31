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

<script lang="ts">
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
import { defineComponent } from "vue";

export default defineComponent({
    name: "BottomBar",
    data: function () {
        return {
            focus: AppStatus.CurrentFocus,

            prev: AlbumsController.CurrentPrev,
            next: AlbumsController.CurrentNext,

            hasPagePrev: PagesController.HasPagePrev,
            hasPageNext: PagesController.HasPageNext,
        };
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onStatusUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, this.onAlbumPosUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, this.onPagePosUpdate.bind(this));
    },
    methods: {
        onStatusUpdate: function () {
            this.focus = AppStatus.CurrentFocus;
        },

        onAlbumPosUpdate: function () {
            this.prev = AlbumsController.CurrentPrev;
            this.next = AlbumsController.CurrentNext;
        },

        onPagePosUpdate: function (prev: boolean, next: boolean) {
            this.hasPagePrev = prev;
            this.hasPageNext = next;
        },

        clickLeft: function () {
            AppStatus.FocusLeft();
        },

        clickRight: function () {
            AppStatus.FocusRight();
        },

        goNext: function () {
            emitAppEvent(EVENT_NAME_GO_NEXT);
        },

        goPrev: function () {
            emitAppEvent(EVENT_NAME_GO_PREV);
        },
    },
});
</script>
