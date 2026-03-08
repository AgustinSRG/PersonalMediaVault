<template>
    <div ref="container" class="resizable-widget-container">
        <ResizableWidget
            v-model:display="display"
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
                                        @input="saveBaseFontSize"
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
                                        @input="saveBaseFontSize"
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
                                    <VoiceSelect
                                        v-model:voice="voiceReadingSettings.voice"
                                        @update:voice="onChangedVoiceReadingSettings"
                                    ></VoiceSelect>
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

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, shallowRef, useTemplateRef, watch } from "vue";
import ResizableWidget from "@/components/widgets/common/ResizableWidget.vue";
import { nextTick } from "vue";
import { AppStatus } from "@/control/app-status";
import { emitAppEvent, EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, EVENT_NAME_MEDIA_UPDATE, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
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
import VoiceSelect from "@/components/utils/VoiceSelect.vue";
import {
    getDescriptionSize,
    getDescriptionWidgetReadSettings,
    setDescriptionSize,
    setDescriptionWidgetReadSettings,
} from "@/control/player-preferences";
import { isSpeechSynthesisAvailable } from "@/utils/voice-synthesis";
import { clone } from "@/utils/objects";
import type { WidgetActionButton } from "@/utils/widgets";
import { elementIsVisibleVertical } from "@/utils/dom";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useRequestId } from "@/composables/use-request-id";
import { useTimeout } from "@/composables/use-timeout";
import { useSpeechReader } from "@/composables/use-speech-reader";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { getStoredDescription, setStoredDescription } from "@/control/description-store";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Display model
const display = defineModel<boolean>("display");

// Props
defineProps({
    /**
     * A context menu is opened
     */
    contextOpen: Boolean,

    /**
     * The title of the media
     */
    title: String,
});

// Emits
const emit = defineEmits<{
    /**
     * The user clicked the widget
     */
    (e: "clicked"): void;

    /**
     * The description was updated
     */
    (e: "update-desc"): void;
}>();

// Unique prefix to identify paragraph elements
const paragraphIdPrefix = getUniqueStringId();

// Description content
const content = ref("");

// Description content to be changed by the user
const contentToChange = ref("");

// Lines of content, to separate by paragraphs
const contentLines = ref<string[]>([]);

// Busy status
const busy = ref(false);

// True to display the busy loader
const busyDisplayLoad = ref(false);

// Current media ID
const mid = ref(AppStatus.CurrentMedia);

// Loading status
const loading = ref(true);

// Editing?
const editing = ref(false);

// Configuring?
const configuring = ref(false);

// Timeout to display the busy status
const busyTimeout = useTimeout();

// Speech synthesis available?
const speechSynthesisAvailable = isSpeechSynthesisAvailable();

// List of voices for speech synthesis
const voices = shallowRef<SpeechSynthesisVoice[]>([]);

// Voice reading settings
const voiceReadingSettings = ref(clone(getDescriptionWidgetReadSettings()));

// Speech reader
const speechReader = useSpeechReader();

// Reading with speech reader?
const reading = ref(false);

// Index of the line being read
const readingLine = ref(-1);

// Sample text to play as a voice sample
const sampleVoiceText = computed(() => $t("This is a sample text in order to test the voice reading settings"));

// True if voice sample is being played
const playingSample = ref(false);

// Font size for the description
const baseFontSize = ref(getDescriptionSize());

