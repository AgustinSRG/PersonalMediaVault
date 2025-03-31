<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :static="true"
        :close-callback="askClose"
    >
        <form v-if="display" class="modal-dialog modal-xl" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Advanced settings") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" :disabled="busy" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body">
                <div class="horizontal-filter-menu">
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :class="{ selected: page === 'general' }"
                        @click="changePage('general')"
                        >{{ $t("General") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :class="{ selected: page == 'resolutions' }"
                        @click="changePage('resolutions')"
                        >{{ $t("Extra resolutions") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :class="{ selected: page === 'css' }"
                        @click="changePage('css')"
                        >{{ $t("Custom style") }}</a
                    >
                </div>

                <div class="form-group"></div>

                <div v-if="page === 'general'">
                    <div class="form-group">
                        <label>{{ $t("Vault title") }}:</label>
                        <input
                            v-model="title"
                            type="text"
                            autocomplete="off"
                            :disabled="busy"
                            :placeholder="$t('Personal Media Vault')"
                            class="form-control form-control-full-width"
                            @change="onChangesMade"
                        />
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Max number of tasks in parallel (0 for unlimited)") }}:</label>
                        <input
                            v-model.number="maxTasks"
                            type="number"
                            autocomplete="off"
                            :disabled="busy"
                            min="0"
                            class="form-control form-control-full-width"
                            @change="onChangesMade"
                        />
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Max number threads for each task (0 to use the number of cores)") }}:</label>
                        <input
                            v-model.number="encodingThreads"
                            type="number"
                            autocomplete="off"
                            :disabled="busy"
                            min="0"
                            class="form-control form-control-full-width"
                            @change="onChangesMade"
                        />
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Video previews interval (seconds) (if set to 0, by default is 3 seconds)") }}:</label>
                        <input
                            v-model.number="videoPreviewsInterval"
                            type="number"
                            autocomplete="off"
                            :disabled="busy"
                            min="0"
                            class="form-control form-control-full-width"
                            @change="onChangesMade"
                        />
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Max number of invited sessions by user (if set to 0, by default is 10 sessions)") }}:</label>
                        <input
                            v-model.number="inviteLimit"
                            type="number"
                            autocomplete="off"
                            :disabled="busy"
                            min="0"
                            class="form-control form-control-full-width"
                            @change="onChangesMade"
                        />
                    </div>
                </div>

                <div v-if="page === 'resolutions'">
                    <div class="form-group">
                        <label
                            >{{
                                $t("Extra resolutions for videos. These resolutions can be used for slow connections or small screens")
                            }}:</label
                        >
                        <div class="table-responsive">
                            <table class="table">
                                <thead>
                                    <tr>
                                        <th class="text-left">{{ $t("Name") }}</th>
                                        <th class="text-left">{{ $t("Properties") }}</th>
                                        <th class="text-right">{{ $t("Enabled") }}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="res in resolutions" :key="res.name">
                                        <td class="bold">{{ res.name }}</td>
                                        <td>{{ res.width }}x{{ res.height }}, {{ res.fps }} fps</td>
                                        <td class="text-right">
                                            <ToggleSwitch v-model:val="res.enabled" @update:val="onChangesMade"></ToggleSwitch>
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div class="form-group">
                        <label
                            >{{
                                $t("Extra resolutions for images. These resolutions can be used for slow connections or small screens")
                            }}:</label
                        >
                        <div class="table-responsive">
                            <table class="table">
                                <thead>
                                    <tr>
                                        <th class="text-left">{{ $t("Name") }}</th>
                                        <th class="text-left">{{ $t("Properties") }}</th>
                                        <th class="text-right">{{ $t("Enabled") }}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="res in imageResolutions" :key="res.name">
                                        <td class="bold">{{ res.name }}</td>
                                        <td>{{ res.width }}x{{ res.height }}</td>
                                        <td class="text-right">
                                            <ToggleSwitch v-model:val="res.enabled" @update:val="onChangesMade"></ToggleSwitch>
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                <div v-if="page === 'css'">
                    <div class="form-group">
                        <label>{{ $t("Custom style (css)") }}:</label>
                        <textarea
                            v-model="css"
                            :disabled="busy"
                            rows="12"
                            class="form-control form-control-full-width form-textarea"
                            :placeholder="'.main-layout.dark-theme {\n\tbackground: blue;\n}'"
                            @change="onChangesMade"
                        ></textarea>
                    </div>
                    <div>
                        {{ $t("Note: This is an advanced and possibly dangerous feature.") }}
                        {{ $t("Do not change this unless you know what you are doing.") }}
                    </div>
                </div>

                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Save changes") }}
                </button>
            </div>
        </form>

        <SaveChangesAskModal v-model:display="displayAskSave" @yes="submit" @no="closeForced"></SaveChangesAskModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { VaultUserConfig } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import SaveChangesAskModal from "@/components/modals/SaveChangesAskModal.vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiConfigGetConfig, apiConfigSetConfig } from "@/api/api-config";

export default defineComponent({
    name: "AdvancedSettingsModal",
    components: {
        ToggleSwitch,
        LoadingIcon,
        SaveChangesAskModal,
    },
    props: {
        display: Boolean,
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
            page: "general",

            dirty: false,
            displayAskSave: false,

            title: "",
            css: "",
            maxTasks: 0,
            encodingThreads: 0,
            videoPreviewsInterval: 0,
            inviteLimit: 0,
            resolutions: [],
            imageResolutions: [],

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

            loading: true,
            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                nextTick(() => {
                    this.$el.focus();
                });
                this.dirty = false;
                this.displayAskSave = false;
                this.load();
            }
        },
    },
    mounted: function () {
        this.load();
        if (this.display) {
            this.error = "";
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    methods: {
        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        askClose: function (callback: () => void) {
            if (this.dirty) {
                this.displayAskSave = true;
            } else {
                callback();
            }
        },

        changePage: function (page: string) {
            this.page = page;
        },

        onChangesMade: function () {
            this.dirty = true;
        },

        updateResolutions: function (resolutions, imageResolutions) {
            this.resolutions = this.standardResolutions.map((r) => {
                let enabled = false;
                for (const res of resolutions) {
                    if (res.width === r.width && res.height === r.height && res.fps === r.fps) {
                        enabled = true;
                        break;
                    }
                }
                return {
                    enabled: enabled,
                    name: r.name,
                    width: r.width,
                    height: r.height,
                    fps: r.fps,
                };
            });

            this.imageResolutions = this.standardResolutions
                .filter((r) => {
                    return r.fps === 30;
                })
                .map((r) => {
                    let enabled = false;
                    for (const res of imageResolutions) {
                        if (res.width === r.width && res.height === r.height) {
                            enabled = true;
                            break;
                        }
                    }
                    return {
                        enabled: enabled,
                        name: r.name,
                        width: r.width,
                        height: r.height,
                    };
                });
        },

        getResolutions: function () {
            return this.resolutions
                .filter((r) => {
                    return r.enabled;
                })
                .map((r) => {
                    return {
                        width: r.width,
                        height: r.height,
                        fps: r.fps,
                    };
                });
        },

        getImageResolutions: function () {
            return this.imageResolutions
                .filter((r) => {
                    return r.enabled;
                })
                .map((r) => {
                    return {
                        width: r.width,
                        height: r.height,
                    };
                });
        },

        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.loadRequestId, apiConfigGetConfig())
                .onSuccess((response: VaultUserConfig) => {
                    this.title = response.title;
                    this.css = response.css;
                    this.maxTasks = response.max_tasks;
                    this.encodingThreads = response.encoding_threads;
                    this.videoPreviewsInterval = response.video_previews_interval;
                    this.inviteLimit = response.invite_limit;
                    this.updateResolutions(response.resolutions, response.image_resolutions);
                    this.loading = false;

                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
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

        submit: function (e) {
            if (e) {
                e.preventDefault();
            }

            if (this.busy) {
                return;
            }

            this.busy = true;
            this.error = "";

            makeApiRequest(
                apiConfigSetConfig({
                    title: this.title,
                    css: this.css,
                    max_tasks: this.maxTasks,
                    encoding_threads: this.encodingThreads,
                    resolutions: this.getResolutions(),
                    image_resolutions: this.getImageResolutions(),
                    video_previews_interval: this.videoPreviewsInterval,
                    invite_limit: this.inviteLimit,
                }),
            )
                .onSuccess(() => {
                    this.busy = false;
                    this.dirty = false;
                    PagesController.ShowSnackBar(this.$t("Vault configuration updated!"));
                    AuthController.CheckAuthStatus();
                    this.close();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            this.error = this.$t("Invalid configuration provided");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.busy = false;
                });
        },

        close: function () {
            this.closeSignal++;
        },

        closeForced: function () {
            this.forceCloseSignal++;
        },
    },
});
</script>
