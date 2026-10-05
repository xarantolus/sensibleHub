import Oruga from '@oruga-ui/oruga-next'
import { bulmaConfig } from '@oruga-ui/theme-bulma'
import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { router } from './router'
import './styles/main.scss'

createApp(App).use(createPinia()).use(router).use(Oruga, bulmaConfig).mount('#app')
