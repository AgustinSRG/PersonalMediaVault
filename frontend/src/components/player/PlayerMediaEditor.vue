<template>
    <div class="player-media-editor" tabindex="-1">
        <!--- Title -->

        <form @submit="changeTitle">
            <div class="form-group">
                <label>{{ $t("Title") }}:</label>
                <input
                    type="text"
                    autocomplete="off"
                    :readonly="!canWrite"
                    maxlength="255"
                    :disabled="busy"
                    v-model="title"
                    class="form-control form-control-full-width"
                />
            </div>
            <div class="form-group" v-if="canWrite">
                <button type="submit" class="btn btn-primary" :disabled="busy || !title || originalTitle === title">
                    <i class="fas fa-pencil-alt"></i> {{ $t("Change title") }}
                </button>
            </div>
        </form>

        <!--- Description -->

        <div class="form-group border-top">
            <label>{{ $t("Description") }}:</label>
            <textarea
                v-model="desc"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea"
                rows="3"
                :disabled="busy"
            ></textarea>
        </div>
        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-primary" :disabled="busy || originalDesc === desc" @click="changeDescription">
                <i class="fas fa-pencil-alt"></i> {{ $t("Change description") }}
            </button>
        </div>

        <!--- Tags -->

        <div class="form-group border-top">
            <label>{{ $t("Tags") }}:</label>
        </div>
        <div class="form-group media-tags">
            <label v-if="tags.length === 0">{{ $t("There are no tags yet for this media.") }}</label>
            <div v-for="tag in tags" :key="tag" class="media-tag">
                <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
                <button
                    v-if="canWrite"
                    type="button"
                    :title="$t('Remove tag')"
                    class="media-tag-btn"
                    :disabled="busy"
                    @click="removeTag(tag)"
                >
                    <i class="fas fa-times"></i>
                </button>
            </div>
        </div>
        <form @submit="addTag" v-if="canWrite">
            <div class="form-group">
                <label>{{ $t("Tag to add") }}:</label>
                <input
                    type="text"
                    autocomplete="off"
                    maxlength="255"
                    v-model="tagToAdd"
                    :disabled="busy"
                    @input="onTagAddChanged"
                    @keydown="onTagAddKeyDown"
                    class="form-control tag-to-add"
                />
            </div>
            <div class="form-group" v-if="matchingTags.length > 0">
                <button
                    v-for="mt in matchingTags"
                    :key="mt.id"
                    type="button"
                    class="btn btn-primary btn-sm btn-tag-mini"
                    :disabled="busy"
                    @click="addMatchingTag(mt.name)"
                >
                    <i class="fas fa-plus"></i> {{ mt.name }}
                </button>
            </div>
            <div class="form-group">
                <button type="submit" class="btn btn-primary" :disabled="busy || !tagToAdd">
                    <i class="fas fa-plus"></i> {{ $t("Add Tag") }}
                </button>
            </div>
        </form>

        <!--- Thumbnail -->

        <div class="form-group border-top">
            <label>{{ $t("Thumbnail") }}:</label>
        </div>
        <div class="form-group" @drop="onDrop">
            <label v-if="!thumbnail">{{ $t("No thumbnail set for this media") }}</label>
            <img v-if="thumbnail" :src="getThumbnail(thumbnail)" :alt="originalTitle" class="form-group-thumbnail" loading="lazy" />
        </div>
        <div class="form-group" v-if="canWrite">
            <input type="file" class="file-hidden" @change="inputFileChanged" name="thumbnail-upload" />
            <button type="button" class="btn btn-primary" :disabled="busy" @click="uploadThumbnail">
                <i class="fas fa-upload"></i> {{ $t("Upload new thumbnail") }}
            </button>
        </div>

        <!--- Subtitles -->

        <div class="form-group border-top" v-if="type === 2 || type === 3">
            <label>{{ $t("Subtitles") }}:</label>
        </div>

        <div v-if="type === 2 || type === 3" class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("ID") }}</th>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-right td-shrink"></th>
                        <th class="text-right td-shrink" v-if="canWrite"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="sub in subtitles" :key="sub.id">
                        <td class="bold">{{ sub.id }}</td>
                        <td class="bold">{{ sub.name }}</td>
                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" :disabled="busy" @click="downloadSubtitles(sub)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite">
                            <button type="button" class="btn btn-danger btn-xs" @click="removeSubtitles(sub)">
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("You can upload subtitles in SubRip format (.srt)") }}:</label>
            <input type="file" class="file-hidden srt-file-hidden" @change="srtFileChanged" name="srt-upload" accept=".srt" />
            <button v-if="!srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                <i class="fas fa-upload"></i> {{ $t("Select SRT file") }}
            </button>

            <button v-if="srtFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectSRTFile">
                <i class="fas fa-upload"></i> {{ $t("SRT file") }}: {{ srtFileName }}
            </button>
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("Subtitles identifier") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="srtId" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("Subtitles name") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="srtName" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <button type="button" class="btn btn-primary" :disabled="busy || !srtId || !srtName || !srtFile" @click="addSubtitles">
                <i class="fas fa-plus"></i> {{ $t("Add subtitles file") }}
            </button>
        </div>

        <!--- Audio tracks -->

        <div class="form-group border-top" v-if="type === 2">
            <label>{{ $t("Extra audio tracks") }}:</label>
        </div>

        <div v-if="type === 2" class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("ID") }}</th>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-right td-shrink"></th>
                        <th class="text-right td-shrink" v-if="canWrite"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="aud in audios" :key="aud.id">
                        <td class="bold">{{ aud.id }}</td>
                        <td class="bold">{{ aud.name }}</td>
                        <td class="text-right td-shrink">
                            <button type="button" class="btn btn-primary btn-xs" :disabled="busy" @click="downloadAudio(aud)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite">
                            <button type="button" class="btn btn-danger btn-xs" @click="removeAudio(aud)">
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group" v-if="canWrite && type === 2">
            <label>{{ $t("You can upload extra audio tracks for the video (.mp3)") }}:</label>
            <input type="file" class="file-hidden audio-file-hidden" @change="audioFileChanged" name="mp3-upload" accept=".mp3" />
            <button v-if="!audioFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectAudioFile">
                <i class="fas fa-upload"></i> {{ $t("Select audio file") }}
            </button>

            <button v-if="audioFileName" type="button" class="btn btn-primary" :disabled="busy" @click="selectAudioFile">
                <i class="fas fa-upload"></i> {{ $t("Audio file") }}: {{ audioFileName }}
            </button>
        </div>
        <div class="form-group" v-if="canWrite && type === 2">
            <label>{{ $t("Audio track identifier") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="audioId" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && type === 2">
            <label>{{ $t("Audio track name") }}:</label>
            <input type="text" autocomplete="off" maxlength="255" :disabled="busy" v-model="audioName" class="form-control" />
        </div>
        <div class="form-group" v-if="canWrite && type === 2">
            <button type="button" class="btn btn-primary" :disabled="busy || !audioId || !audioName || !audioFile" @click="addAudio">
                <i class="fas fa-plus"></i> {{ $t("Add audio track file") }}
            </button>
        </div>

        <!--- Attachments -->

        <div class="form-group border-top">
            <label>{{ $t("Attachments") }}:</label>
        </div>

        <div class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-left">{{ $t("Size") }}</th>
                        <th class="text-right"></th>
                        <th class="text-right td-shrink" v-if="canWrite"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="att in attachments" :key="att.id">
                        <td class="bold" v-if="attachmentEdit != att.id">{{ att.name }}</td>
                        <td v-if="attachmentEdit == att.id">
                            <input
                                type="text"
                                maxlength="255"
                                :disabled="busy"
                                class="form-control form-control-full-width"
                                v-model="attachmentEditName"
                            />
                        </td>
                        <td>{{ renderSize(att.size) }}</td>
                        <td class="text-right td-shrink one-line">
                            <button type="button" class="btn btn-primary btn-xs mr-1" :disabled="busy" @click="downloadAttachment(att)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink one-line" v-if="canWrite">
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy"
                                @click="editAttachment(att)"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy"
                                @click="saveEditAttachment"
                            >
                                <i class="fas fa-check"></i> {{ $t("Save") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy"
                                @click="cancelEditAttachment"
                            >
                                <i class="fas fa-times"></i> {{ $t("Cancel") }}
                            </button>
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy"
                                @click="removeAttachment(att)"
                            >
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group" v-if="canWrite">
            <input type="file" class="file-hidden attachment-file-hidden" @change="attachmentFileChanged" name="attachment-upload" />
            <button
                v-if="!busy || attachmentUploadProgress <= 0"
                type="button"
                class="btn btn-primary"
                :disabled="busy"
                @click="selectAttachmentFile"
            >
                <i class="fas fa-upload"></i> {{ $t("Upload attachment") }}
            </button>
            <button
                v-if="busy && attachmentUploadProgress > 0 && attachmentUploadProgress < 100"
                type="button"
                class="btn btn-primary"
                disabled
            >
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ attachmentUploadProgress }}%)
            </button>
            <button v-if="busy && attachmentUploadProgress >= 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
        </div>

        <!--- Time slices -->

        <div class="form-group border-top" v-if="type === 2 || type === 3">
            <label>{{ $t("Time slices") }}:</label>
            <textarea
                v-model="timeSlices"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea"
                :placeholder="'00:00:00 A\n00:01:00 B'"
                rows="5"
                :disabled="busy"
            ></textarea>
        </div>

        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <button type="button" class="btn btn-primary" :disabled="busy || originalTimeSlices === timeSlices" @click="changeTimeSlices">
                <i class="fas fa-pencil-alt"></i> {{ $t("Change time slices") }}
            </button>
        </div>

        <!--- Resolutions -->

        <div class="form-group border-top" v-if="canWrite && (type === 2 || type === 1)">
            <label v-if="type === 2"
                >{{ $t("Extra resolutions for videos. These resolutions can be used for slow connections or small screens") }}:</label
            >
            <label v-if="type === 1"
                >{{ $t("Extra resolutions for images. These resolutions can be used for slow connections or small screens") }}:</label
            >
        </div>

        <div class="form-group" v-if="canWrite && (type === 2 || type === 1)">
            <label v-if="type === 1">{{ $t("Original resolution") }}: {{ width }}x{{ height }}</label>
            <label v-if="type === 2"> {{ $t("Original resolution") }}: {{ width }}x{{ height }}, {{ fps }} fps </label>
        </div>

        <div v-if="canWrite && (type === 2 || type === 1)" class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-left">{{ $t("Properties") }}</th>
                        <th class="text-right"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="res in resolutions" :key="res.name">
                        <td class="bold">{{ res.name }}</td>
                        <td v-if="type === 1">
                            {{ renderResolutionProperties(res.width, res.height, width, height) }}
                        </td>
                        <td v-if="type === 2">{{ res.width }}x{{ res.height }}, {{ res.fps }} fps</td>
                        <td class="text-right">
                            <button
                                v-if="!res.enabled"
                                type="button"
                                class="btn btn-primary btn-xs"
                                :disabled="busy"
                                @click="addResolution(res)"
                            >
                                <i class="fas fa-plus"></i> {{ $t("Encode") }}
                            </button>
                            <button
                                v-if="res.enabled"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy"
                                @click="deleteResolution(res)"
                            >
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!--- Extra config -->

        <div class="form-group border-top" v-if="canWrite && (type === 2 || type === 3)">
            <label>{{ $t("Extra media configuration") }}:</label>
        </div>
        <div class="table-responsive" v-if="canWrite && (type === 2 || type === 3)">
            <table class="table">
                <tr v-if="type === 2 || type === 3">
                    <td class="">
                        {{ $t("Reset time to the beginning every time the media reloads?") }}
                    </td>
                    <td class="text-right">
                        <toggle-switch v-model:val="startBeginning" :disabled="busy"></toggle-switch>
                    </td>
                </tr>
            </table>
        </div>
        <div class="form-group" v-if="canWrite && (type === 2 || type === 3)">
            <button
                type="button"
                class="btn btn-primary"
                :disabled="busy || originalStartBeginning === startBeginning"
                @click="changeExtraParams"
            >
                <i class="fas fa-pencil-alt"></i> {{ $t("Change extra configuration") }}
            </button>
        </div>

        <!--- Re-Encode -->

        <div class="form-group border-top" v-if="canWrite">
            <label>
                {{ $t("If the media resource did not encode properly, try using the button below.") }}
                {{ $t("If it still does not work, try re-uploading the media.") }}
            </label>
        </div>
        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-primary" :disabled="busy" @click="encodeMedia">
                <i class="fas fa-sync-alt"></i> {{ $t("Re-Encode") }}
            </button>
        </div>

        <!--- Delete -->

        <div class="form-group border-top" v-if="canWrite">
            <label>{{ $t("If you want to delete this media resource, click the button below.") }}</label>
        </div>
        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-danger" :disabled="busy" @click="deleteMedia">
                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
            </button>
        </div>

        <MediaDeleteModal v-model:display="displayMediaDelete"></MediaDeleteModal>
        <ResolutionConfirmationModal
            ref="resolutionConfirmationModal"
            v-model:display="displayResolutionConfirmation"
        ></ResolutionConfirmationModal>
        <SubtitlesDeleteModal ref="subtitlesDeleteModal" v-model:display="displaySubtitlesDelete"></SubtitlesDeleteModal>
        <AudioTrackDeleteModal ref="audioTrackDeleteModal" v-model:display="displayAudioTrackDelete"></AudioTrackDeleteModal>
        <AttachmentDeleteModal ref="attachmentDeleteModal" v-model:display="displayAttachmentDelete"></AttachmentDeleteModal>
        <ReEncodeConfirmationModal v-model:display="displayReEncode" @confirm="doEncodeMedia"></ReEncodeConfirmationModal>
    </div>
</template>

<script lang="ts">
import { MediaAPI, MediaAttachment, MediaAudioTrack, MediaSubtitle } from "@/api/api-media";
import { TagsAPI } from "@/api/api-tags";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
import { TagsController } from "@/control/tags";
import { MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO } from "@/utils/constants";
import { clone } from "@/utils/objects";
import { GetAssetURL, Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";

import MediaDeleteModal from "../modals/MediaDeleteModal.vue";
import ResolutionConfirmationModal from "../modals/ResolutionConfirmationModal.vue";
import ReEncodeConfirmationModal from "../modals/ReEncodeConfirmationModal.vue";
import SubtitlesDeleteModal from "../modals/SubtitlesDeleteModal.vue";
import AudioTrackDeleteModal from "../modals/AudioTrackDeleteModal.vue";
import AttachmentDeleteModal from "../modals/AttachmentDeleteModal.vue";
import { parseTimeSlices, renderTimeSlices } from "@/utils/time-slices";

export default defineComponent({
    components: {
        ToggleSwitch,
        MediaDeleteModal,
        ResolutionConfirmationModal,
        ReEncodeConfirmationModal,
        SubtitlesDeleteModal,
        AudioTrackDeleteModal,
        AttachmentDeleteModal,
    },
    name: "PlayerMediaEditor",
    emits: ["changed"],
    data: function () {
        return {
            type: 0,

            title: "",
            originalTitle: "",

            desc: "",
            originalDesc: "",

            timeSlices: "",
            originalTimeSlices: "",

            tags: [],
            tagToAdd: "",
            tagData: {},
            matchingTags: [],

            thumbnail: "",

            width: 0,
            height: 0,
            fps: 0,

            standardResolutions: [
                {
                    name: "144p",
                    width: 256,
                    height: 144,
                    fps: 30,
                },
                {
                    name: "240p",
                    width: 352,
                    height: 240,
                    fps: 30,
                },
                {
                    name: "360p",
                    width: 480,
                    height: 360,
                    fps: 30,
                },
                {
                    name: "480p",
                    width: 858,
                    height: 480,
                    fps: 30,
                },
                {
                    name: "720p",
                    width: 1280,
                    height: 720,
                    fps: 30,
                },
                {
                    name: "720p60",
                    width: 1280,
                    height: 720,
                    fps: 60,
                },
                {
                    name: "1080p",
                    width: 1920,
                    height: 1080,
                    fps: 30,
                },
                {
                    name: "1080p60",
                    width: 1920,
                    height: 1080,
                    fps: 60,
                },
                {
                    name: "2k",
                    width: 2048,
                    height: 1152,
                    fps: 30,
                },
                {
                    name: "2k60",
                    width: 2048,
                    height: 1152,
                    fps: 60,
                },
                {
                    name: "4k",
                    width: 3860,
                    height: 2160,
                    fps: 30,
                },
                {
                    name: "4k60",
                    width: 3860,
                    height: 2160,
                    fps: 60,
                },
            ],

            resolutions: [],

            subtitles: [],
            srtFile: null,
            srtFileName: "",
            srtId: "en",
            srtName: "English",

            audios: [],
            audioFile: null,
            audioFileName: "",
            audioId: "en",
            audioName: "English",

            attachments: [],
            attachmentUploadProgress: 0,
            attachmentEdit: -1,
            attachmentEditName: "",

            busy: false,

            canWrite: AuthController.CanWrite,

            originalStartBeginning: false,
            startBeginning: false,

            displayMediaDelete: false,

            displayResolutionConfirmation: false,

            displaySubtitlesDelete: false,

            displayAudioTrackDelete: false,

            displayAttachmentDelete: false,

            displayReEncode: false,
        };
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.originalTitle = MediaController.MediaData.title;
            this.title = this.originalTitle;

            this.originalDesc = MediaController.MediaData.description;
            this.desc = this.originalDesc;

            this.originalStartBeginning = MediaController.MediaData.force_start_beginning;
            this.startBeginning = this.originalStartBeginning;

            this.width = MediaController.MediaData.width;
            this.height = MediaController.MediaData.height;
            this.fps = MediaController.MediaData.fps;

            this.tags = (MediaController.MediaData.tags || []).slice();

            this.thumbnail = MediaController.MediaData.thumbnail;

            this.subtitles = (MediaController.MediaData.subtitles || []).map((a) => {
                return {
                    id: a.id,
                    name: a.name,
                    url: a.url,
                };
            });

            this.audios = (MediaController.MediaData.audios || []).map((a) => {
                return {
                    id: a.id,
                    name: a.name,
                    url: a.url,
                };
            });

            this.attachments = (MediaController.MediaData.attachments || []).map((a) => {
                return {
                    id: a.id,
                    name: a.name,
                    size: a.size,
                    url: a.url,
                };
            });
            this.attachmentUploadProgress = 0;
            this.attachmentEdit = -1;
            this.attachmentEditName = "";

            this.originalTimeSlices = renderTimeSlices(MediaController.MediaData.time_slices);
            this.timeSlices = this.originalTimeSlices;

            this.updateResolutions(MediaController.MediaData.resolutions || []);
        },

        updateResolutions: function (resolutions) {
            this.resolutions = this.standardResolutions
                .filter((r) => {
                    if (this.type === MEDIA_TYPE_IMAGE) {
                        return r.fps === 30;
                    } else if (this.type === MEDIA_TYPE_VIDEO) {
                        return true;
                    } else {
                        return false;
                    }
                })
                .map((r) => {
                    let enabled = false;
                    let fps = r.fps;
                    for (let res of resolutions) {
                        if (res.width === r.width && res.height === r.height && (this.type === MEDIA_TYPE_IMAGE || res.fps === r.fps)) {
                            enabled = true;
                            fps = res.fps;
                            break;
                        }
                    }
                    return {
                        enabled: enabled,
                        name: r.name,
                        width: r.width,
                        height: r.height,
                        fps: fps,
                    };
                });
        },

        getThumbnail(thumb: string) {
            return GetAssetURL(thumb);
        },

        uploadThumbnail: function () {
            this.$el.querySelector(".file-hidden").click();
        },

        inputFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.changeThumbnail(file);
            }
        },

        onDrop: function (e) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.changeThumbnail(file);
            }
        },

        changeThumbnail: function (file) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.ChangeMediaThumbnail(mediaId, file))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Successfully changed thumbnail"));
                    this.busy = false;
                    this.thumbnail = res.url;
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit("media-meta-change");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid thumbnail provided"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        changeTitle: function (e) {
            if (e) {
                e.preventDefault();
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.ChangeMediaTitle(mediaId, this.title))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed title"));
                    this.busy = false;
                    this.originalTitle = this.title;
                    this.$emit("changed");
                    AlbumsController.LoadCurrentAlbum();
                    AppEvents.Emit("media-meta-change");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        changeDescription: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.ChangeMediaDescription(mediaId, this.desc))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed description"));
                    this.busy = false;
                    this.originalDesc = this.desc;
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        changeTimeSlices: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            const slices = parseTimeSlices(this.timeSlices);

            Request.Pending("media-editor-busy", MediaAPI.ChangeTimeSlices(mediaId, slices))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed time slices"));
                    this.busy = false;
                    this.originalTimeSlices = renderTimeSlices(slices);
                    this.timeSlices = this.originalTimeSlices;
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        changeExtraParams: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.ChangeExtraParams(mediaId, this.startBeginning))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed media extra params"));
                    this.busy = false;
                    this.originalStartBeginning = this.startBeginning;
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        doEncodeMedia: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.EncodeMedia(mediaId))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully requested pending encoding tasks"));
                    this.busy = false;
                    MediaController.Load();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        encodeMedia: function () {
            this.displayReEncode = true;
        },

        deleteMedia: function () {
            this.displayMediaDelete = true;
        },

        updateTagData: function () {
            this.tagData = clone(TagsController.Tags);
        },

        getTagName: function (tag, data) {
            if (data[tag + ""]) {
                return data[tag + ""].name;
            } else {
                return "???";
            }
        },

        removeTag: function (tag) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const tagName = this.getTagName(tag, this.tagData);

            Request.Pending("media-editor-busy", TagsAPI.UntagMedia(mediaId, tag))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Removed tag") + ": " + tagName);
                    this.busy = false;
                    for (let i = 0; i < this.tags.length; i++) {
                        if (this.tags[i] === tag) {
                            this.tags.splice(i, 1);
                            break;
                        }
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid tag name"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        addTag: function (e) {
            if (e) {
                e.preventDefault();
            }
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const tag = this.tagToAdd;

            Request.Pending("media-editor-busy", TagsAPI.TagMedia(mediaId, tag))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    this.tagToAdd = "";
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    this.$emit("changed");
                    nextTick(() => {
                        const elemFocus = this.$el.querySelector(".tag-to-add");

                        if (elemFocus) {
                            elemFocus.focus();
                        }
                    });
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid tag name"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        addMatchingTag: function (tag) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", TagsAPI.TagMedia(mediaId, tag))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added tag") + ": " + res.name);
                    this.busy = false;
                    if (this.tags.indexOf(res.id) === -1) {
                        this.tags.push(res.id);
                    }
                    this.findTags();
                    TagsController.AddTag(res.id, res.name);
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Invalid tag name"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        onTagAddChanged: function () {
            if (this._handles.findTagTimeout) {
                return;
            }
            this._handles.findTagTimeout = setTimeout(() => {
                this._handles.findTagTimeout = null;
                this.findTags();
            }, 200);
        },

        findTags: function () {
            const tagFilter = this.tagToAdd
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();
            if (!tagFilter) {
                this.matchingTags = [];
                return;
            }
            this.matchingTags = Object.values(this.tagData)
                .map((a: any) => {
                    const i = a.name.indexOf(tagFilter);
                    return {
                        id: a.id,
                        name: a.name,
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    if (this.tags.indexOf(a.id) >= 0) {
                        return false;
                    }
                    return a.starts || a.contains;
                })
                .sort((a, b) => {
                    if (a.starts && !b.starts) {
                        return -1;
                    } else if (b.starts && !a.starts) {
                        return 1;
                    } else if (a.name < b.name) {
                        return -1;
                    } else {
                        return 1;
                    }
                })
                .slice(0, 10);
        },

        onTagAddKeyDown: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();
                this.addTag();
            } else if (e.key === "Tab") {
                this.findTags();
                if (this.matchingTags.length > 0) {
                    if (this.matchingTags[0].name != this.tagToAdd) {
                        e.preventDefault();
                        this.tagToAdd = this.matchingTags[0].name;
                    }
                }
            }
        },

        addResolution: function (r) {
            this.$refs.resolutionConfirmationModal.show({
                type: this.type,
                deleting: false,
                name: r.name,
                width: r.width,
                height: r.height,
                fps: r.fps,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;

                    Request.Pending("media-editor-busy", MediaAPI.AddResolution(mediaId, r.width, r.height, r.fps))
                        .onSuccess((result) => {
                            AppEvents.Emit("snack", this.$t("Added resolution") + ": " + r.name);
                            this.busy = false;
                            r.enabled = true;
                            r.fps = result.fps;
                            this.$emit("changed");
                        })
                        .onCancel(() => {
                            this.busy = false;
                        })
                        .onRequestError((err) => {
                            this.busy = false;
                            Request.ErrorHandler()
                                .add(400, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Bad request"));
                                })
                                .add(401, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                    AppEvents.Emit("unauthorized");
                                })
                                .add(403, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                })
                                .add(404, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Not found"));
                                })
                                .add(500, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Internal server error"));
                                })
                                .add("*", "*", () => {
                                    AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                                })
                                .handle(err);
                        })
                        .onUnexpectedError((err) => {
                            AppEvents.Emit("snack", err.message);
                            console.error(err);
                            this.busy = false;
                        });
                },
            });
        },

        deleteResolution: function (r) {
            this.$refs.resolutionConfirmationModal.show({
                type: this.type,
                deleting: true,
                name: r.name,
                width: r.width,
                height: r.height,
                fps: r.fps,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;

                    Request.Pending("media-editor-busy", MediaAPI.RemoveResolution(mediaId, r.width, r.height, r.fps))
                        .onSuccess(() => {
                            AppEvents.Emit("snack", this.$t("Removed resolution") + ": " + r.name);
                            this.busy = false;
                            r.enabled = false;
                            this.$emit("changed");
                        })
                        .onCancel(() => {
                            this.busy = false;
                        })
                        .onRequestError((err) => {
                            this.busy = false;
                            Request.ErrorHandler()
                                .add(400, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Bad request"));
                                })
                                .add(401, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                    AppEvents.Emit("unauthorized");
                                })
                                .add(403, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                })
                                .add(404, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Not found"));
                                })
                                .add(500, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Internal server error"));
                                })
                                .add("*", "*", () => {
                                    AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                                })
                                .handle(err);
                        })
                        .onUnexpectedError((err) => {
                            AppEvents.Emit("snack", err.message);
                            console.error(err);
                            this.busy = false;
                        });
                },
            });
        },

        // Subtitles

        selectSRTFile: function () {
            this.$el.querySelector(".srt-file-hidden").click();
        },

        srtFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.srtFile = file;
                this.srtFileName = file.name;
            }
        },

        addSubtitles: function () {
            if (!this.srtFile) {
                AppEvents.Emit("snack", this.$t("Please, select a SubRip file first"));
                return;
            }

            const id = this.srtId;
            const name = this.srtName;

            let duped = false;
            for (let sub of this.subtitles) {
                if (sub.id === id) {
                    duped = true;
                    break;
                }
            }

            if (duped) {
                AppEvents.Emit("snack", this.$t("There is already another subtitles file with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.SetSubtitles(mediaId, id, name, this.srtFile))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added subtitles") + ": " + res.name);
                    this.busy = false;
                    this.subtitles.push(res);
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "INVALID_SRT", () => {
                            AppEvents.Emit("snack", this.$t("Invalid SubRip file"));
                        })
                        .add(400, "INVALID_ID", () => {
                            AppEvents.Emit("snack", this.$t("Invalid subtitles identifier"));
                        })
                        .add(400, "INVALID_NAME", () => {
                            AppEvents.Emit("snack", this.$t("Invalid subtitles name"));
                        })
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(413, "*", () => {
                            AppEvents.Emit("snack", this.$t("Subtitles file too big (max is $MAX)").replace("$MAX", "10MB"));
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        removeSubtitles: function (sub) {
            this.$refs.subtitlesDeleteModal.show({
                name: sub.name,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;
                    const id = sub.id;

                    Request.Pending("media-editor-busy", MediaAPI.RemoveSubtitles(mediaId, id))
                        .onSuccess(() => {
                            AppEvents.Emit("snack", this.$t("Removed subtitles") + ": " + sub.name);
                            this.busy = false;
                            for (let i = 0; i < this.subtitles.length; i++) {
                                if (this.subtitles[i].id === id) {
                                    this.subtitles.splice(i, 1);
                                    break;
                                }
                            }
                            this.$emit("changed");
                        })
                        .onCancel(() => {
                            this.busy = false;
                        })
                        .onRequestError((err) => {
                            this.busy = false;
                            Request.ErrorHandler()
                                .add(400, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Bad request"));
                                })
                                .add(401, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                    AppEvents.Emit("unauthorized");
                                })
                                .add(403, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                })
                                .add(404, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Not found"));
                                })
                                .add(500, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Internal server error"));
                                })
                                .add("*", "*", () => {
                                    AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                                })
                                .handle(err);
                        })
                        .onUnexpectedError((err) => {
                            AppEvents.Emit("snack", err.message);
                            console.error(err);
                            this.busy = false;
                        });
                },
            });
        },

        downloadSubtitles: function (sub: MediaSubtitle) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = GetAssetURL(sub.url);
            link.click();
        },

        // Audios

        selectAudioFile: function () {
            this.$el.querySelector(".audio-file-hidden").click();
        },

        audioFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.audioFile = file;
                this.audioFileName = file.name;
            }
        },

        addAudio: function () {
            if (!this.audioFile) {
                AppEvents.Emit("snack", this.$t("Please, select an audio file first"));
                return;
            }

            const id = this.audioId;
            const name = this.audioName;

            let duped = false;
            for (let aud of this.audios) {
                if (aud.id === id) {
                    duped = true;
                    break;
                }
            }

            if (duped) {
                AppEvents.Emit("snack", this.$t("There is already another audio track with the same identifier"));
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.SetAudioTrack(mediaId, id, name, this.audioFile))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added audio track") + ": " + res.name);
                    this.busy = false;
                    this.audios.push(res);
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "INVALID_AUDIO", () => {
                            AppEvents.Emit("snack", this.$t("Invalid audio file"));
                        })
                        .add(400, "INVALID_ID", () => {
                            AppEvents.Emit("snack", this.$t("Invalid audio track identifier"));
                        })
                        .add(400, "INVALID_NAME", () => {
                            AppEvents.Emit("snack", this.$t("Invalid audio track name"));
                        })
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        removeAudio: function (aud: MediaAudioTrack) {
            this.$refs.audioTrackDeleteModal.show({
                name: aud.name,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;
                    const id = aud.id;

                    Request.Pending("media-editor-busy", MediaAPI.RemoveAudioTrack(mediaId, id))
                        .onSuccess(() => {
                            AppEvents.Emit("snack", this.$t("Removed audio track") + ": " + aud.name);
                            this.busy = false;
                            for (let i = 0; i < this.audios.length; i++) {
                                if (this.audios[i].id === id) {
                                    this.audios.splice(i, 1);
                                    break;
                                }
                            }
                            this.$emit("changed");
                        })
                        .onCancel(() => {
                            this.busy = false;
                        })
                        .onRequestError((err) => {
                            this.busy = false;
                            Request.ErrorHandler()
                                .add(400, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Bad request"));
                                })
                                .add(401, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                    AppEvents.Emit("unauthorized");
                                })
                                .add(403, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                })
                                .add(404, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Not found"));
                                })
                                .add(500, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Internal server error"));
                                })
                                .add("*", "*", () => {
                                    AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                                })
                                .handle(err);
                        })
                        .onUnexpectedError((err) => {
                            AppEvents.Emit("snack", err.message);
                            console.error(err);
                            this.busy = false;
                        });
                },
            });
        },

        downloadAudio: function (aud: MediaAudioTrack) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = GetAssetURL(aud.url);
            link.click();
        },

        // Attachments

        selectAttachmentFile: function () {
            this.$el.querySelector(".attachment-file-hidden").click();
        },

        attachmentFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.addAttachment(file);
            }
        },

        addAttachment: function (file: File) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            Request.Pending("media-editor-busy", MediaAPI.UploadAttachment(mediaId, file))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Added attachment") + ": " + res.name);
                    this.busy = false;
                    this.attachmentUploadProgress = 0;
                    this.attachments.push(res);
                    this.$emit("changed");
                })
                .onUploadProgress((loaded, total) => {
                    if (total) {
                        this.attachmentUploadProgress = Math.floor(((loaded * 100) / total) * 100) / 100;
                    }
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        removeAttachment: function (att: MediaAttachment) {
            this.$refs.attachmentDeleteModal.show({
                name: att.name,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;
                    const id = att.id;

                    Request.Pending("media-editor-busy", MediaAPI.RemoveAttachment(mediaId, id))
                        .onSuccess(() => {
                            AppEvents.Emit("snack", this.$t("Removed attachment") + ": " + att.name);
                            this.busy = false;
                            for (let i = 0; i < this.attachments.length; i++) {
                                if (this.attachments[i].id === id) {
                                    this.attachments.splice(i, 1);
                                    break;
                                }
                            }
                            this.$emit("changed");
                        })
                        .onCancel(() => {
                            this.busy = false;
                        })
                        .onRequestError((err) => {
                            this.busy = false;
                            Request.ErrorHandler()
                                .add(400, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Bad request"));
                                })
                                .add(401, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                    AppEvents.Emit("unauthorized");
                                })
                                .add(403, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Access denied"));
                                })
                                .add(404, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Not found"));
                                })
                                .add(500, "*", () => {
                                    AppEvents.Emit("snack", this.$t("Internal server error"));
                                })
                                .add("*", "*", () => {
                                    AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                                })
                                .handle(err);
                        })
                        .onUnexpectedError((err) => {
                            AppEvents.Emit("snack", err.message);
                            console.error(err);
                            this.busy = false;
                        });
                },
            });
        },

        downloadAttachment: function (att: MediaAttachment) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = GetAssetURL(att.url);
            link.click();
        },

        editAttachment: function (att: MediaAttachment) {
            this.attachmentEdit = att.id;
            this.attachmentEditName = att.name;
        },

        cancelEditAttachment: function () {
            this.attachmentEdit = -1;
            this.attachmentEditName = "";
        },

        saveEditAttachment: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const id = this.attachmentEdit;

            Request.Pending("media-editor-busy", MediaAPI.RenameAttachment(mediaId, id, this.attachmentEditName))
                .onSuccess((res) => {
                    AppEvents.Emit("snack", this.$t("Renamed attachment") + ": " + res.name);
                    this.busy = false;
                    this.attachmentEdit = -1;
                    this.attachmentEditName = "";
                    for (let i = 0; i < this.attachments.length; i++) {
                        if (this.attachments[i].id === res.id) {
                            this.attachments[i].name = res.name;
                            this.attachments[i].url = res.url;
                            break;
                        }
                    }
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "INVALID_NAME", () => {
                            AppEvents.Emit("snack", this.$t("Invalid attachment name"));
                        })
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        // --

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
            if (!this.canWrite) {
                this.editAttachment = -1;
            }
        },

        renderResolutionProperties: function (resWidth: number, resHeight: number, originalWidth: number, originalHeight: number): string {
            let width = originalWidth;
            let height = originalHeight;

            if (width > height) {
                const proportionalHeight = Math.round((height * resWidth) / width);

                if (proportionalHeight > resHeight) {
                    width = Math.round((width * resHeight) / height);
                    height = resHeight;
                } else {
                    width = resWidth;
                    height = proportionalHeight;
                }
            } else {
                const proportionalWidth = Math.round((width * resHeight) / height);

                if (proportionalWidth > resWidth) {
                    height = Math.round((height * resWidth) / width);
                    width = resWidth;
                } else {
                    width = proportionalWidth;
                    height = resHeight;
                }
            }

            return width + "x" + height;
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
    },

    mounted: function () {
        this._handles = Object.create(null);
        this.updateMediaData();
        this.updateTagData();

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this._handles.mediaUpdateH);

        this._handles.tagUpdateH = this.updateTagData.bind(this);

        AppEvents.AddEventListener("tags-update", this._handles.tagUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);

        TagsController.Load();
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);

        AppEvents.RemoveEventListener("tags-update", this._handles.tagUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        if (this._handles.findTagTimeout) {
            clearTimeout(this._handles.findTagTimeout);
        }

        Request.Abort("media-editor-busy");
    },
});
</script>
