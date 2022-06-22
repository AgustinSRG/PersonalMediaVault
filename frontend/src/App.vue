<template>
  <MainLayout></MainLayout>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import MainLayout from "./components/layout/MainLayout.vue";

@Options({
  components: {
    MainLayout,
  },
  data: function () {
    return {};
  },
})
export default class App extends Vue {}
</script>

<style>

/* Add here any global / shared styles */ 

.bold {
  font-weight: bold;
}

.text-right {
  text-align: right;
}

.text-center {
  text-align: center;
}

.text-left {
  text-align: left;
}

.one-line {
  white-space: nowrap;
}


.btn {
  display: inline-block;
  
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: white;
  background: transparent;
  outline: none;
  border-radius: 100vw;
  padding: 12px 24px;
  white-space: nowrap;
}

.btn i {
  margin-right: 12px;
}

.btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.btn:hover {
  opacity: 1;
}

.btn:disabled:hover {
  opacity: 0.7;
}

.btn:focus {
  outline: none;
}

.btn-primary {
  border: solid 1px rgba(255, 255, 255, 0.1);
}

.btn-primary:not(disabled):hover {
  background: rgba(255, 255, 255, 0.1);
}

.form-group {
  padding-bottom: 0.75rem;
}

.form-group label {
  display: block;
  width: 100%;
  padding-bottom: 0.3rem;
}

.form-control {
    outline: none;
    padding: 6px 12px;
    font-size: 14px;
    line-height: 1.42857143;
    color: #fff;
    background: hsl(0, 0%, 7%);
    border: none;
    box-shadow: inset 0 1px 1px rgb(0 0 0 / 8%);
    transition: border-color ease-in-out .15s, box-shadow ease-in-out .15s, -webkit-box-shadow ease-in-out .15s;
}

.form-control-full-width {
  width: 100%;
}

.form-error {
  color: red;
}


</style>
