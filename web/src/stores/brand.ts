import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getBrandConfig } from '@/api/system'
import type { BrandConfig } from '@/types/system'

export const useBrandStore = defineStore('brand', () => {
  const systemName = ref('TicketDesk')
  const systemDescription = ref('项目化工单与告警联动系统')
  const copyrightText = ref('© 2026 TicketDesk. All rights reserved.')
  const logoUrl = ref('')
  const faviconUrl = ref('')
  const loginTitle = ref('工单与告警联动系统')
  const loginDescription = ref(
    '一切问题都是工单，一切告警都必须被跟进。\n为运维与技术团队打造的项目化工单管理平台。',
  )
  const loaded = ref(false)

  async function loadBrandConfig() {
    try {
      const res = await getBrandConfig()
      const config = res.data.data
      updateBrand(config)
      loaded.value = true
    } catch {
      // 品牌配置获取失败使用默认值
      loaded.value = true
    }
  }

  function updateBrand(config: BrandConfig) {
    systemName.value = config.system_name || 'TicketDesk'
    systemDescription.value = config.system_description || '项目化工单与告警联动系统'
    copyrightText.value = config.copyright_text || '© 2026 TicketDesk. All rights reserved.'
    logoUrl.value = config.logo_url || ''
    faviconUrl.value = config.favicon_url || ''
    loginTitle.value = config.login_title || '工单与告警联动系统'
    loginDescription.value =
      config.login_description ||
      '一切问题都是工单，一切告警都必须被跟进。\n为运维与技术团队打造的项目化工单管理平台。'

    // 动态更新 Favicon
    applyFavicon(config.favicon_url)
  }

  function applyFavicon(url: string) {
    const link = document.querySelector<HTMLLinkElement>("link[rel='icon']")
    if (link) {
      link.href = url || '/favicon.svg'
    }
  }

  return {
    systemName,
    systemDescription,
    copyrightText,
    logoUrl,
    faviconUrl,
    loginTitle,
    loginDescription,
    loaded,
    loadBrandConfig,
    updateBrand,
  }
})