// List of action buttons for the widget
const actionButtons = computed<WidgetActionButton[]>(() => {
    const buttons: WidgetActionButton[] = [];

    if (loading.value) {
        return [];
    }

    if (!editing.value && !configuring.value) {
        if (reading.value) {
            buttons.push({
                id: "read-pause",
                name: $t("Pause reading"),
                icon: "fas fa-pause",
            });
        } else if (speechSynthesisAvailable && contentLines.value.length > 0 && voiceReadingSettings.value.enabled) {
            buttons.push({
                id: "read-play",
                name: $t("Start reading"),
                icon: "fas fa-play",
            });
        }

        buttons.push({
            id: "size-minus",
            name: $t("Smaller font size"),
            icon: "fas fa-magnifying-glass-minus",
            key: "-",
        });

        buttons.push({
            id: "size-plus",
            name: $t("Bigger font size"),
            icon: "fas fa-magnifying-glass-plus",
            key: "+",
        });
    }

    if (!editing.value) {
        if (configuring.value) {
            buttons.push({
                id: "config-done",
                name: $t("Done"),
                icon: "fas fa-check",
            });

            buttons.push({
                id: "config-reset",
                name: $t("Reset to default values"),
                icon: "fas fa-broom",
            });
        } else {
            buttons.push({
                id: "config",
                name: $t("Configuration"),
                icon: "fas fa-cog",
            });
        }
    }

    if (canWrite.value && !configuring.value) {
        if (editing.value) {
            buttons.push({
                id: "save",
                name: $t("Save changes"),
                icon: busyDisplayLoad.value ? "fa fa-spinner fa-spin" : "fas fa-check",
            });
        } else {
            buttons.push({
                id: "edit",
                name: $t("Edit"),
                icon: "fas fa-pencil-alt",
            });
        }
    }

    return buttons;
});

// Load request ID
const loadRequestId = useRequestId();

// Delay to retry loading after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the content
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (!display.value) {
        return;
    }

    if (!MediaController.MediaData) {
        return;
    }

    const descFilePath = MediaController.MediaData.description_url;

    if (!descFilePath) {
        content.value = "";
        contentToChange.value = "";
        contentLines.value = [];
        readingLine.value = -1;
        loading.value = false;
        editing.value = !!canWrite.value;

        stopReading();

        const [contentStoredId, contentStored] = getStoredDescription();

        if (contentStoredId === MediaController.MediaData.id) {
            contentToChange.value = contentStored;
        }

        setStoredDescription(-1, "");

        autoFocus();
        return;
    }

    loading.value = true;

    makeNamedApiRequest(loadRequestId, {
        method: "GET",
        url: getAssetURL(descFilePath),
    })
        .onSuccess((descriptionText) => {
            content.value = descriptionText;
            contentLines.value = content.value
                .split("\n")
                .map((l) => l.trim())
                .filter((l) => l.length > 0);

            readingLine.value = -1;

            contentToChange.value = descriptionText;

            loading.value = false;

            editing.value = canWrite.value && !content.value;

            stopReading();

            const [contentStoredId, contentStored] = getStoredDescription();

            if (contentStoredId === MediaController.MediaData.id) {
                contentToChange.value = contentStored;
                editing.value = !!canWrite.value;
            }

            setStoredDescription(-1, "");

            autoFocus();
        })
        .onRequestError((err) => {
            new RequestErrorHandler()
                .add(401, "*", () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                })
                .add(404, "*", () => {
                    content.value = "";
                    contentToChange.value = "";
                    contentLines.value = [];

                    readingLine.value = -1;

                    loading.value = false;

                    editing.value = !!canWrite.value;

                    stopReading();

                    const [contentStoredId, contentStored] = getStoredDescription();

                    if (contentStoredId === MediaController.MediaData.id) {
                        contentToChange.value = contentStored;
                    }

                    setStoredDescription(-1, "");

                    autoFocus();
                })
                .add("*", "*", () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                })
                .handle(err);
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

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, () => {
    mid.value = AppStatus.CurrentMedia;
    load();
});

watch(display, () => {
    if (display.value) {
        setStoredDescription(-1, "");
        load();
    }
});

onApplicationEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, (source) => {
    if (source === "widget") {
        return;
    }

    load();
});

/**
 * Handler for action buttons
 * @param id The button ID
 */
const clickActionButton = (id: string) => {
    switch (id) {
        case "save":
            saveChanges();
            break;
        case "edit":
            stopReading();
            startEdit();
            break;
        case "config":
            stopReading();
            configuring.value = true;
            if (speechSynthesisAvailable) {
                voices.value = speechSynthesis.getVoices();
            }
            break;
        case "config-done":
            configuring.value = false;
            stopPlayingSample();
            break;
        case "config-reset":
            resetDefaultValues();
            break;
        case "size-plus":
            baseFontSize.value = Math.min(128, baseFontSize.value + 1);
            saveBaseFontSize();
            break;
        case "size-minus":
            baseFontSize.value = Math.max(1, baseFontSize.value - 1);
            saveBaseFontSize();
            break;
        case "read-play":
            startReading();
            break;
        case "read-pause":
            pauseReading();
            break;
    }
};

