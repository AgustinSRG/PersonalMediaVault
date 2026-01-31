<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            v-model:display="displayStatus"
            :title="$t('Description') + ' | ' + (title || $t('Untitled'))"
            :context-open="contextOpen"
            :position-key="'desc-widget-pos'"
            :busy="busy"
            :action-buttons="actionButtons"
            @clicked="propagateClick"
            @action-btn="clickActionButton"
        >
            <div class="media-description-body" :class="{ editing: editing }" tabindex="-1">
                <LoadingOverlay v-if="loading"></LoadingOverlay>
                <div v-else-if="editing" class="media-description-edit">
                    <textarea
                        v-model="contentToChange"
                        :disabled="busy || !canWrite"
                        class="form-control form-textarea auto-focus"
                        :placeholder="$t('Input your description here') + '...'"
                    ></textarea>
                </div>
                <div v-else-if="configuring" class="widget-settings">
                    <table>
                        <tbody>
                            <tr>
                                <td>{{ $t("Font size") }}:</td>
                                <td>
                                    <input
                                        v-model.number="baseFontSize"
                                        type="range"
                                        class="form-range"
                                        :min="1"
                                        :max="64"
                                        :step="1"
                                        @input="onFontSizeChange"
                                    />
                                </td>
                                <td>
                                    <input
                                        v-model.number="baseFontSize"
                                        type="number"
                                        class="form-control form-control-full-width"
                                        :min="1"
                                        :max="128"
                                        :step="1"
                                        @input="onFontSizeChange"
                                    />
                                </td>
                            </tr>
                            <tr v-if="speechSynthesisAvailable">
                                <td>{{ $t("Enable voice reading?") }}</td>
                                <td colspan="2">
                                    <toggle-switch
                                        v-model:val="voiceReadingSettings.enabled"
                                        @update:val="onChangedVoiceReadingSettings"
                                    ></toggle-switch>
                                </td>
                            </tr>
                            <tr v-if="speechSynthesisAvailable">
                                <td>{{ $t("Preferred voice") }}:</td>
                                <td colspan="2">
                                    <select
                                        v-model="voiceReadingSettings.voice"
                                        class="form-control form-control-full-width"
                                        @change="onChangedVoiceReadingSettings"
                                    >
                                        <option :value="''">--- {{ $t("Default voice") }} ---</option>
                                        <option v-for="v in voices" :key="v.voiceURI" :value="v.voiceURI">
                                            {{ v.name }} ({{ v.lang }})
                                        </option>
                                    </select>
                                </td>
                            </tr>
                            <tr v-if="speechSynthesisAvailable">
                                <td>{{ $t("Volume") }}:</td>
                                <td>
                                    <input
                                        v-model.number="voiceReadingSettings.volume"
                                        type="range"
                                        class="form-range"
                                        :min="0"
                                        :max="1"
                                        :step="0.01"
                                        @input="onChangedVoiceReadingSettings"
                                    />
                                </td>
                                <td>
                                    <input
                                        v-model.number="voiceReadingSettings.volume"
                                        type="number"
                                        class="form-control form-control-full-width"
                                        :min="0"
                                        :max="1"
                                        :step="0.01"
                                        @input="onChangedVoiceReadingSettings"
                                    />
                                </td>
                            </tr>
                            <tr v-if="speechSynthesisAvailable">
                                <td>{{ $t("Voice pitch") }}:</td>
                                <td>
                                    <input
                                        v-model.number="voiceReadingSettings.pitch"
                                        type="range"
                                        class="form-range"
                                        :min="0"
                                        :max="2"
                                        :step="0.01"
                                        @input="onChangedVoiceReadingSettings"
                                    />
                                </td>
                                <td>
                                    <input
                                        v-model.number="voiceReadingSettings.pitch"
                                        type="number"
                                        class="form-control form-control-full-width"
                                        :min="0"
                                        :max="2"
                                        :step="0.01"
                                        @input="onChangedVoiceReadingSettings"
                                    />
                                </td>
                            </tr>
                            <tr v-if="speechSynthesisAvailable">
                                <td>{{ $t("Reading rate") }}:</td>
                                <td>
                                    <input
                                        v-model.number="voiceReadingSettings.rate"
                                        type="range"
                                        class="form-range"
                                        :min="0.1"
                                        :max="10"
                                        :step="0.01"
                                        @input="onChangedVoiceReadingSettings"
                                    />
                                </td>
                                <td>
                                    <input
                                        v-model.number="voiceReadingSettings.rate"
                                        type="number"
                                        class="form-control form-control-full-width"
                                        :min="0.1"
                                        :max="10"
                                        :step="0.01"
                                        @input="onChangedVoiceReadingSettings"
                                    />
                                </td>
                            </tr>
                            <tr v-if="speechSynthesisAvailable">
                                <td>{{ $t("Voice sample") }}:</td>
                                <td>
                                    <input
                                        v-model="sampleVoiceText"
                                        type="text"
                                        class="form-control form-control-full-width"
                                        :disabled="playingSample"
                                    />
                                </td>
                                <td class="text-right">
                                    <button v-if="playingSample" type="button" class="btn btn-xs btn-primary" @click="stopPlayingSample">
                                        <i class="fas fa-times"></i> {{ $t("Stop") }}
                                    </button>
                                    <button
                                        v-else
                                        type="button"
                                        class="btn btn-xs btn-primary"
                                        :disabled="!sampleVoiceText"
                                        @click="playSample"
                                    >
                                        <i class="fas fa-play"></i> {{ $t("Play") }}
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div
                    v-else
                    class="media-description-view"
                    :class="{ reading: reading }"
                    :style="{ '--base-font-size': baseFontSize + 'px' }"
                >
                    <p
                        v-for="(line, i) in contentLines"
                        :id="paragraphIdPrefix + '-' + i"
                        :key="i"
                        class="media-description-paragraph"
                        :class="{ reading: reading && readingLine === i }"
                        @click="readSkipToLine(i)"
                        v-html="renderContent(line)"
                    ></p>
                </div>
            </div>
        </ResizableWidget>
    </div>
