// Voice Synthesis utils

"use strict";

/**
 * Checks if the SpeechSynthesis API is available
 * @returns True if available
 */
export function isSpeechSynthesisAvailable(): boolean {
    return !!window.speechSynthesis;
}

/**
 * Gets the best voice based on the user settings
 * @param locale The user locale
 * @param preferredVoice The user preferred voice (manually selected)
 * @returns The voice object, or null
 */
export function getVoiceToSpeak(locale: string, preferredVoice: string): SpeechSynthesisVoice | null {
    if (!isSpeechSynthesisAvailable()) {
        return null;
    }

    const voices = speechSynthesis.getVoices();

    for (const voice of voices) {
        if (preferredVoice && voice.voiceURI === preferredVoice) {
            return voice;
        }
    }

    const localeBase = (locale + "").split("-")[0];
    const localeVoice = voices.find((voice) => {
        const voiceLocaleBase = voice.lang.split("-")[0];

        return localeBase === voiceLocaleBase;
    });

    if (localeVoice) {
        return localeVoice;
    }

    const defaultVoice = voices.find((voice) => voice.default);

    if (defaultVoice) {
        return defaultVoice;
    }

    return voices[0] || null;
}