/**
 * Start editing the description
 */
const startEdit = () => {
    if (editing.value) {
        return;
    }
    editing.value = true;
    autoFocus();
};

/**
 * Stop editing and discards the changes
 */
const cancelEdit = () => {
    if (!editing.value) {
        return;
    }
    contentToChange.value = content.value;
    editing.value = false;
    autoFocus();
};

watch(canWrite, () => {
    if (!canWrite.value) {
        cancelEdit();
    }
});

/**
 * Sets the base font size to the settings
 */
const saveBaseFontSize = () => {
    setDescriptionSize(baseFontSize.value);
};

/**
 * Resets default settings
 */
const resetDefaultValues = () => {
    baseFontSize.value = 18;

    saveBaseFontSize();

    voiceReadingSettings.value.enabled = false;
    voiceReadingSettings.value.volume = 1;
    voiceReadingSettings.value.pitch = 1;
    voiceReadingSettings.value.rate = 1;
    voiceReadingSettings.value.voice = "";
};

// Delay to display the busy loader (milliseconds)
const BUSY_TIMEOUT_DELAY = 333;

/**
 * Sets the timeout to display the busy loader
 */
const setBusyTimeout = () => {
    busyTimeout.set(() => {
        busyDisplayLoad.value = true;
    }, BUSY_TIMEOUT_DELAY);
};

/**
 * Clears busy timeout
 */
const clearBusyTimeout = () => {
    busyTimeout.clear();

    busyDisplayLoad.value = false;
};

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

/**
 * Saves changes to the description
 */
const saveChanges = () => {
    if (!editing.value) {
        return;
    }

    if (busy.value) {
        return;
    }

    clearBusyTimeout();

    if (content.value === contentToChange.value) {
        editing.value = false;
        autoFocus();
        return;
    }

    busy.value = true;

    setBusyTimeout();

    const mediaId = mid.value;

    makeApiRequest(apiMediaSetDescription(mediaId, contentToChange.value))
        .onSuccess((res) => {
            busy.value = false;

            clearBusyTimeout();

            PagesController.ShowSnackBar($t("Successfully saved description"));

            content.value = contentToChange.value;
            contentLines.value = content.value
                .split("\n")
                .map((l) => l.trim())
                .filter((l) => l.length > 0);

            readingLine.value = -1;
            editing.value = false;

            stopReading();

            if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                MediaController.MediaData.description_url = res.url || "";
            }

            emitAppEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, "widget");

            emit("update-desc");

            autoFocus();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            clearBusyTimeout();

            handleErr(err, {
                unauthorized: () => {
                    setStoredDescription(mid.value, contentToChange.value);
                    unauthorized();
                },
                badRequest,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            clearBusyTimeout();
            console.error(err);
            setError(err.message);
        });
};

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    if (!display.value) {
        return;
    }

    nextTick(() => {
        let elem = container.value?.querySelector(".auto-focus") as HTMLElement;
        if (elem) {
            elem.focus();
        } else {
            elem = container.value?.querySelector(".media-description-body") as HTMLElement;
            if (elem) {
                elem.focus();
            }
        }
    });
};

/**
 * Renders content in sanitized HTML
 * @param text The text
 * @returns The sanitized HTML
 */
const renderContent = (text: string): string => {
    return replaceLinks(escapeHTML(text)).replace(/\n/g, "<br>");
};

/**
 * Propagates click event to the parent element
 */
const propagateClick = () => {
    emit("clicked");
};

/**
 * Plays the text sample
 */
const playSample = () => {
    speechReader.stop();
    reading.value = false;
    playingSample.value = false;

    const success = speechReader.read(
        sampleVoiceText.value,
        voiceReadingSettings.value,
        () => {
            playingSample.value = false;
        },
        (ev: SpeechSynthesisErrorEvent) => {
            switch (ev.error) {
                case "canceled":
                case "interrupted":
                    return;
                default:
                    console.error(ev.error);
            }

            playingSample.value = false;
        },
    );

    if (success) {
        playingSample.value = true;
    }
};

