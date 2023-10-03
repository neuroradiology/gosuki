import './styles/style.css';

import Vue from 'vue';
import App from './App.vue';


window.app = new Vue({
    data: {
    // declare message with an empty value
    message: ''
  },
    render: createElement => createElement(App)
})
window.app.$mount("#app");

