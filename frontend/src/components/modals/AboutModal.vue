<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title about-modal-title">
                    <img class="about-modal-logo" src="/img/icons/favicon.png" alt="PMV" />
                    {{ $t("Personal Media Vault") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding table-responsive">
                <table class="table table-text-overflow">
                    <tbody>
                        <tr>
                            <td>{{ $t("Client version") }}</td>
                            <td>{{ version }}</td>
                        </tr>

                        <tr>
                            <td>{{ $t("Server version") }}</td>
                            <td v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</td>
                            <td v-else>{{ serverVersion }}</td>
                        </tr>

                        <tr>
                            <td>{{ $t("FFmpeg version") }}</td>
                            <td v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</td>
                            <td v-else>{{ ffmpegVersion }}</td>
                        </tr>

                        <tr>
                            <td>{{ $t("Update status") }}</td>
                            <td v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</td>
                            <td v-else>
                                <span v-if="serverVersion === lastRelease"> {{ $t("You are using the latest version") }}</span>
                                <span v-if="serverVersion !== lastRelease"
                                    ><a :href="lastReleaseUrl" target="_blank" rel="noopener noreferrer"
                                        >{{ $t("Download the latest release") }}: {{ lastRelease }}</a
                                    ></span
                                >
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Home page") }}</td>
                            <td>
                                <a :href="homePage" target="_blank" rel="noopener noreferrer">{{ homePage }}</a>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Git repository") }}</td>
                            <td>
                                <a :href="gitRepo" target="_blank" rel="noopener noreferrer">{{ gitRepo }}</a>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("License") }}</td>
                            <td>
                                <a :href="license" target="_blank" rel="noopener noreferrer">MIT</a>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { computed, onMounted, ref, useTemplateRef, watch } from "vue";
import { abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { apiAbout } from "@/api/api-about";
import { emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Client version
const version = import.meta.env.VITE__VERSION || "-";

// Home page URL
const homePage = import.meta.env.VITE__HOME_URL || "#";

// Git repository URL
const gitRepo = import.meta.env.VITE__GIT_URL || "#";

// License URL
const license = import.meta.env.VITE__LICENSE_URL || "#";

// Server version
const serverVersion = ref("");

// Last release version
const lastRelease = ref("");

// FFmpeg version
const ffmpegVersion = ref("");

// Last release URL
const lastReleaseUrl = computed(() => {
    return gitRepo.value + "/releases/tag/v" + lastRelease.value;
});

// Request ID
const requestId = useRequestId();

// Loading status
const loading = ref(true);

/**
 * Loads the data
 */
const load = () => {
    abortNamedApiRequest(requestId);
    clearNamedTimeout(requestId);

    if (!display.value || AuthController.Locked) {
        return;
    }

    loading.value = true;

    makeNamedApiRequest(requestId, apiAbout())
        .onSuccess((res) => {
            loading.value = false;

            serverVersion.value = res.version;
            lastRelease.value = res.last_release;
            ffmpegVersion.value = res.ffmpeg_version;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    setNamedTimeout(requestId, 1500, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            setNamedTimeout(requestId, 1500, load);
        });
};

onMounted(() => {
    if (display.value) {
        load();
    }
});

watch(display, () => {
    if (display.value) {
        load();
    }
});

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, load);
</script>
