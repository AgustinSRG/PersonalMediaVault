<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy" :static="true" :close-callback="askClose">
        <form v-if="display" class="modal-dialog modal-xl modal-height-100-wf" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Advanced settings") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" :disabled="busy" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body no-padding">
                <LoadingOverlay></LoadingOverlay>
            </div>
            <div v-if="!loading" class="modal-body no-padding">
                <div class="horizontal-filter-menu modal-top-menu">
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('General')"
                        :class="{ selected: page === 'general' }"
                        @click="changePage('general')"
                        >{{ $t("General") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Extra resolutions')"
                        :class="{ selected: page == 'resolutions' }"
                        @click="changePage('resolutions')"
                        >{{ $t("Extra resolutions") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Custom style')"
                        :class="{ selected: page === 'css' }"
                        @click="changePage('css')"
                        >{{ $t("Custom style") }}</a
                    >
                </div>

                <div class="modal-body-content">
                    <div v-if="page === 'general'">
                        <div class="form-group">
                            <label>{{ $t("Vault title") }}:</label>
                            <input
                                v-model="title"
                                type="text"
                                autocomplete="off"
                                :disabled="busy"
                                maxlength="100"
                                :placeholder="$t('Personal Media Vault')"
                                class="form-control form-control-full-width"
                                @change="markDirty"
                            />
                        </div>

                        <div class="form-group">
                            <label>{{ $t("Logo text") }}:</label>
                            <input
                                v-model="logo"
                                type="text"
                                autocomplete="off"
                                :disabled="busy"
                                maxlength="20"
                                placeholder="PMV"
                                class="form-control form-control-full-width"
                                @change="markDirty"
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
                                @change="markDirty"
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
                                @change="markDirty"
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
                                @change="markDirty"
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
                                @change="markDirty"
                            />
                        </div>

                        <div class="form-group">
                            <table class="table no-border">
                                <tbody>
                                    <tr>
                                        <td class="text-right td-shrink no-padding">
                                            <ToggleSwitch v-model:val="preserveOriginals" @update:val="markDirty"></ToggleSwitch>
                                        </td>
                                        <td>
                                            {{ $t("Preserve original media files, before encoding, as attachments?") }}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <div v-else-if="page === 'resolutions'">
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
                                                <ToggleSwitch v-model:val="res.enabled" @update:val="markDirty"></ToggleSwitch>
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
                                                <ToggleSwitch v-model:val="res.enabled" @update:val="markDirty"></ToggleSwitch>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>

                    <div v-else-if="page === 'css'">
                        <div class="form-group">
                            <label>{{ $t("Custom style (css)") }}:</label>
                            <textarea
                                v-model="css"
                                :disabled="busy"
                                rows="12"
                                class="form-control form-control-full-width form-textarea"
                                :placeholder="'.main-layout.dark-theme {\n\tbackground: blue;\n}'"
                                @change="markDirty"
                            ></textarea>
                        </div>
                        <div>
                            {{ $t("Note: This is an advanced and possibly dangerous feature.") }}
                            {{ $t("Do not change this unless you know what you are doing.") }}
                        </div>
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

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="performRequest"
        ></AuthConfirmationModal>

        <SaveChangesAskModal v-model:display="displayAskSave" @yes="submit" @no="forceClose"></SaveChangesAskModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import type { ImageResolution, VaultUserConfig, VideoResolution } from "@/api/models";
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { onMounted, ref, useTemplateRef, watch } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { AuthController } from "@/control/auth";
import SaveChangesAskModal from "@/components/modals/SaveChangesAskModal.vue";
import { PagesController } from "@/control/pages";
import { apiConfigGetConfig, apiConfigSetConfig } from "@/api/api-config";
import type { ImageResolutionStandardToggleable, VideoResolutionStandardToggleable } from "@/utils/resolutions";
import { STANDARD_IMAGE_RESOLUTIONS, STANDARD_VIDEO_RESOLUTIONS } from "@/utils/resolutions";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose, focus } = useModal(display, container);

// Changes made?
const dirty = ref(false);

/**
 * Indicates changes were made
 * by the user
 */
const markDirty = () => {
    dirty.value = true;
};

// Ask for unsaved changes to be saved
const displayAskSave = ref(false);

/**
 * Asks before closing the modal
 * @param callback The callback
 */
const askClose = (callback: () => void) => {
    if (dirty.value) {
        displayAskSave.value = true;
    } else {
        callback();
    }
};

// Title
const title = ref("");

// Logo
const logo = ref("");

// Custom CSS
const css = ref("");

// Max parallel tasks
const maxTasks = ref(0);

// Max encoding threads
const encodingThreads = ref(0);

// Video previews interval
const videoPreviewsInterval = ref(0);

// Max number of invites
const inviteLimit = ref(0);

// Preserve original assets?
const preserveOriginals = ref(false);

// Video resolutions
const resolutions = ref<VideoResolutionStandardToggleable[]>([]);

// Image resolutions
const imageResolutions = ref<ImageResolutionStandardToggleable[]>([]);

/**
 * Updates resolution arrays
 * @param newResolutions New resolutions
 * @param newImageResolutions New image resolutions
 */
const updateResolutions = (newResolutions: VideoResolution[], newImageResolutions: ImageResolution[]) => {
    resolutions.value = STANDARD_VIDEO_RESOLUTIONS.map((r) => {
        let enabled = false;
        for (const res of newResolutions) {
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

    imageResolutions.value = STANDARD_IMAGE_RESOLUTIONS.map((r) => {
        let enabled = false;
        for (const res of newImageResolutions) {
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
};

// Loading status
const loading = ref(true);

// Load request ID
const loadRequestId = useRequestId();

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the data
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    loading.value = true;

    makeNamedApiRequest(loadRequestId, apiConfigGetConfig())
        .onSuccess((response: VaultUserConfig) => {
            title.value = response.title || "";
            logo.value = response.logo || "";
            css.value = response.css || "";
            maxTasks.value = response.max_tasks;
            encodingThreads.value = response.encoding_threads;
            videoPreviewsInterval.value = response.video_previews_interval;
            inviteLimit.value = response.invite_limit;
            preserveOriginals.value = response.preserve_originals || false;

            updateResolutions(response.resolutions, response.image_resolutions);

            loading.value = false;

            focus();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
        });
};

onMounted(() => {
    if (display.value) {
        load();
    }
});

watch(display, () => {
    if (display.value) {
        displayAskSave.value = false;

        load();
    }
});

// Page names
type LocalPageName = "general" | "resolutions" | "css";

// Current page
const page = ref<LocalPageName>("general");

/**
 * Changes the current page
 * @param p The page name
 */
const changePage = (p: LocalPageName) => {
    page.value = p;
};

// Auth confirmation
const {
    displayAuthConfirmation,
    authConfirmationCooldown,
    authConfirmationTfa,
    authConfirmationError,
    requiredAuthConfirmationPassword,
    invalidPassword,
    requiredAuthConfirmationTfa,
    invalidTfaCode,
    cooldown,
} = useAuthConfirmation();

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Resets the error messages
const resetErrors = () => {
    error.value = "";
};

/**
 * Performs the request
 * @param confirmation The auth confirmation
 */
const performRequest = (confirmation: ProvidedAuthConfirmation) => {
    if (busy.value) {
        return;
    }

    resetErrors();

    busy.value = true;

    makeApiRequest(
        apiConfigSetConfig(
            {
                title: title.value,
                logo: logo.value,
                css: css.value,
                max_tasks: maxTasks.value,
                encoding_threads: encodingThreads.value,
                resolutions: getResolutions(),
                image_resolutions: getImageResolutions(),
                video_previews_interval: videoPreviewsInterval.value,
                invite_limit: inviteLimit.value,
                preserve_originals: preserveOriginals.value,
            },
            confirmation,
        ),
    )
        .onSuccess(() => {
            busy.value = false;
            dirty.value = false;

            PagesController.ShowSnackBar($t("Vault configuration updated!"));

            AuthController.CheckAuthStatus();

            forceClose();
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                badRequest: () => {
                    error.value = $t("Invalid configuration provided");
                },
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            error.value = err.message;

            console.error(err);
        });
};

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    performRequest({});
};

/**
 * Gets the list of resolutions for the API
 */
const getResolutions = () => {
    return resolutions.value
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
};

/**
 * Gets the list of image resolutions for the API
 */
const getImageResolutions = () => {
    return imageResolutions.value
        .filter((r) => {
            return r.enabled;
        })
        .map((r) => {
            return {
                width: r.width,
                height: r.height,
            };
        });
};
</script>
