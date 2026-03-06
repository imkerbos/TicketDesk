import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

import App from './App.vue'
import router from './router'
import './styles/theme.scss'
import './styles/index.scss'
import { useUserStore } from './stores/user'
import { useThemeStore } from './stores/theme'
import { useBrandStore } from './stores/brand'

const app = createApp(App)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

const pinia = createPinia()
app.use(pinia)

// 初始化用户状态（必须在 pinia 注册后）
const userStore = useUserStore()
userStore.initFromStorage()

// 初始化主题
const themeStore = useThemeStore()
themeStore.init()

// 加载品牌配置
const brandStore = useBrandStore()

app.use(router)
app.use(ElementPlus)

// 等待品牌配置和路由初始导航完成后再挂载，避免闪屏
Promise.all([brandStore.loadBrandConfig(), router.isReady()]).then(() => {
  app.mount('#app')
})
