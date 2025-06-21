<template>
    <div class="six-digit-code-input" :class="{ min: min }">
        <div v-for="(c, i) in code" :key="i" class="six-digit-code-char-input" :class="{ first: i === 0 }">
            <input
                v-model="c.c"
                type="text"
                :disabled="disabled"
                :class="'form-control auto-focus code-char-' + i"
                maxlength="1"
                @input="goNextChar(c, i)"
                @paste="onPaste($event, i)"
                @keydown="onKeyDown($event, c, i)"
            />
        </div>
    </div>
</template>

<script lang="ts">
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

function updateCode(code: { c: string }[], val: string) {
    for (let i = 0; i < code.length; i++) {
        code[i].c = val.charAt(i) || "";
    }
}

export default defineComponent({
    name: "SixDigitCodeInput",
    props: {
        min: Boolean,
        val: String,
        disabled: Boolean,
    },
    emits: ["update:val"],
    setup: function (props) {
        return {
            valState: useVModel(props, "val"),
        };
    },
    data: function () {
        const code: { c: string }[] = [{ c: "" }, { c: "" }, { c: "" }, { c: "" }, { c: "" }, { c: "" }];

        updateCode(code, this.val || "");

        return {
            code,
        };
    },
    watch: {
        val: function () {
            updateCode(this.code, this.val || "");
        },
    },
    methods: {
        onCodeUpdated: function () {
            this.valState = this.code.map((c) => c.c).join("");
        },

        onKeyDown: function (event: KeyboardEvent, c: { c: string }, i: number) {
            if (event.key === "Backspace" && !c.c) {
                event.preventDefault();
                this.goNextChar(c, i);
            }
        },

        goNextChar: function (c: { c: string }, i: number) {
            c.c = c.c.charAt(0).toUpperCase();

            this.onCodeUpdated();

            if (!c.c) {
                // Go back
                if (i > 0) {
                    const nextInput = this.$el.querySelector(".code-char-" + (i - 1));

                    if (nextInput) {
                        nextInput.focus();
                        if (nextInput.select) {
                            nextInput.select();
                        }
                    }
                }

                return;
            }

            if (i < this.code.length - 1) {
                const nextInput = this.$el.querySelector(".code-char-" + (i + 1));

                if (nextInput) {
                    nextInput.focus();
                    if (nextInput.select) {
                        nextInput.select();
                    }
                }
            }
        },

        onPaste: function (ev: ClipboardEvent, i: number) {
            ev.preventDefault();

            const text = ev.clipboardData.getData("text/plain").replace(/[^a-z0-9]+/gi, "");

            let k = 0;
            for (let j = i; j < this.code.length; j++) {
                if (k >= text.length) {
                    break;
                }

                const c = text.charAt(k).toUpperCase();
                k++;

                this.code[j].c = c;
                this.goNextChar(this.code[j], j);
            }

            this.onCodeUpdated();
        },
    },
});
</script>
