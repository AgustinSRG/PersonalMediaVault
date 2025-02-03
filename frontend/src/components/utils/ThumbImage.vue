<template>
    <div :class="'thumb' + (className ? ' ' + className : '')">
        <img :src="src" :alt="$t('Thumbnail')" loading="lazy" @load="onLoadingFalse" @error="onLoadingFalse" />
        <div class="thumb-load" v-if="displayLoader">
            <i class="fa fa-spinner fa-spin"></i>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

const DEFAULT_DELAY = 333;

export default defineComponent({
    name: "ThumbImage",
    props: {
        src: String,
        className: String,
    },
    setup: function () {
        return {
            displayTimeout: null,
        };
    },
    data: function () {
        return {
            displayLoader: false,
        };
    },
    methods: {
        onLoadingTrue: function () {
            if (this.displayTimeout) {
                clearTimeout(this.displayTimeout);
                this.displayTimeout = null;
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

            this.displayLoader = false;
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
    },
    watch: {
        src: function () {
            this.onLoadingTrue();
        },
    },
});
</script>
