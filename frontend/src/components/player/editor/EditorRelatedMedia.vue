<template>
    <div class="player-editor-sub-content">
        <!--- Related media -->

        <div class="form-group">
            <label>{{
                $t("You can add here references to other media assets related to this one, in order to find them more easily.")
            }}</label>
        </div>
        <div class="editor-related-media-list">
            <div v-for="(item, i) in relatedMedia" :key="item.id" class="editor-related-media-item">
                <div class="album-body-item-thumbnail" :title="item.title || $t('Untitled')">
                    <div v-if="!item.thumbnail" class="no-thumb">
                        <i v-if="item.type === 1" class="fas fa-image"></i>
                        <i v-else-if="item.type === 2" class="fas fa-video"></i>
                        <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
                        <i v-else class="fas fa-ban"></i>
                    </div>
                    <ThumbImage v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)"></ThumbImage>
                    <DurationIndicator
                        v-if="item.type === 2 || item.type === 3"
                        :type="item.type"
                        :duration="item.duration"
                        :small="true"
                    ></DurationIndicator>
                </div>
                <div class="editor-related-media-right">
                    <div class="editor-related-media-title">
                        <a :href="getMediaURL(item.id)" target="_blank" rel="noopener noreferrer">{{ item.title || $t("Untitled") }}</a>
                    </div>
                    <div v-if="canWrite" class="editor-related-media-buttons">
                        <button
                            type="button"
                            :disabled="busyRelated || i === 0"
                            class="btn btn-xs btn-mr btn-primary"
                            @click="moveRelatedMediaUp(i)"
                        >
                            <i class="fas fa-arrow-up"></i> {{ $t("Move up") }}
                        </button>
                        <button
                            type="button"
                            :disabled="busyRelated || i >= relatedMedia.length - 1"
                            class="btn btn-xs btn-mr btn-primary"
                            @click="moveRelatedMediaDown(i)"
                        >
                            <i class="fas fa-arrow-down"></i> {{ $t("Move down") }}
                        </button>
                        <button type="button" :disabled="busyRelated" class="btn btn-xs btn-mr btn-danger" @click="removeRelatedMedia(i)">
                            <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
        <div v-if="canWrite && relatedMedia.length < maxRelatedMediaCount" class="form-group">
            <button type="button" :disabled="busyRelated" class="btn btn-xs btn-primary" @click="addRelatedMedia">
                <i class="fas fa-plus"></i> {{ $t("Add related media") }}
            </button>
        </div>
        <div v-if="canWrite" class="form-group">
            <button
                v-if="!compareMediaArrays(originalRelatedMedia, relatedMedia) || busyRelated || !savedRelated"
                type="button"
                class="btn btn-primary"
                :disabled="busyRelated || compareMediaArrays(originalRelatedMedia, relatedMedia)"
                @click="changeRelatedMedia"
            >
                <LoadingIcon icon="fas fa-pencil-alt" :loading="busyRelated"></LoadingIcon> {{ $t("Change related media") }}
            </button>
            <button v-else type="button" disabled class="btn btn-primary">
                <i class="fas fa-check"></i> {{ $t("Saved related media") }}
            </button>
            <div v-if="errorRelated" class="form-error form-error-pt">{{ errorRelated }}</div>
        </div>

        <AddRelatedMediaModal
            v-if="displayAddRelatedMediaModal"
            v-model:display="displayAddRelatedMediaModal"
            :mid="mid"
            :related-media="relatedMedia"
            @add-media="onAddRelatedMedia"
        ></AddRelatedMediaModal>

        <SaveChangesAskModal
            v-if="displayExitConfirmation"
            v-model:display="displayExitConfirmation"
            @yes="onExitSaveChanges"
            @no="onExitDiscardChanges"
        ></SaveChangesAskModal>
    </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { generateURIQuery, getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { EVENT_NAME_MEDIA_METADATA_CHANGE, PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiMediaChangeRelatedMedia } from "@/api/api-media-edit";
import ThumbImage from "@/components/utils/ThumbImage.vue";
import DurationIndicator from "@/components/utils/DurationIndicator.vue";
import SaveChangesAskModal from "@/components/modals/SaveChangesAskModal.vue";

import type { MediaListItem } from "@/api/models";
import { ExitPreventer } from "@/control/exit-prevent";

const AddRelatedMediaModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AddRelatedMediaModal.vue"),
});

