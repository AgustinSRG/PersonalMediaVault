<template>
    <div ref="container" class="six-digit-code-input" :class="{ min: min }">
        <div v-for="(c, i) in code" :key="i" class="six-digit-code-char-input" :class="{ first: i === 0 }">
            <input
                v-model="c.c"
                type="text"
                :disabled="disabled"
                :class="'form-control auto-focus code-char-' + i"
                maxlength="1"
                @input="goNextDigit(c, i)"
                @paste="onPaste($event, i)"
                @keydown="onKeyDown($event, c, i)"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { reactive, useTemplateRef, watch } from "vue";

// Size of the code
const CODE_SIZE = 6;

/**
 * Code digit
 */
type CodeDigit = {
    /**
     * The digit
     */
    c: string;
};

/**
 * Code digit array type
 */
type CodeDigitArray = CodeDigit[];

/**
 * Updates the code array
 * @param code The code array
 * @param val The new value
 */
function updateCode(code: CodeDigitArray, val: string) {
    for (let i = 0; i < code.length; i++) {
        code[i].c = val.charAt(i) || "";
    }
}

/**
 * Creates code digit array from initial value
 * @param val The value
 * @returns The code digit array
 */
function makeCode(val: string): CodeDigitArray {
    const code: CodeDigitArray = [];

    for (let i = 0; i < CODE_SIZE; i++) {
        code.push({ c: " " });
    }

    updateCode(code, val);

    return code;
}

defineProps({
    /**
     * True to use miniature style
     */
    min: Boolean,

    /**
     * True if the input is disabled
     */
    disabled: Boolean,
});

// Value model
const val = defineModel<string>("val");

// The reactive code digit array
const code = reactive<CodeDigitArray>(makeCode(val.value || ""));

// Watch for value changes to update the code
watch(val, () => {
    updateCode(code, val.value || "");
});

/**
 * Call when the code digit array is updated by the user
 * to update the value model.
 */
const onCodeUpdated = () => {
    val.value = code.map((c) => c.c || " ").join("");
};

// Ref to the container element
const container = useTemplateRef("container");

/**
 * Focuses the text input corresponding to a character
 * @param i The character index
 */
const focusInput = (i: number) => {
    const inputEl = container.value.querySelector(".code-char-" + i) as HTMLInputElement;

    if (inputEl) {
        inputEl.focus();
        if (inputEl.select) {
            inputEl.select();
        }
    }
};

/**
 * Goes to the next digit based on the current char digit
 * @param c The current digit
 * @param i The current digit index
 */
const goNextDigit = (c: CodeDigit, i: number) => {
    c.c = c.c.charAt(0).toUpperCase();

    onCodeUpdated();

    if (!c.c) {
        // Go back
        if (i > 0) {
            focusInput(i - 1);
        }

        return;
    }

    if (i < code.length - 1) {
        focusInput(i + 1);
    }
};

/**
 * Called on 'keydown' event for the inputs
 * @param event The event
 * @param c The current digit
 * @param i The current digit index
 */
const onKeyDown = (event: KeyboardEvent, c: CodeDigit, i: number) => {
    if (event.ctrlKey) {
        return;
    }

    if (event.key === "Backspace" && !c.c) {
        event.preventDefault();
        goNextDigit(c, i);
    } else if (event.key === "ArrowRight") {
        event.preventDefault();
        if (i < code.length - 1) {
            focusInput(i + 1);
        }
    } else if (event.key === "ArrowLeft") {
        event.preventDefault();
        if (i > 0) {
            focusInput(i - 1);
        }
    } else if (event.key.length === 1) {
        event.preventDefault();
        c.c = event.key.toUpperCase();
        goNextDigit(c, i);
    }
};

/**
 * Called when a code is pasted
 * @param ev The clipboard event
 * @param i The index of the digit input where the code was pasted
 */
const onPaste = (ev: ClipboardEvent, i: number) => {
    ev.preventDefault();

    const text = ev.clipboardData.getData("text/plain").replace(/[^a-z0-9]+/gi, "");

    let k = 0;
    for (let j = i; j < code.length; j++) {
        if (k >= text.length) {
            break;
        }

        const c = text.charAt(k).toUpperCase();
        k++;

        code[j].c = c;
        goNextDigit(code[j], j);
    }

    onCodeUpdated();
};
</script>
