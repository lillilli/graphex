import Vue from "vue";
import ElementUI from "element-ui";
import "element-ui/lib/theme-chalk/index.css";
import App from "./App.vue";
import VueNativeSock from "vue-native-websocket";
import VueChartkick from 'vue-chartkick'
import Chart from 'chart.js'

Vue.use(ElementUI);
Vue.use(VueNativeSock, `ws://${window.location.host}/ws`, {
  format: "json"
});
Vue.use(VueChartkick, {
  adapter: Chart
})

new Vue({
  el: "#app",
  render: h => h(App)
});