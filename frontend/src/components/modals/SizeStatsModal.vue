<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <div v-if="display" class="modal-dialog modal-sm" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Size Statistics") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body no-padding table-responsive">
                <table class="table table-text-overflow no-margin">
                    <tbody>
                        <tr>
                            <td>{{ $t("Metadata") }}</td>
                            <td class="text-right">{{ renderSize(metaSize) }}</td>
                        </tr>

                        <tr v-if="originalSize > 0">
                            <td>{{ $t("Original") }}</td>
                            <td class="text-right">{{ renderSize(originalSize) }}</td>
                        </tr>

                        <tr v-if="resizedSize > 0">
                            <td>{{ $t("Resized media") }}</td>
                            <td class="text-right">{{ renderSize(resizedSize) }}</td>
                        </tr>

                        <tr v-if="thumbnailSize > 0">
                            <td>{{ $t("Thumbnail") }}</td>
                            <td class="text-right">{{ renderSize(thumbnailSize) }}</td>
                        </tr>

                        <tr v-if="videoPreviewsSize > 0">
                            <td>{{ $t("Video previews") }}</td>
                            <td class="text-right">{{ renderSize(videoPreviewsSize) }}</td>
                        </tr>

                        <tr v-if="subtitlesSize > 0">
                            <td>{{ $t("Subtitles") }}</td>
                            <td class="text-right">{{ renderSize(subtitlesSize) }}</td>
                        </tr>

                        <tr v-if="audioTracksSize > 0">
                            <td>{{ $t("Audio tracks") }}</td>
                            <td class="text-right">{{ renderSize(audioTracksSize) }}</td>
                        </tr>

                        <tr v-if="imageNotesSize > 0">
                            <td>{{ $t("Image notes") }}</td>
                            <td class="text-right">{{ renderSize(imageNotesSize) }}</td>
                        </tr>

                        <tr v-if="descriptionSize > 0">
                            <td>{{ $t("Description") }}</td>
                            <td class="text-right">{{ renderSize(descriptionSize) }}</td>
                        </tr>

                        <tr v-if="attachmentsSize > 0">
                            <td>{{ $t("Attachments") }}</td>
                            <td class="text-right">{{ renderSize(attachmentsSize) }}</td>
                        </tr>

                        <tr v-if="otherSize > 0">
                            <td>{{ $t("Other assets") }}</td>
                            <td class="text-right">{{ renderSize(otherSize) }}</td>
                        </tr>

                        <tr>
                            <td class="bold">{{ $t("Total") }}</td>
                            <td class="bold text-right">{{ renderSize(total) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { AppEvents } from "@/control/app-events";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiMediaGetMediaSizeStats } from "@/api/api-media";

export default defineComponent({
    name: "SizeStatsModal",
    props: {
        display: Boolean,
        mid: Number,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            loading: false,
            metaSize: 0,

            originalSize: 0,

            thumbnailSize: 0,

            resizedSize: 0,

            videoPreviewsSize: 0,

            subtitlesSize: 0,
            audioTracksSize: 0,
            imageNotesSize: 0,
            descriptionSize: 0,

            attachmentsSize: 0,

            otherSize: 0,

            total: 0,

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
                this.load();
            }
        },

        mid: function () {
            this.load();
        },
    },
    mounted: function () {
        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
            this.load();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    methods: {
        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            this.loading = true;

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiMediaGetMediaSizeStats(this.mid))
                .onSuccess((result) => {
                    this.loading = false;
                    this.metaSize = result.meta_size;

                    this.originalSize = 0;

                    this.resizedSize = 0;

                    this.thumbnailSize = 0;

                    this.subtitlesSize = 0;
                    this.audioTracksSize = 0;
                    this.imageNotesSize = 0;

                    this.descriptionSize = 0;

                    this.attachmentsSize = 0;

                    this.videoPreviewsSize = 0;

                    this.otherSize = 0;

                    this.total = 0;

                    this.total += result.meta_size;

                    for (const asset of result.assets) {
                        const assetName = asset.name + "";

                        if (assetName === "ORIGINAL") {
                            this.originalSize += asset.size;
                        } else if (assetName === "THUMBNAIL") {
                            this.thumbnailSize += asset.size;
                        } else if (assetName === "VIDEO_PREVIEWS") {
                            this.videoPreviewsSize += asset.size;
                        } else if (assetName === "IMG_NOTES") {
                            this.imageNotesSize += asset.size;
                        } else if (assetName === "DESCRIPTION") {
                            this.descriptionSize += asset.size;
                        } else if (assetName.startsWith("RESIZED_")) {
                            this.resizedSize += asset.size;
                        } else if (assetName.startsWith("SUBTITLES_")) {
                            this.subtitlesSize += asset.size;
                        } else if (assetName.startsWith("AUDIO_TRACK_")) {
                            this.audioTracksSize += asset.size;
                        } else if (assetName.startsWith("ATTACHMENT:")) {
                            this.attachmentsSize += asset.size;
                        } else {
                            this.otherSize += asset.size;
                        }

                        this.total += asset.size;
                    }
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        notFound: () => {
                            this.close();
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        renderSize: function (bytes: number): string {
            if (bytes > 1024 * 1024 * 1024) {
                let gb = bytes / (1024 * 1024 * 1024);
                gb = Math.floor(gb * 100) / 100;
                return gb + " GB";
            } else if (bytes > 1024 * 1024) {
                let mb = bytes / (1024 * 1024);
                mb = Math.floor(mb * 100) / 100;
                return mb + " MB";
            } else if (bytes > 1024) {
                let kb = bytes / 1024;
                kb = Math.floor(kb * 100) / 100;
                return kb + " KB";
            } else {
                return bytes + " Bytes";
            }
        },

        close: function () {
            this.closeSignal++;
        },
    },
});
</script>
