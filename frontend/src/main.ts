import { Autocomplete, Collapse, createOruga, Field, Modal, Notification, Switch } from '@oruga-ui/oruga-next'
import { bulmaConfig } from '@oruga-ui/theme-bulma'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { createPinia } from 'pinia'
import { createApp } from 'vue'

import { isApiError } from './api/client'
import { startServerEvents } from './api/events'
import App from './App.vue'
import { notifyError } from './lib/notify'
import { router } from './router'
import './styles/main.scss'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (failures, err) => failures < 3 && !(isApiError(err) && err.status >= 400 && err.status < 500),
    },
    mutations: {
      onError: (err) => {
        notifyError(err)
      },
    },
  },
})

const app = createApp(App)
app.config.errorHandler = (err, _instance, info) => {
  console.error(info, err)
  notifyError(err)
}

const oruga = createOruga(bulmaConfig, [Autocomplete, Collapse, Field, Modal, Notification, Switch])

app.use(createPinia()).use(router).use(VueQueryPlugin, { queryClient }).use(oruga).mount('#app')

startServerEvents(queryClient)
