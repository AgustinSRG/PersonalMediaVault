<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <div v-if="display" class="modal-dialog modal-lg" role="document">
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
                            <td>{{ $t("Version") }}</td>
                            <td>{{ version }}</td>
                        </tr>

                        <tr>
                            <td>{{ $t("Server version") }}</td>
                            <td v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</td>
                            <td v-else>{{ serverVersion }}</td>
                        </tr>

                        <tr>
                            <td>{{ $t("Update status") }}</td>
                            <td v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</td>
                            <td v-else>
                                <span v-if="serverVersion === lastRelease"> {{ $t("You are using the latest version") }}</span>
                                <span v-if="serverVersion !== lastRelease"
                                    ><a :href="getReleaseLink(lastRelease, gitRepo)" target="_blank" rel="noopener noreferrer"
                                        >{{ $t("Download the latest release") }}: {{ lastRelease }}</a
                                    ></span
                                >
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("FFmpeg version") }}</td>
                            <td v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</td>
                            <td v-else>{{ ffmpegVersion }}</td>
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

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { getUniqueStringId } from "@/utils/unique-id";
import { abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { apiAbout } from "@/api/api-about";
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";

export default defineComponent({
    name: "AboutModal",
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            requestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            version: import.meta.env.VITE__VERSION || "-",
            homePage: import.meta.env.VITE__HOME_URL || "#",
            gitRepo: import.meta.env.VITE__GIT_URL || "#",
            license: import.meta.env.VITE__LICENSE_URL || "#",

            loading: true,

            serverVersion: "",
            lastRelease: "",
            ffmpegVersion: "",

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.load();
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.load.bind(this));
        if (this.display) {
            this.load();
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);
        clearNamedTimeout(this.requestId);
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        load: function () {
            abortNamedApiRequest(this.requestId);
            clearNamedTimeout(this.requestId);

            if (!this.display || AuthController.Locked) {
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.requestId, apiAbout())
                .onSuccess((res) => {
                    this.loading = false;
                    this.serverVersion = res.version;
                    this.lastRelease = res.last_release;
                    this.ffmpegVersion = res.ffmpeg_version;
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            setNamedTimeout(this.requestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    setNamedTimeout(this.requestId, 1500, this.load.bind(this));
                });
        },

        getReleaseLink: function (version: string, gitRepository: string) {
            return gitRepository + "/releases/tag/v" + version;
        },
    },
});
</script>
