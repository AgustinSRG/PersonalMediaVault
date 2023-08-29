<template>
    <div class="snackbar" :class="{ hidden: !shown }" @mouseenter="hide" @click="hide">{{ message }}</div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { defineComponent } from "vue";

export default defineComponent({
    name: "SnackBar",
    data: function () {
        return {
            shown: false,
            message: "",
        };
    },
    methods: {
        show: function (msg: string) {
            if (this._handles.timeout) {
                clearTimeout(this._handles.timeout);
                this._handles.timeout = null;
            }

            this.shown = true;
            this.message = msg;

            this._handles.timeout = setTimeout(() => {
                this.shown = false;
            }, 3000);
        },

        hide: function () {
            if (this._handles.timeout) {
                clearTimeout(this._handles.timeout);
                this._handles.timeout = null;
            }
            this.shown = false;
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.showH = this.show.bind(this);
        AppEvents.AddEventListener("snack", this._handles.showH);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("snack", this._handles.showH);
        if (this._handles.timeout) {
            clearTimeout(this._handles.timeout);
            this._handles.timeout = null;
        }
    },
});
</script>