</template>

<script lang="ts">
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

import type { ActionButton } from "@/components/player/ResizableWidget.vue";
import ResizableWidget from "@/components/player/ResizableWidget.vue";
import { nextTick } from "vue";
import { AuthController } from "@/control/auth";
import { AppStatus } from "@/control/app-status";
import {
    emitAppEvent,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_DESCRIPTION_UPDATE,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest, RequestErrorHandler, makeApiRequest } from "@asanrom/request-browser";
import { MediaController } from "@/control/media";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { getAssetURL } from "@/utils/api";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { apiMediaSetDescription } from "@/api/api-media-edit";
import { escapeHTML, replaceLinks } from "@/utils/html";

import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import ToggleSwitch from "@/components/utils/ToggleSwitch.vue";
import {
    getDescriptionSize,
    getDescriptionWidgetReadSettings,
    setDescriptionSize,
    setDescriptionWidgetReadSettings,
} from "@/control/player-preferences";
import { getVoiceToSpeak, isSpeechSynthesisAvailable } from "@/utils/voice-synthesis";
import { clone } from "@/utils/objects";

function elementIsVisibleVertical(elementRect: DOMRect, containerRect: DOMRect) {
    const elementY1 = elementRect.top;
    const elementY2 = elementY1 + elementRect.height;

    const containerY1 = containerRect.top;
    const containerY2 = containerY1 + containerRect.height;

    return !(elementY2 < containerY1 || containerY2 < elementY1);
}