export default defineComponent({
    name: "EditorRelatedMedia",
    components: {
        LoadingIcon,
        ThumbImage,
        DurationIndicator,
        AddRelatedMediaModal,
        SaveChangesAskModal,
    },
    emits: ["changed"],
    setup() {
        return {
            maxRelatedMediaCount: 16,

            requestIdRelated: getUniqueStringId(),

            exitCallback: null as () => void,
        };
    },
    data: function () {
        return {
            mid: AppStatus.CurrentMedia,

            originalRelatedMedia: [] as MediaListItem[],
            relatedMedia: [] as MediaListItem[],

            busyRelated: false,

            savedRelated: false,

            errorRelated: "",

            canWrite: AuthController.CanWrite,

            displayAddRelatedMediaModal: false,

            displayExitConfirmation: false,
            exitOnSave: false,
        };
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.autoFocus();

        ExitPreventer.SetupExitPrevent(this.checkExitPrevent.bind(this), this.onExit.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestIdRelated);

        ExitPreventer.RemoveExitPrevent();
    },

    methods: {
        autoFocus: function () {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.mid = MediaController.MediaData.id;

            this.originalRelatedMedia = (MediaController.MediaData.related || []).slice();
            this.relatedMedia = this.originalRelatedMedia.slice();
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
        },

        addRelatedMedia: function () {
            this.displayAddRelatedMediaModal = true;
        },

        moveRelatedMediaUp: function (i: number) {
            this.relatedMedia.splice(i - 1, 0, this.relatedMedia.splice(i, 1)[0]);
        },

        moveRelatedMediaDown: function (i: number) {
            this.relatedMedia.splice(i + 1, 0, this.relatedMedia.splice(i, 1)[0]);
        },

        removeRelatedMedia: function (i: number) {
            this.relatedMedia.splice(i, 1);
        },

        onAddRelatedMedia: function (m: MediaListItem, callback: () => void) {
            this.relatedMedia.push(m);

            if (this.relatedMedia.length >= this.maxRelatedMediaCount) {
                this.displayAddRelatedMediaModal = false;
            }

            callback();
        },

        changeRelatedMedia: function (e?: Event) {
            if (e) {
                e.preventDefault();
            }

            if (this.busyRelated) {
                return;
            }

            this.busyRelated = true;
            this.errorRelated = "";

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(
                this.requestIdRelated,
                apiMediaChangeRelatedMedia(
                    mediaId,
                    this.relatedMedia.map((m) => m.id),
                ),
            )
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed related media"));
                    this.busyRelated = false;
                    this.savedRelated = true;
                    this.originalRelatedMedia = this.relatedMedia.slice();
                    if (MediaController.MediaData) {
                        MediaController.MediaData.related = this.relatedMedia.slice();
                    }
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit(EVENT_NAME_MEDIA_METADATA_CHANGE);

                    if (this.exitOnSave) {
                        this.exitOnSave = false;
                        if (this.exitCallback) {
                            this.exitCallback();
                        }
                    }
                })
                .onCancel(() => {
                    this.busyRelated = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busyRelated = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.errorRelated = this.$t("Error") + ": " + this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            this.errorRelated = this.$t("Error") + ": " + this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.errorRelated = this.$t("Error") + ": " + this.$t("Access denied");
                        },
                        notFound: () => {
                            this.errorRelated = this.$t("Error") + ": " + this.$t("Not found");
                        },
                        serverError: () => {
                            this.errorRelated = this.$t("Error") + ": " + this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.errorRelated = this.$t("Error") + ": " + this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.errorRelated = this.$t("Error") + ": " + err.message;
                    console.error(err);
                    this.busyRelated = false;
                });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        compareMediaArrays: function (m1: MediaListItem[], m2: MediaListItem[]): boolean {
            return m1.map((m) => m.id).join(",") === m2.map((m) => m.id).join(",");
        },

        getMediaURL: function (mid: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    media: mid + "",
                })
            );
        },

        checkExitPrevent: function (): boolean {
            return !this.compareMediaArrays(this.relatedMedia, this.originalRelatedMedia);
        },

        onExit: function (callback: () => void) {
            this.exitCallback = callback;
            this.displayExitConfirmation = true;
        },

        onExitSaveChanges: function () {
            if (!this.compareMediaArrays(this.relatedMedia, this.originalRelatedMedia)) {
                this.exitOnSave = true;
                this.changeRelatedMedia();
            } else {
                if (this.exitCallback) {
                    this.exitCallback();
                }
            }
        },

        onExitDiscardChanges: function () {
            if (this.exitCallback) {
                this.exitCallback();
            }
        },
    },
});
</script>
