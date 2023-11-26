<template>
    <div class="player-editor-sub-content">
        <!--- Time slices -->

        <div class="form-group">
            <label>{{ $t("Time slices") }}. {{ $t("You can split the time in slices and name them to display at the timeline.") }}</label>
        </div>

        <div class="form-group mt-1">
            <textarea
                v-model="timeSlices"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea auto-focus"
                :placeholder="'00:00:00 ' + $t('Opening') + '\n00:01:00 ' + $t('Rest of the video')"
                rows="10"
                :disabled="busy"
                @input="markDirty"
            ></textarea>
        </div>

        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-primary" :disabled="busy || !dirty" @click="changeTimeSlices">
                <i class="fas fa-pencil-alt"></i> {{ $t("Change time slices") }}
            </button>
        </div>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { Request } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { parseTimeSlices, renderTimeSlices } from "@/utils/time-slices";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaChangeTimeSlices } from "@/api/api-media-edit";

export default defineComponent({
    components: {},
    name: "EditorTimeSlices",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            timeSlices: "",
            originalTimeSlices: "",

            busy: false,

            dirty: false,

            canWrite: AuthController.CanWrite,
        };
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

        markDirty: function () {
            this.dirty = true;
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.originalTimeSlices = renderTimeSlices(MediaController.MediaData.time_slices);
            this.timeSlices = this.originalTimeSlices;
            this.dirty = false;
        },

        changeTimeSlices: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            const slices = parseTimeSlices(this.timeSlices);

            Request.Pending(this.requestId, apiMediaChangeTimeSlices(mediaId, slices))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Successfully changed time slices"));
                    this.busy = false;
                    this.dirty = false;
                    this.originalTimeSlices = renderTimeSlices(slices);
                    this.timeSlices = this.originalTimeSlices;

                    if (MediaController.MediaData) {
                        MediaController.MediaData.time_slices = clone(slices);
                    }

                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBar(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.autoFocus();
    },

    beforeUnmount: function () {
        Request.Abort(this.requestId);
    },
});
</script>
