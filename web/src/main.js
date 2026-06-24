import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

import VCalendar from 'v-calendar';
import './assets/scss/main.scss'
import 'v-calendar/style.css';

const app = createApp(App)

app.use(router)
   .use(VCalendar, {})

app.mount('#app')

window.WEBAPP = app

import API from '@/api'

window.WEBAPI = API