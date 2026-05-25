<template>
  <div class="api-docs-container">
    <iframe
      v-if="ready"
      :src="swaggerUrl"
      class="api-docs-frame"
      title="TicketDesk API 文档"
    ></iframe>
    <div v-else class="api-docs-loading">
      <el-icon class="loading-spin"><Loading /></el-icon>
      <p>正在加载 API 文档...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import request from '@/utils/request'

const router = useRouter()
const ready = ref(false)
const swaggerUrl = ref('')

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    ElMessage.warning('请先登录')
    router.replace('/login')
    return
  }

  try {
    // 调后端用 JWT 鉴权设 HttpOnly cookie (path=/api/v1/swagger), 后续 swagger UI 静态资源请求自动带 cookie
    await request.post('/auth/swagger-session')
    swaggerUrl.value = '/api/v1/swagger/index.html'
    ready.value = true
  } catch {
    ElMessage.error('初始化 API 文档失败, 请重试')
  }
})
</script>

<style scoped lang="scss">
.api-docs-container {
  width: 100%;
  height: calc(100vh - 60px);
  margin: calc(-1 * var(--td-space-6));
  background: var(--td-bg-card);
  display: flex;
  flex-direction: column;
}

.api-docs-frame {
  width: 100%;
  height: 100%;
  border: none;
  background: #fff;
}

.api-docs-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--td-space-3);
  color: var(--td-text-secondary);

  .loading-spin {
    font-size: 48px;
    color: var(--td-color-primary);
    animation: spin 1s linear infinite;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
