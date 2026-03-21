import { createApp } from 'vue';
import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import './style.css';
import App from './App.vue';
import router from './router';
import { createNotivue } from 'notivue';
import 'notivue/notifications.css';
import 'notivue/animations.css';

const app = createApp(App);
const pinia = createPinia();
const notivue = createNotivue({ position: 'bottom-right' });
pinia.use(piniaPluginPersistedstate);

app.use(pinia);
app.use(router);
app.use(notivue);

app.mount('#app');