export default defineComponent({
    name: "DescriptionWidget",
    components: {
        ResizableWidget,
        LoadingOverlay,
        ToggleSwitch,
    },
    props: {
        display: Boolean,
        contextOpen: Boolean,
        title: String,
    },
    emits: ["update:display", "clicked", "update-desc"],
    setup(props) {
        return {
            speechSynthesisAvailable: isSpeechSynthesisAvailable(),
            voices: isSpeechSynthesisAvailable() ? speechSynthesis.getVoices() : [],
            loadRequestId: getUniqueStringId(),
            paragraphIdPrefix: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
            busyTimeout: null as ReturnType<typeof setTimeout> | null,

            voiceReaderSample: null as SpeechSynthesisUtterance | null,
            voiceReader: null as SpeechSynthesisUtterance | null,
        };
    },
    data: function () {
        return {
            mid: AppStatus.CurrentMedia,

            editing: false,
            configuring: false,

            content: "",
            contentToChange: "",

            contentLines: [] as string[],

            contentStoredId: -1,
            contentStored: "",

            reading: false,
            readingLine: -1,

            loading: true,

            busy: false,
            busyDisplayLoad: false,

            canWrite: AuthController.CanWrite,

            baseFontSize: getDescriptionSize(),
            voiceReadingSettings: clone(getDescriptionWidgetReadSettings()),

            sampleVoiceText: this.$t("This is a sample text in order to test the voice reading settings"),
            playingSample: false,
        };
    },
    computed: {
        actionButtons: function (): ActionButton[] {
            const buttons: ActionButton[] = [];

            if (this.loading) {
                return [];
            }

            if (!this.editing && !this.configuring) {
                if (this.reading) {
                    buttons.push({
                        id: "read-pause",
                        name: this.$t("Pause reading"),
                        icon: "fas fa-pause",
                        key: "-",
                    });
                } else if (this.speechSynthesisAvailable && this.contentLines.length > 0 && this.voiceReadingSettings.enabled) {
                    buttons.push({
                        id: "read-play",
                        name: this.$t("Start reading"),
                        icon: "fas fa-play",
                        key: "-",
                    });
                }

                buttons.push({
                    id: "size-minus",
                    name: this.$t("Smaller font size"),
                    icon: "fas fa-magnifying-glass-minus",
                    key: "-",
                });

                buttons.push({
                    id: "size-plus",
                    name: this.$t("Bigger font size"),
                    icon: "fas fa-magnifying-glass-plus",
                    key: "+",
                });
            }

            if (!this.editing) {
                if (this.configuring) {
                    buttons.push({
                        id: "config-done",
                        name: this.$t("Done"),
                        icon: "fas fa-check",
                        key: "-",
                    });

                    buttons.push({
                        id: "config-reset",
                        name: this.$t("Reset to default values"),
                        icon: "fas fa-broom",
                        key: "-",
                    });
                } else {
                    buttons.push({
                        id: "config",
                        name: this.$t("Configuration"),
                        icon: "fas fa-cog",
                        key: "-",
                    });
                }
            }

            if (this.canWrite && !this.configuring) {
                if (this.editing) {
                    buttons.push({
                        id: "save",
                        name: this.$t("Save changes"),
                        icon: this.busyDisplayLoad ? "fa fa-spinner fa-spin" : "fas fa-check",
                    });
                } else {
                    buttons.push({
                        id: "edit",
                        name: this.$t("Edit"),
                        icon: "fas fa-pencil-alt",
                    });
                }
            }

            return buttons;
        },
    },
    watch: {
        display: function () {
            if (this.display) {
                this.contentStoredId = -1;
                this.contentStored = "";
                this.load();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, this.updateDescription.bind(this));

        if (this.display) {
            this.load();
        }
    },
    beforeUnmount: function () {
        if (this.busyTimeout) {
            clearTimeout(this.busyTimeout);
            this.busyTimeout = null;
        }

        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);

        if (this.voiceReaderSample) {
            speechSynthesis.cancel();
            this.voiceReaderSample = null;
        }

        if (this.voiceReader) {
            speechSynthesis.cancel();
            this.voiceReader = null;
        }
    },
    methods: {
        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            if (!MediaController.MediaData) {
                return;
            }

            const descFilePath = MediaController.MediaData.description_url;

            if (!descFilePath) {
                this.content = "";
                this.contentToChange = "";
                this.contentLines = [];
                this.readingLine = -1;
                this.loading = false;
                this.editing = !!this.canWrite;
                this.stopReading();

                if (this.contentStoredId === MediaController.MediaData.id) {
                    this.contentToChange = this.contentStored;
                } else {
                    this.contentStoredId = -1;
                    this.contentStored = "";
                }

                this.autoFocus();
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.loadRequestId, {
                method: "GET",
                url: getAssetURL(descFilePath),
            })
                .onSuccess((descriptionText) => {
                    this.content = descriptionText;
                    this.contentLines = this.content
                        .split("\n\n")
                        .map((l) => l.trim())
                        .filter((l) => l.length > 0);
                    this.readingLine = -1;
                    this.contentToChange = descriptionText;
                    this.loading = false;
                    this.editing = this.canWrite && !this.content;
                    this.stopReading();

                    if (this.contentStoredId === MediaController.MediaData.id) {
                        this.contentToChange = this.contentStored;
                        this.editing = !!this.canWrite;
                    } else {
                        this.contentStoredId = -1;
                        this.contentStored = "";
                    }

                    this.autoFocus();
                })
                .onRequestError((err) => {
                    new RequestErrorHandler()
                        .add(401, "*", () => {
                            emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(404, "*", () => {
                            this.content = "";
                            this.contentToChange = "";
                            this.contentLines = [];
                            this.readingLine = -1;
                            this.loading = false;
                            this.editing = !!this.canWrite;
                            this.stopReading();

                            if (this.contentStoredId === MediaController.MediaData.id) {
                                this.contentToChange = this.contentStored;
                            } else {
                                this.contentStoredId = -1;
                                this.contentStored = "";
                            }
                            this.autoFocus();
                        })
                        .add("*", "*", () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        clickActionButton: function (id: string) {
            switch (id) {
                case "save":
                    this.saveChanges();
                    break;
                case "edit":
                    this.stopReading();
                    this.startEdit();
                    break;
                case "config":
                    this.stopReading();
                    this.configuring = true;
                    break;
                case "config-done":
                    this.configuring = false;
                    this.stopPlayingSample();
                    break;
                case "config-reset":
                    this.resetDefaultValues();
                    break;
                case "size-plus":
                    this.baseFontSize = Math.min(128, this.baseFontSize + 1);
                    this.saveBaseFontSize();
                    break;
                case "size-minus":
                    this.baseFontSize = Math.max(1, this.baseFontSize - 1);
                    this.saveBaseFontSize();
                    break;
                case "read-play":
                    this.startReading();
                    break;
                case "read-pause":
                    this.pauseReading();
                    break;
            }
        },

        saveBaseFontSize: function () {
            setDescriptionSize(this.baseFontSize);
        },

        propagateClick: function () {
            this.$emit("clicked");
        },

        close: function () {
            this.displayStatus = false;
        },

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                let elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                } else {
                    elem = this.$el.querySelector(".media-description-body");
                    if (elem) {
                        elem.focus();
                    }
                }
            });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
            if (!this.canWrite) {
                this.cancelEdit();
            }
        },

        updateMediaData: function () {
            this.mid = AppStatus.CurrentMedia;
            this.load();
        },

        updateDescription: function (source: string) {
            if (source === "widget") {
                return;
            }

            this.updateMediaData();
        },

        startEdit: function () {
            if (this.editing) {
                return;
            }
            this.editing = true;
            this.autoFocus();
        },

        cancelEdit: function () {
            if (!this.editing) {
                return;
            }
            this.contentToChange = this.content;
            this.editing = false;
            this.autoFocus();
        },

        renderContent: function (text: string): string {
            return replaceLinks(escapeHTML(text)).replace(/\n/g, "<br>");
        },

        saveChanges: function () {
            if (!this.editing) {
                return;
            }

            if (this.busy) {
                return;
            }

            this.clearBusyTimeout();

            if (this.content === this.contentToChange) {
                this.editing = false;
                this.autoFocus();
                return;
            }

            this.busy = true;

            this.setBusyTimeout();

            const mid = this.mid;

            makeApiRequest(apiMediaSetDescription(mid, this.contentToChange))
                .onSuccess((res) => {
                    this.busy = false;
                    this.clearBusyTimeout();

                    PagesController.ShowSnackBar(this.$t("Successfully saved description"));
                    this.content = this.contentToChange;
                    this.contentLines = this.content
                        .split("\n\n")
                        .map((l) => l.trim())
                        .filter((l) => l.length > 0);
                    this.readingLine = -1;
                    this.editing = false;
                    this.stopReading();

                    if (MediaController.MediaData && MediaController.MediaData.id === mid) {
                        MediaController.MediaData.description_url = res.url || "";
                    }

                    emitAppEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, "widget");

                    this.$emit("update-desc");

                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    this.clearBusyTimeout();
                    handleErr(err, {
                        unauthorized: () => {
                            this.contentStoredId = this.mid;
                            this.contentStored = this.contentToChange;
                            emitAppEvent(EVENT_NAME_UNAUTHORIZED);
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
                    this.busy = false;
                    this.clearBusyTimeout();
                    console.error(err);
                });
        },

        setBusyTimeout: function () {
            this.busyTimeout = setTimeout(() => {
                this.busyDisplayLoad = true;
                this.busyTimeout = null;
            }, 333);
        },

        clearBusyTimeout: function () {
            if (this.busyTimeout) {
                clearTimeout(this.busyTimeout);
                this.busyTimeout = null;
            }

            this.busyDisplayLoad = false;
        },

        resetDefaultValues: function () {
            this.baseFontSize = 18;
            this.onFontSizeChange();
            this.voiceReadingSettings.enabled = false;
            this.voiceReadingSettings.volume = 1;
            this.voiceReadingSettings.pitch = 1;
            this.voiceReadingSettings.rate = 1;
            this.voiceReadingSettings.voice = "";
        },

        onFontSizeChange: function () {
            setDescriptionSize(this.baseFontSize);
        },

        onChangedVoiceReadingSettings: function () {
            setDescriptionWidgetReadSettings(this.voiceReadingSettings);
            this.stopPlayingSample();
        },

        stopPlayingSample: function () {
            if (this.voiceReaderSample) {
                speechSynthesis.cancel();
                this.voiceReaderSample = null;
            }

            this.playingSample = false;
        },

        playSample: function () {
            if (this.voiceReaderSample) {
                speechSynthesis.cancel();
                this.voiceReaderSample = null;
            }

            const voice = getVoiceToSpeak(this.$locale.value, this.voiceReadingSettings.voice);

            if (!voice) {
                return;
            }

            const reader = new SpeechSynthesisUtterance(this.sampleVoiceText);

            reader.addEventListener("error", (ev) => {
                switch (ev.error) {
                    case "canceled":
                    case "interrupted":
                        break;
                    default:
                        console.error(ev.error);
                }

                if (this.voiceReaderSample !== reader) {
                    return;
                }

                this.voiceReaderSample = null;
                this.playingSample = false;
            });

            reader.addEventListener("end", () => {
                if (this.voiceReaderSample !== reader) {
                    return;
                }

                this.voiceReaderSample = null;
                this.playingSample = false;
            });

            reader.voice = voice;
            reader.volume = this.voiceReadingSettings.volume;
            reader.pitch = this.voiceReadingSettings.pitch;
            reader.rate = this.voiceReadingSettings.rate;

            this.voiceReaderSample = reader;
            this.playingSample = true;

            speechSynthesis.speak(reader);
        },

        selectCurrentLine: function () {
            const widgetBody = this.$el.querySelector(".resizable-widget-body") as HTMLElement;

            if (!widgetBody) {
                if (this.readingLine < 0) {
                    this.readingLine = 0;
                }
                return;
            }

            const widgetBodyRect = widgetBody.getBoundingClientRect();

            if (this.readingLine >= 0) {
                const currentElement = document.getElementById(this.paragraphIdPrefix + "-" + this.readingLine) as HTMLElement;

                if (currentElement) {
                    const currentElementRect = currentElement.getBoundingClientRect();

                    if (elementIsVisibleVertical(currentElementRect, widgetBodyRect)) {
                        return;
                    }
                }
            }

            const elements = this.$el.querySelectorAll(".media-description-paragraph") as HTMLElement[];

            for (let i = 0; i < elements.length; i++) {
                const elRect = elements[i].getBoundingClientRect();

                if (elementIsVisibleVertical(elRect, widgetBodyRect)) {
                    this.readingLine = Math.min(i, this.contentLines.length - 1);
                    return;
                }
            }

            this.readingLine = 0;
        },

        readNextLine: function () {
            if (this.readingLine >= this.contentLines.length - 1) {
                // End
                this.readingLine = -1;
                this.reading = false;
                return;
            }

            this.readingLine++;
            this.setupVoiceReader();
        },

        readSkipToLine: function (i: number) {
            if (!this.reading) {
                return;
            }

            this.readingLine = i;
            this.setupVoiceReader();
        },

        setupVoiceReader: function () {
            if (this.voiceReader) {
                speechSynthesis.cancel();
                this.voiceReader = null;
            }

            const voice = getVoiceToSpeak(this.$locale.value, this.voiceReadingSettings.voice);

            if (!voice) {
                this.reading = false;
                return;
            }

            let text = this.contentLines[this.readingLine] || "";

            while (!text) {
                if (this.readingLine >= this.contentLines.length - 1) {
                    // End
                    this.readingLine = -1;
                    this.reading = false;
                    return;
                }

                this.readingLine++;
                text = this.contentLines[this.readingLine] || "";
            }

            const reader = new SpeechSynthesisUtterance(text);

            reader.addEventListener("error", (ev) => {
                switch (ev.error) {
                    case "canceled":
                    case "interrupted":
                        return;
                    default:
                        console.error(ev.error);
                }

                if (this.voiceReader !== reader) {
                    return;
                }

                this.voiceReader = null;
                this.reading = false;
            });

            reader.addEventListener("end", () => {
                if (this.voiceReader !== reader) {
                    return;
                }

                this.voiceReader = null;
                this.readNextLine();
            });

            reader.voice = voice;
            reader.volume = this.voiceReadingSettings.volume;
            reader.pitch = this.voiceReadingSettings.pitch;
            reader.rate = this.voiceReadingSettings.rate;

            this.voiceReader = reader;

            speechSynthesis.speak(reader);

            this.scrollIntoReadingElement();
        },

        scrollIntoReadingElement: function () {
            const el = document.getElementById(this.paragraphIdPrefix + "-" + this.readingLine) as HTMLElement;

            if (!el) {
                return;
            }

            el.scrollIntoView();
        },

        startReading: function () {
            if (this.voiceReader) {
                const oldReadingLine = this.readingLine;
                this.selectCurrentLine();

                if (oldReadingLine === this.readingLine) {
                    speechSynthesis.resume();
                    this.reading = true;
                    return;
                }
            }

            if (this.contentLines.length === 0) {
                return;
            }

            this.selectCurrentLine();

            this.setupVoiceReader();

            this.reading = true;
        },

        pauseReading: function () {
            if (this.voiceReader) {
                speechSynthesis.pause();
                this.reading = false;
                return;
            }

            this.stopReading();
        },

        stopReading: function () {
            if (this.voiceReader) {
                speechSynthesis.cancel();
                this.voiceReader = null;
            }
            this.reading = false;
        },
    },
});
</script>
