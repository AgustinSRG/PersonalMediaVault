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
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";

export default defineComponent({
    name: "BottomBar",
    data: function () {
        return {
            focus: AppStatus.CurrentFocus,

            prev: AlbumsController.CurrentPrev,
            next: AlbumsController.CurrentNext,

            hasPagePrev: AlbumsController.HasPagePrev,
            hasPageNext: AlbumsController.HasPageNext,
        };
    },
    methods: {
        onStatusUpdate: function () {
            this.focus = AppStatus.CurrentFocus;
        },

        onAlbumPosUpdate: function () {
            this.prev = AlbumsController.CurrentPrev;
            this.next = AlbumsController.CurrentNext;
        },

        onPagePosUpdate: function () {
            this.hasPagePrev = AlbumsController.HasPagePrev;
            this.hasPageNext = AlbumsController.HasPageNext;
        },

        clickLeft: function () {
            AppStatus.FocusLeft();
        },

        clickRight: function () {
            AppStatus.FocusRight();
        },

        goNext: function () {
            AppEvents.Emit("media-go-next");
        },

        goPrev: function () {
            AppEvents.Emit("media-go-prev");
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.updateStatusH = this.onStatusUpdate.bind(this);
        AppStatus.AddEventListener(this._handles.updateStatusH);

        this._handles.posUpdateH = this.onAlbumPosUpdate.bind(this);
        AppEvents.AddEventListener("album-pos-update", this._handles.posUpdateH);

        this._handles.onPagePosUpdateH = this.onPagePosUpdate.bind(this);
        AppEvents.AddEventListener("page-media-nav-update", this._handles.onPagePosUpdateH);
    },
    beforeUnmount: function () {
        AppStatus.RemoveEventListener(this._handles.updateStatusH);

        AppEvents.RemoveEventListener("album-pos-update", this._handles.posUpdateH);
        AppEvents.RemoveEventListener("page-media-nav-update", this._handles.onPagePosUpdateH);
    },
});
</script>
