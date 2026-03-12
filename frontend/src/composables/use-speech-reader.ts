// Speech reader composable

"use strict";

import type { DescriptionWidgetReadSettings } from "@/global-state/player-preferences";
import { getUniqueStringId } from "@/utils/unique-id";
import { onBeforeUnmount, ref, shallowRef } from "vue";
import { useI18n } from "./use-i18n";
import { getVoiceToSpeak, supportsSpeechSynthesisPausing } from "@/utils/voice-synthesis";

/**
 * Speech reader composable
 */
export type SpeechReaderComposable = {
    /**
     * Reads a text
     * @param text The text
     * @param settings The settings
     * @param onEnd Callback called when the reader finishes
     * @param onError Callback called when the reader cannot read due to an error
     * @returns True only if the reader started, false if there is no voice available
     */
    read: (
        text: string,
        settings: DescriptionWidgetReadSettings,
        onEnd: () => void,
        onError: (e: SpeechSynthesisErrorEvent) => void,
    ) => boolean;

    /**
     * Stops reading. Cancels the callback.
     */
    stop: () => void;

    /**
     * Pauses the reader
     * @returns True only if the pause was successful
     */
    pause: () => boolean;

    /**
     * Checks if the reader is paused
     * @returns True if paused
     */
    isPaused: () => boolean;

    /**
     * Resumes the reader after paused
     */
    resume: () => void;
};

/**
 * Creates a speech reader
 * @returns The speech reader
 */
export function useSpeechReader(): SpeechReaderComposable {
    // Supports pausing?
    const supportsPausing = supportsSpeechSynthesisPausing();

    // Actual reader
    const reader = shallowRef<SpeechSynthesisUtterance | null>(null);

    // Unique reader ID
    const readerId = ref("");

    // Paused?
    const paused = ref(false);

    // Event handlers
    let onErrorEventHandler: (ev: SpeechSynthesisErrorEvent) => void;
    let onEndEventHandler: () => void;

    /**
     * Stops the reader
     */
    const stop = () => {
        if (reader.value) {
            reader.value.removeEventListener("error", onErrorEventHandler);
            reader.value.removeEventListener("end", onEndEventHandler);

            try {
                speechSynthesis.cancel();
            } catch (ex) {
                console.error(ex);
            }
        }

        reader.value = null;
        paused.value = false;
    };

    onBeforeUnmount(stop);

    // Locale
    const { locale } = useI18n();

    const read = (
        text: string,
        settings: DescriptionWidgetReadSettings,
        onEnd: () => void,
        onError: (e: SpeechSynthesisErrorEvent) => void,
    ): boolean => {
        stop();

        // Voice
        const voice = getVoiceToSpeak(locale.value, settings.voice);

        if (!voice) {
            return false;
        }

        const id = getUniqueStringId();

        readerId.value = id;

        const newReader = new SpeechSynthesisUtterance(text);

        reader.value = newReader;

        // Add error event handler

        onErrorEventHandler = (ev: SpeechSynthesisErrorEvent) => {
            if (readerId.value !== id) {
                return;
            }

            reader.value = null;
            paused.value = false;

            onError(ev);
        };

        newReader.addEventListener("error", onErrorEventHandler);

        // Add end event handler

        onEndEventHandler = () => {
            if (readerId.value !== id) {
                return;
            }

            reader.value = null;
            paused.value = false;

            onEnd();
        };

        newReader.addEventListener("end", onEndEventHandler);

        // Apply settings to voice synthesis
        newReader.voice = voice;
        newReader.volume = settings.volume;
        newReader.pitch = settings.pitch;
        newReader.rate = settings.rate;

        speechSynthesis.speak(newReader);

        return true;
    };

    /**
     * Checks if the reader is paused
     * @returns True if paused
     */
    const isPaused = (): boolean => {
        return !!reader.value && paused.value;
    };

    /**
     * Pauses the reader
     * @returns True if successful
     */
    const pause = (): boolean => {
        if (!reader.value || paused.value) {
            return false;
        }

        if (!supportsPausing) {
            return false;
        }

        speechSynthesis.pause();
        paused.value = true;

        return true;
    };

    /**
     * Resumes the reader
     */
    const resume = () => {
        if (!reader.value || !paused.value) {
            return;
        }

        speechSynthesis.resume();
        paused.value = false;
    };

    return {
        read,
        stop,
        isPaused,
        pause,
        resume,
    };
}