/**
 * Stops playing sample audio
 */
const stopPlayingSample = () => {
    speechReader.stop();
    reading.value = false;
    playingSample.value = false;
};

/**
 * Called when the voice reading settings are changed
 */
const onChangedVoiceReadingSettings = () => {
    setDescriptionWidgetReadSettings(voiceReadingSettings.value);
    stopPlayingSample();
};

/**
 * Setups voice reader
 */
const setupVoiceReader = () => {
    speechReader.stop();
    reading.value = false;
    playingSample.value = false;

    let text = contentLines.value[readingLine.value] || "";

    while (!text) {
        if (readingLine.value >= contentLines.value.length - 1) {
            // End
            readingLine.value = -1;
            reading.value = false;
            return;
        }

        readingLine.value++;

        text = contentLines.value[readingLine.value] || "";
    }

    const success = speechReader.read(text, voiceReadingSettings.value, readNextLine, (ev: SpeechSynthesisErrorEvent) => {
        switch (ev.error) {
            case "canceled":
            case "interrupted":
                return;
            default:
                console.error(ev.error);
        }

        reading.value = false;
    });

    if (success) {
        reading.value = true;
        scrollIntoReadingElement();
    }
};

/**
 * Reads the next line
 */
const readNextLine = () => {
    if (readingLine.value >= contentLines.value.length - 1) {
        // End
        readingLine.value = -1;
        reading.value = false;
        return;
    }

    readingLine.value++;

    setupVoiceReader();
};

/**
 * Starts reading from a specific line
 * @param i The line index
 */
const readSkipToLine = (i: number) => {
    if (!reading.value) {
        return;
    }

    readingLine.value = i;

    setupVoiceReader();
};

/**
 * Scrolls into the current line being read
 */
const scrollIntoReadingElement = () => {
    const el = document.getElementById(paragraphIdPrefix + "-" + readingLine.value) as HTMLElement;

    if (!el) {
        return;
    }

    el.scrollIntoView();
};

/**
 * Selects a line to read based on the current scroll
 */
const selectCurrentLine = () => {
    const widgetBody = container.value?.querySelector(".resizable-widget-body") as HTMLElement;

    if (!widgetBody) {
        if (readingLine.value < 0) {
            readingLine.value = 0;
        }
        return;
    }

    const widgetBodyRect = widgetBody.getBoundingClientRect();

    if (readingLine.value >= 0) {
        const currentElement = document.getElementById(paragraphIdPrefix + "-" + readingLine.value) as HTMLElement;

        if (currentElement) {
            const currentElementRect = currentElement.getBoundingClientRect();

            if (elementIsVisibleVertical(currentElementRect, widgetBodyRect)) {
                return;
            }
        }
    }

    const elements = (container.value?.querySelectorAll(".media-description-paragraph") || []) as HTMLElement[];

    for (let i = 0; i < elements.length; i++) {
        const elRect = elements[i].getBoundingClientRect();

        if (elementIsVisibleVertical(elRect, widgetBodyRect)) {
            readingLine.value = Math.min(i, contentLines.value.length - 1);
            return;
        }
    }

    readingLine.value = 0;
};

/**
 * Starts reading
 */
const startReading = () => {
    if (speechReader.isPaused()) {
        const oldReadingLine = readingLine.value;
        selectCurrentLine();

        if (oldReadingLine === readingLine.value) {
            speechReader.resume();
            reading.value = true;
            return;
        }
    }

    if (contentLines.value.length === 0) {
        return;
    }

    selectCurrentLine();

    setupVoiceReader();

    reading.value = true;
};

/**
 * Pauses the reader
 */
const pauseReading = () => {
    const paused = speechReader.pause();

    if (!paused) {
        stopReading();
        return;
    }

    reading.value = false;
};

/**
 * Stops the reader
 */
const stopReading = () => {
    speechReader.stop();

    reading.value = false;
};
</script>
