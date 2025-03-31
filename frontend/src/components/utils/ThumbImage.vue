<template>
    <div :class="'thumb' + (className ? ' ' + className : '')">
        <img v-if="!isError" :src="src" :alt="$t('Thumbnail')" loading="lazy" @load="onLoadingFalse" @error="onError" />
        <div v-if="displayLoader" class="thumb-load">
            <i class="fa fa-spinner fa-spin"></i>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

const DEFAULT_DELAY = 333;
const RELOAD_DELAY = 1500;

export default defineComponent({
    name: "ThumbImage",
    props: {
        src: String,
        className: String,
    },
    setup: function () {
        return {
            displayTimeout: null,
            reloadTimeout: null,
        };
    },
    data: function () {
        return {
            displayLoader: false,
            isError: false,
        };
    },
    watch: {
        src: function () {
            this.onLoadingTrue();
        },
    },
    mounted: function () {
        const elem = this.$el as HTMLImageElement;
        if (!elem || !elem.complete) {
            this.onLoadingTrue();
        }
    },
    beforeUnmount: function () {
        if (this.displayTimeout) {
            clearTimeout(this.displayTimeout);
            this.displayTimeout = null;
        }

        if (this.reloadTimeout) {
            clearTimeout(this.reloadTimeout);
            this.reloadTimeout = null;
        }
    },
    methods: {
        onLoadingTrue: function () {
            this.isError = false;

            if (this.displayTimeout) {
                clearTimeout(this.displayTimeout);
                this.displayTimeout = null;
            }

            if (this.reloadTimeout) {
                clearTimeout(this.reloadTimeout);
                this.reloadTimeout = null;
            }

            this.displayTimeout = setTimeout(() => {
                this.displayTimeout = null;
                this.displayLoader = true;
            }, DEFAULT_DELAY);
        },

        onLoadingFalse: function () {
            if (this.displayTimeout) {
                clearTimeout(this.displayTimeout);
                this.displayTimeout = null;
            }

            if (this.reloadTimeout) {
                clearTimeout(this.reloadTimeout);
                this.reloadTimeout = null;
            }

            this.displayLoader = false;
        },

        onError: function () {
            this.isError = true;

            if (this.reloadTimeout) {
                clearTimeout(this.reloadTimeout);
                this.reloadTimeout = null;
            }

            this.reloadTimeout = setTimeout(() => {
                this.reloadTimeout = null;
                this.onLoadingTrue();
            }, RELOAD_DELAY);
        },
    },
});
</script>
