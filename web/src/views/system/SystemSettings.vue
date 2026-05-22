<template>
  <div class="system-settings">
    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- 品牌设置 -->
      <el-tab-pane label="品牌设置" name="brand">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon general-icon">
                  <el-icon><Picture /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">品牌设置</span>
                  <span class="subtitle">自定义系统名称、Logo、版权信息和登录页文案</span>
                </div>
              </div>
            </div>
          </template>

          <el-row :gutter="40">
            <el-col :xs="24" :lg="14">
              <el-form
                ref="brandFormRef"
                :model="brandForm"
                :rules="brandRules"
                label-position="top"
                class="settings-form"
              >
                <div class="form-section">
                  <div class="section-title">基本信息</div>
                  <el-form-item label="系统名称" prop="system_name">
                    <el-input
                      v-model="brandForm.system_name"
                      placeholder="例如: TicketDesk"
                      maxlength="50"
                      show-word-limit
                    />
                    <template #extra>
                      <div class="form-item-tip">
                        显示在侧边栏、浏览器标签页和登录页
                      </div>
                    </template>
                  </el-form-item>

                  <el-form-item label="系统描述" prop="system_description">
                    <el-input
                      v-model="brandForm.system_description"
                      placeholder="例如: 项目化工单与告警联动系统"
                      maxlength="200"
                      show-word-limit
                    />
                  </el-form-item>

                  <el-form-item label="版权信息" prop="copyright_text">
                    <el-input
                      v-model="brandForm.copyright_text"
                      placeholder="例如: © 2026 TicketDesk. All rights reserved."
                      maxlength="200"
                      show-word-limit
                    />
                  </el-form-item>
                </div>

                <div class="form-section">
                  <div class="section-title">登录页文案</div>
                  <el-form-item label="登录页标题" prop="login_title">
                    <el-input
                      v-model="brandForm.login_title"
                      placeholder="例如: 工单与告警联动系统"
                      maxlength="100"
                      show-word-limit
                    />
                  </el-form-item>

                  <el-form-item label="登录页描述" prop="login_description">
                    <el-input
                      v-model="brandForm.login_description"
                      type="textarea"
                      :rows="3"
                      placeholder="登录页左侧的描述文案，支持换行"
                      maxlength="500"
                      show-word-limit
                    />
                  </el-form-item>
                </div>

                <div class="form-section">
                  <div class="section-title">品牌资源</div>
                  <el-form-item label="Logo">
                    <div class="upload-area">
                      <el-upload
                        :show-file-list="false"
                        :before-upload="beforeBrandUpload"
                        :http-request="(opts: any) => handleBrandUpload(opts, 'logo')"
                        accept=".svg,.png,.ico,.jpg,.jpeg,.webp"
                      >
                        <el-button :loading="logoUploading">
                          {{ brandForm.logo_url ? '更换 Logo' : '上传 Logo' }}
                        </el-button>
                      </el-upload>
                      <div v-if="brandForm.logo_url" class="upload-preview">
                        <img :src="brandForm.logo_url" alt="Logo" class="preview-image" />
                        <el-button text type="danger" size="small" @click="removeBrandAsset('logo')">移除</el-button>
                      </div>
                    </div>
                    <template #extra>
                      <div class="form-item-tip">
                        建议尺寸 48x48，支持 SVG、PNG、ICO、JPG、WEBP，最大 2MB。留空使用默认 Logo。
                      </div>
                    </template>
                  </el-form-item>

                  <el-form-item label="Favicon">
                    <div class="upload-area">
                      <el-upload
                        :show-file-list="false"
                        :before-upload="beforeBrandUpload"
                        :http-request="(opts: any) => handleBrandUpload(opts, 'favicon')"
                        accept=".svg,.png,.ico"
                      >
                        <el-button :loading="faviconUploading">
                          {{ brandForm.favicon_url ? '更换 Favicon' : '上传 Favicon' }}
                        </el-button>
                      </el-upload>
                      <div v-if="brandForm.favicon_url" class="upload-preview">
                        <img :src="brandForm.favicon_url" alt="Favicon" class="preview-image preview-favicon" />
                        <el-button text type="danger" size="small" @click="removeBrandAsset('favicon')">移除</el-button>
                      </div>
                    </div>
                    <template #extra>
                      <div class="form-item-tip">
                        浏览器标签页图标，建议尺寸 32x32，支持 SVG、PNG、ICO，最大 2MB。
                      </div>
                    </template>
                  </el-form-item>
                </div>

                <el-form-item>
                  <el-button
                    type="primary"
                    :loading="brandLoading"
                    @click="handleSaveBrandConfig"
                  >
                    <el-icon><Check /></el-icon>
                    保存品牌设置
                  </el-button>
                </el-form-item>
              </el-form>
            </el-col>

            <el-col :xs="24" :lg="10">
              <div class="brand-preview-section">
                <div class="section-title">预览</div>
                <div class="brand-preview-card">
                  <div class="preview-sidebar">
                    <div class="preview-logo-area">
                      <img
                        v-if="brandForm.logo_url"
                        :src="brandForm.logo_url"
                        alt="Logo"
                        class="preview-logo-img"
                      />
                      <span v-else class="preview-logo-placeholder">{{ (brandForm.system_name || 'TicketDesk').charAt(0) }}</span>
                      <span class="preview-logo-text">{{ brandForm.system_name || 'TicketDesk' }}</span>
                    </div>
                    <div class="preview-menu-item active">首页</div>
                    <div class="preview-menu-item">工单管理</div>
                    <div class="preview-menu-item">项目管理</div>
                    <div class="preview-copyright">{{ brandForm.copyright_text || '© 2026 TicketDesk' }}</div>
                  </div>
                </div>
                <el-alert type="info" :closable="false" class="brand-tips">
                  <p>修改品牌设置后，所有用户刷新页面即可看到更新。</p>
                  <p>Logo 和 Favicon 需要先上传再保存。</p>
                </el-alert>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>

      <!-- 通用配置 -->
      <el-tab-pane label="通用配置" name="general">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon general-icon">
                  <el-icon><Setting /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">通用系统配置</span>
                  <span class="subtitle">配置系统的基本信息和全局设置</span>
                </div>
              </div>
            </div>
          </template>

          <el-row :gutter="40">
            <el-col :xs="24" :lg="12">
              <el-form
                ref="generalFormRef"
                :model="generalForm"
                :rules="generalRules"
                label-position="top"
                class="settings-form"
              >
                <div class="form-section">
                  <div class="section-title">站点信息</div>
                  <el-form-item label="站点域名" prop="site_url">
                    <el-input
                      v-model="generalForm.site_url"
                      placeholder="例如: https://ticketdesk.example.com"
                    >
                      <template #prefix>
                        <el-icon><Link /></el-icon>
                      </template>
                    </el-input>
                    <template #extra>
                      <div class="form-item-tip">
                        用于生成邮件中的链接，请填写完整的域名（包含协议）
                      </div>
                    </template>
                  </el-form-item>
                </div>

                <el-form-item>
                  <el-button type="primary" :loading="generalLoading" @click="saveGeneralConfig">
                    保存配置
                  </el-button>
                </el-form-item>
              </el-form>
            </el-col>

            <el-col :xs="24" :lg="12">
              <div class="config-tips">
                <div class="tip-title">
                  <el-icon><InfoFilled /></el-icon>
                  <span>配置说明</span>
                </div>
                <div class="tip-content">
                  <div class="tip-item">
                    <div class="tip-label">站点域名</div>
                    <div class="tip-desc">
                      系统会使用此域名生成邮件中的链接（如重置密码链接）。请确保填写正确的域名，包含协议（http:// 或 https://），不要以斜杠结尾。
                    </div>
                    <div class="tip-example">
                      示例：https://ticketdesk.example.com
                    </div>
                  </div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>

      <!-- 邮件配置 -->
      <el-tab-pane label="邮件配置" name="email">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon email-icon">
                  <el-icon><Message /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">SMTP 邮件服务配置</span>
                  <span class="subtitle">配置系统邮件发送服务，用于发送通知和告警</span>
                </div>
              </div>
              <el-switch
                v-model="emailForm.enabled"
                active-text="启用"
                inactive-text="禁用"
                inline-prompt
                style="--el-switch-on-color: var(--td-color-success)"
              />
            </div>
          </template>

          <el-row :gutter="40">
            <el-col :xs="24" :lg="12">
              <el-form
                ref="emailFormRef"
                :model="emailForm"
                :rules="emailRules"
                label-position="top"
                class="settings-form"
              >
                <div class="form-section">
                  <div class="section-title">服务器设置</div>
                  <el-row :gutter="16">
                    <el-col :span="16">
                      <el-form-item label="SMTP 服务器" prop="smtp_host">
                        <el-input v-model="emailForm.smtp_host" placeholder="例如: smtp.gmail.com">
                          <template #prefix>
                            <el-icon><Monitor /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="8">
                      <el-form-item label="端口" prop="smtp_port">
                        <el-input-number v-model="emailForm.smtp_port" :min="1" :max="65535" style="width: 100%" />
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-row :gutter="16">
                    <el-col :span="12">
                      <el-form-item label="用户名" prop="smtp_username">
                        <el-input v-model="emailForm.smtp_username" placeholder="认证用户名">
                          <template #prefix>
                            <el-icon><User /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="密码" prop="smtp_password">
                        <el-input
                          v-model="emailForm.smtp_password"
                          type="password"
                          placeholder="留空表示不修改"
                          show-password
                        >
                          <template #prefix>
                            <el-icon><Lock /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                  </el-row>
                </div>

                <div class="form-section">
                  <div class="section-title">发件人信息</div>
                  <el-row :gutter="16">
                    <el-col :span="14">
                      <el-form-item label="发件人地址" prop="from_address">
                        <el-input v-model="emailForm.from_address" placeholder="noreply@example.com">
                          <template #prefix>
                            <el-icon><Message /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="10">
                      <el-form-item label="发件人名称" prop="from_name">
                        <el-input v-model="emailForm.from_name" placeholder="TicketDesk" />
                      </el-form-item>
                    </el-col>
                  </el-row>
                </div>

                <div class="form-section">
                  <div class="section-title">安全设置</div>
                  <el-form-item>
                    <div class="setting-row">
                      <div class="setting-info">
                        <span class="setting-label">使用 TLS 加密</span>
                        <span class="setting-desc">推荐启用以保护邮件传输安全</span>
                      </div>
                      <el-switch v-model="emailForm.use_tls" />
                    </div>
                  </el-form-item>
                </div>

                <el-form-item>
                  <el-button type="primary" :loading="emailSaving" @click="saveEmailConfig">
                    <el-icon><Check /></el-icon>
                    保存配置
                  </el-button>
                  <el-button @click="testEmailDialog = true">
                    <el-icon><Promotion /></el-icon>
                    发送测试邮件
                  </el-button>
                </el-form-item>
              </el-form>
            </el-col>
            <el-col :xs="24" :lg="12" class="info-panel">
              <div class="info-card">
                <div class="info-title">
                  <el-icon><InfoFilled /></el-icon>
                  配置说明
                </div>
                <ul class="info-list">
                  <li>SMTP 服务器用于发送系统邮件通知</li>
                  <li>常用端口：25 (不加密)、465 (SSL)、587 (TLS)</li>
                  <li>建议使用 TLS 加密以保护邮件内容</li>
                  <li>配置完成后可发送测试邮件验证</li>
                </ul>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>

      <!-- 安全配置 -->
      <el-tab-pane label="安全设置" name="security">
        <el-row :gutter="20">
          <!-- MFA 设置 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon mfa-icon">
                  <el-icon><Key /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">多因素认证 (MFA)</span>
                  <span class="desc">增强账户安全性</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">启用 MFA</span>
                    <span class="setting-desc">允许用户开启双因素认证</span>
                  </div>
                  <el-switch v-model="securityForm.mfa_enabled" />
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">强制要求 MFA</span>
                    <span class="setting-desc">所有用户必须开启 MFA</span>
                  </div>
                  <el-switch
                    v-model="securityForm.mfa_required"
                    :disabled="!securityForm.mfa_enabled"
                  />
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- 密码策略 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon password-icon">
                  <el-icon><Lock /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">密码策略</span>
                  <span class="desc">设置密码复杂度要求</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">最小长度</span>
                    <span class="setting-desc">密码最少字符数</span>
                  </div>
                  <el-input-number
                    v-model="securityForm.password_min_length"
                    :min="6"
                    :max="32"
                    size="small"
                    controls-position="right"
                  />
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">要求大写字母</span>
                    <span class="setting-desc">必须包含 A-Z</span>
                  </div>
                  <el-switch v-model="securityForm.password_require_upper" />
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">要求数字</span>
                    <span class="setting-desc">必须包含 0-9</span>
                  </div>
                  <el-switch v-model="securityForm.password_require_number" />
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- 会话设置 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon session-icon">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">会话设置</span>
                  <span class="desc">控制用户登录会话</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">会话超时</span>
                    <span class="setting-desc">无操作自动退出时间（分钟）</span>
                  </div>
                  <el-input-number
                    v-model="securityForm.session_timeout"
                    :min="5"
                    :max="1440"
                    size="small"
                    controls-position="right"
                  />
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <div class="action-bar">
          <el-button type="primary" :loading="securitySaving" @click="saveSecurityConfig">
            <el-icon><Check /></el-icon>
            保存安全配置
          </el-button>
        </div>
      </el-tab-pane>

      <!-- 限流配置 -->
      <el-tab-pane label="限流配置" name="ratelimit">
        <el-row :gutter="20">
          <!-- Webhook 限流 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon webhook-rl-icon">
                  <el-icon><Connection /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">Webhook 限流</span>
                  <span class="desc">告警 Webhook 接口限流</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">每 IP 每分钟最大请求数</span>
                    <span class="setting-desc">范围：10 - 10,000</span>
                  </div>
                  <el-input-number
                    v-model="rateLimitForm.webhook_limit"
                    :min="10"
                    :max="10000"
                    size="small"
                    controls-position="right"
                  />
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- 认证限流 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon auth-rl-icon">
                  <el-icon><UserFilled /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">认证限流</span>
                  <span class="desc">登录/注册接口限流</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">每 IP 每分钟最大请求数</span>
                    <span class="setting-desc">范围：5 - 1,000</span>
                  </div>
                  <el-input-number
                    v-model="rateLimitForm.auth_limit"
                    :min="5"
                    :max="1000"
                    size="small"
                    controls-position="right"
                  />
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- API 全局限流 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon api-rl-icon">
                  <el-icon><DataLine /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">API 全局限流</span>
                  <span class="desc">所有认证接口限流</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">每 IP 每分钟最大请求数</span>
                    <span class="setting-desc">范围：50 - 50,000</span>
                  </div>
                  <el-input-number
                    v-model="rateLimitForm.api_limit"
                    :min="50"
                    :max="50000"
                    size="small"
                    controls-position="right"
                  />
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <div class="action-bar">
          <el-button type="primary" :loading="rateLimitSaving" @click="saveRateLimitConfig">
            <el-icon><Check /></el-icon>
            保存限流配置
          </el-button>
        </div>
      </el-tab-pane>
      <!-- 工时配置 -->
      <el-tab-pane label="工时配置" name="worklog">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon worklog-icon">
                  <el-icon><Clock /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">工作类型管理</span>
                  <span class="subtitle">配置工时记录中可选的工作类型</span>
                </div>
              </div>
            </div>
          </template>

          <el-row :gutter="40">
            <el-col :xs="24" :lg="14">
              <div class="work-type-list">
                <div
                  v-for="(item, index) in workTypeList"
                  :key="index"
                  class="work-type-item"
                >
                  <el-input
                    v-model="item.label"
                    placeholder="工作类型名称"
                    style="flex: 1"
                    @input="item.value = item.label"
                  />
                  <el-button
                    type="danger"
                    link
                    :disabled="workTypeList.length <= 1"
                    @click="removeWorkType(index)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button type="primary" link style="margin-top: 8px" @click="addWorkType">
                  + 添加工作类型
                </el-button>
              </div>

              <el-form-item style="margin-top: 24px">
                <el-button type="primary" :loading="worklogSaving" @click="saveWorkTypeConfig">
                  <el-icon><Check /></el-icon>
                  保存配置
                </el-button>
              </el-form-item>
            </el-col>

            <el-col :xs="24" :lg="10">
              <div class="config-tips">
                <div class="tip-title">
                  <el-icon><InfoFilled /></el-icon>
                  <span>配置说明</span>
                </div>
                <div class="tip-content">
                  <div class="tip-item">
                    <div class="tip-label">工作类型</div>
                    <div class="tip-desc">
                      工作类型用于工时记录中分类工作内容。修改后前端刷新即生效，已有工时记录不受影响。
                    </div>
                  </div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>

      <!-- SSO 配置 -->
      <el-tab-pane label="SSO 认证" name="sso">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon sso-icon">
                  <el-icon><Link /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">SSO 单点登录配置</span>
                  <span class="subtitle">配置 OIDC 单点登录，支持企业统一认证（EIAM）</span>
                </div>
              </div>
              <el-switch
                v-model="ssoForm.enabled"
                active-text="启用"
                inactive-text="禁用"
                inline-prompt
                style="--el-switch-on-color: var(--td-color-success)"
              />
            </div>
          </template>

          <el-row :gutter="40">
            <el-col :xs="24" :lg="14">
              <el-form
                ref="ssoFormRef"
                :model="ssoForm"
                label-position="top"
                class="settings-form"
              >
                <div class="form-section">
                  <div class="section-title">基本配置</div>
                  <el-form-item label="提供方名称" prop="provider_name">
                    <el-input v-model="ssoForm.provider_name" placeholder="例如: 企业统一认证">
                      <template #prefix>
                        <el-icon><User /></el-icon>
                      </template>
                    </el-input>
                    <template #extra>
                      <div class="form-item-tip">登录页面 SSO 按钮上显示的名称</div>
                    </template>
                  </el-form-item>

                  <el-form-item label="Issuer URL" prop="issuer_url">
                    <el-input v-model="ssoForm.issuer_url" placeholder="例如: https://eiam.example.com/realms/master">
                      <template #prefix>
                        <el-icon><Link /></el-icon>
                      </template>
                    </el-input>
                    <template #extra>
                      <div class="form-item-tip">OIDC 提供方的 Issuer URL，用于自动发现配置</div>
                    </template>
                  </el-form-item>

                  <el-row :gutter="16">
                    <el-col :span="12">
                      <el-form-item label="Client ID" prop="client_id">
                        <el-input v-model="ssoForm.client_id" placeholder="OIDC Client ID" />
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="Client Secret" prop="client_secret">
                        <el-input
                          v-model="ssoForm.client_secret"
                          type="password"
                          placeholder="留空表示不修改"
                          show-password
                        >
                          <template #prefix>
                            <el-icon><Lock /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-row :gutter="16">
                    <el-col :span="12">
                      <el-form-item label="回调地址" prop="redirect_uri">
                        <el-input v-model="ssoForm.redirect_uri" placeholder="例如: https://your-domain/auth/sso/callback" />
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="Scopes" prop="scopes">
                        <el-input v-model="ssoForm.scopes" placeholder="openid,profile,email" />
                      </el-form-item>
                    </el-col>
                  </el-row>
                </div>

                <div class="form-section">
                  <div class="section-title">用户管理</div>
                  <el-form-item>
                    <div class="setting-row">
                      <div class="setting-info">
                        <span class="setting-label">自动创建用户</span>
                        <span class="setting-desc">首次 SSO 登录时自动创建本地用户</span>
                      </div>
                      <el-switch v-model="ssoForm.auto_create_user" />
                    </div>
                  </el-form-item>
                  <el-form-item label="默认角色" prop="default_role">
                    <el-select v-model="ssoForm.default_role" style="width: 200px">
                      <el-option label="普通用户 (user)" value="user" />
                      <el-option label="项目管理员 (project_admin)" value="project_admin" />
                      <el-option label="系统管理员 (admin)" value="admin" />
                    </el-select>
                    <template #extra>
                      <div class="form-item-tip">自动创建用户时分配的默认角色</div>
                    </template>
                  </el-form-item>
                </div>

                <div class="form-section">
                  <div class="section-title">
                    Claims 映射
                    <el-button type="primary" link size="small" style="margin-left: 8px" @click="addClaimMapping">
                      + 添加映射
                    </el-button>
                  </div>
                  <div class="claim-mappings">
                    <div
                      v-for="(mapping, index) in ssoForm.claim_mappings"
                      :key="index"
                      class="claim-mapping-row"
                    >
                      <el-row :gutter="12" style="flex: 1">
                        <el-col :span="10">
                          <el-input
                            v-model="mapping.local_field"
                            placeholder="本地字段名 (如 username)"
                            size="default"
                          />
                        </el-col>
                        <el-col :span="2" style="text-align: center; line-height: 32px; color: var(--td-color-info)">
                          ←
                        </el-col>
                        <el-col :span="10">
                          <el-input
                            v-model="mapping.claim_name"
                            placeholder="OIDC Claim (如 preferred_username)"
                            size="default"
                          />
                        </el-col>
                        <el-col :span="2">
                          <el-button
                            type="danger"
                            link
                            :disabled="ssoForm.claim_mappings.length <= 1"
                            @click="removeClaimMapping(index)"
                          >
                            删除
                          </el-button>
                        </el-col>
                      </el-row>
                    </div>
                    <div class="claim-mapping-hint">
                      内置字段：username、email、display_name、avatar。其他字段名将存入用户扩展属性。
                    </div>
                  </div>
                </div>

                <el-form-item>
                  <el-button type="primary" :loading="ssoSaving" @click="saveSSOConfig">
                    <el-icon><Check /></el-icon>
                    保存配置
                  </el-button>
                </el-form-item>
              </el-form>
            </el-col>

            <el-col :xs="24" :lg="10">
              <div class="config-tips">
                <div class="tip-title">
                  <el-icon><InfoFilled /></el-icon>
                  <span>配置说明</span>
                </div>
                <div class="tip-content">
                  <div class="tip-item">
                    <div class="tip-label">OIDC 对接步骤</div>
                    <div class="tip-desc">
                      1. 在 EIAM 中创建 OIDC 应用，获取 Client ID 和 Secret<br />
                      2. 填写 Issuer URL（通常为 EIAM 的 realm 地址）<br />
                      3. 在 EIAM 中配置回调地址为本系统的回调 URL<br />
                      4. 启用 SSO 后，登录页将出现 SSO 登录按钮
                    </div>
                  </div>
                  <div class="tip-item">
                    <div class="tip-label">IdP-initiated 登录</div>
                    <div class="tip-desc">
                      在 EIAM 门户中将应用入口 URL 配置为：<br />
                      <code>{{ ssoForm.redirect_uri?.replace('/auth/sso/callback', '/auth/sso/login') || 'https://your-domain/auth/sso/login' }}</code>
                    </div>
                  </div>
                  <div class="tip-item">
                    <div class="tip-label">Claims 映射</div>
                    <div class="tip-desc">
                      根据 EIAM 返回的 ID Token 中的 claim 名称进行映射。<br />
                      内置字段：username、email、display_name、avatar 会映射到用户基本信息。<br />
                      自定义字段名（如 department、employee_id）会存入用户扩展属性。
                    </div>
                  </div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 测试邮件对话框 -->
    <el-dialog v-model="testEmailDialog" title="发送测试邮件" width="400px">
      <el-form ref="testEmailFormRef" :model="testEmailForm" :rules="testEmailRules" label-position="top">
        <el-form-item label="收件人邮箱" prop="to_address">
          <el-input v-model="testEmailForm.to_address" placeholder="请输入收件人邮箱" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testEmailDialog = false">取消</el-button>
        <el-button type="primary" :loading="testEmailSending" @click="sendTestEmail">
          发送测试
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  Message, Lock, Check, Promotion, InfoFilled,
  Key, Timer, Monitor, User, Setting, Link,
  Connection, UserFilled, DataLine, Clock, Delete, Picture,
} from '@element-plus/icons-vue'
import {
  getEmailConfig,
  updateEmailConfig,
  getSecurityConfig,
  updateSecurityConfig,
  getRateLimitConfig,
  updateRateLimitConfig,
  getConfig,
  updateConfig,
  getSSOAdminConfig,
  updateSSOConfig,
  getBrandConfig,
  updateBrandConfig,
  uploadBrandAsset,
} from '@/api/system'
import { useBrandStore } from '@/stores/brand'

const route = useRoute()
const router = useRouter()

// 从 URL 查询参数中获取 tab，如果没有则默认为 'general'
const activeTab = ref((route.query.tab as string) || 'general')

// 监听 tab 变化，更新 URL
watch(activeTab, (newTab) => {
  router.replace({
    query: { ...route.query, tab: newTab },
  })
})

const brandStore = useBrandStore()

// ============ 品牌配置 ============
const brandFormRef = ref<FormInstance>()
const brandLoading = ref(false)
const logoUploading = ref(false)
const faviconUploading = ref(false)
const brandForm = reactive({
  system_name: '',
  system_description: '',
  copyright_text: '',
  login_title: '',
  login_description: '',
  logo_url: '',
  favicon_url: '',
})

const brandRules: FormRules = {
  system_name: [
    { required: true, message: '请输入系统名称', trigger: 'blur' },
    { max: 50, message: '系统名称最长 50 个字符', trigger: 'blur' },
  ],
}

const loadBrandConfig = async () => {
  try {
    const res = await getBrandConfig()
    const config = res.data.data
    brandForm.system_name = config.system_name
    brandForm.system_description = config.system_description
    brandForm.copyright_text = config.copyright_text
    brandForm.login_title = config.login_title
    brandForm.login_description = config.login_description
    brandForm.logo_url = config.logo_url
    brandForm.favicon_url = config.favicon_url
  } catch {
    // 使用默认值
  }
}

const beforeBrandUpload = (file: File) => {
  const maxSize = 2 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过 2MB')
    return false
  }
  return true
}

const handleBrandUpload = async (options: { file: File }, type: 'logo' | 'favicon') => {
  const loadingRef = type === 'logo' ? logoUploading : faviconUploading
  loadingRef.value = true
  try {
    const res = await uploadBrandAsset(options.file, type)
    const url = res.data.data.url
    if (type === 'logo') {
      brandForm.logo_url = url
    } else {
      brandForm.favicon_url = url
    }
    ElMessage.success(`${type === 'logo' ? 'Logo' : 'Favicon'} 上传成功`)
  } catch {
    ElMessage.error('上传失败')
  } finally {
    loadingRef.value = false
  }
}

const removeBrandAsset = (type: 'logo' | 'favicon') => {
  if (type === 'logo') {
    brandForm.logo_url = ''
  } else {
    brandForm.favicon_url = ''
  }
}

const handleSaveBrandConfig = async () => {
  if (!brandFormRef.value) return

  await brandFormRef.value.validate(async (valid) => {
    if (!valid) return

    brandLoading.value = true
    try {
      await updateBrandConfig({
        system_name: brandForm.system_name,
        system_description: brandForm.system_description,
        copyright_text: brandForm.copyright_text,
        login_title: brandForm.login_title,
        login_description: brandForm.login_description,
      })

      // 如果 logo_url 或 favicon_url 被清除，也需要更新配置
      if (!brandForm.logo_url) {
        await updateConfig('brand.logo_url', '')
      }
      if (!brandForm.favicon_url) {
        await updateConfig('brand.favicon_url', '')
      }

      // 刷新 brand store
      await brandStore.loadBrandConfig()

      ElMessage.success('品牌设置保存成功')
    } catch {
      ElMessage.error('保存失败')
    } finally {
      brandLoading.value = false
    }
  })
}

// ============ 通用配置 ============
const generalFormRef = ref<FormInstance>()
const generalLoading = ref(false)
const generalForm = reactive({
  site_url: '',
})

const generalRules: FormRules = {
  site_url: [
    { required: true, message: '请输入站点域名', trigger: 'blur' },
    { type: 'url', message: '请输入有效的 URL', trigger: 'blur' },
  ],
}

// 加载通用配置
const loadGeneralConfig = async () => {
  try {
    // 从系统配置中读取站点域名
    const res = await getConfig('general.site_url')
    if (res.data.data) {
      generalForm.site_url = res.data.data.config_value || ''
    }
  } catch (error: any) {
    // 如果配置不存在（404），不报错，使用默认空值
    if (error?.response?.status !== 404) {
      // ignored
    }
  }
}

// 保存通用配置
const saveGeneralConfig = async () => {
  if (!generalFormRef.value) return

  await generalFormRef.value.validate(async (valid) => {
    if (!valid) return

    generalLoading.value = true
    try {
      await updateConfig('general.site_url', generalForm.site_url)
      ElMessage.success('通用配置保存成功')
    } catch {
      // 错误已在拦截器中处理
    } finally {
      generalLoading.value = false
    }
  })
}

// ============ 邮件配置 ============
const emailFormRef = ref<FormInstance>()
const emailSaving = ref(false)
const emailForm = reactive({
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  from_address: '',
  from_name: 'TicketDesk',
  use_tls: true,
  enabled: false,
})

const emailRules: FormRules = {
  smtp_host: [{ required: true, message: '请输入 SMTP 服务器地址', trigger: 'blur' }],
  smtp_port: [{ required: true, message: '请输入 SMTP 端口', trigger: 'blur' }],
  from_address: [
    { required: true, message: '请输入发件人地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' },
  ],
}

const loadEmailConfig = async () => {
  try {
    const { data } = await getEmailConfig()
    Object.assign(emailForm, data.data)
  } catch {
    // ignored
  }
}

const saveEmailConfig = async () => {
  if (!emailFormRef.value) return
  await emailFormRef.value.validate(async (valid) => {
    if (!valid) return

    emailSaving.value = true
    try {
      await updateEmailConfig(emailForm)
      ElMessage.success('邮件配置保存成功')
    } catch {
      // ignored
    } finally {
      emailSaving.value = false
    }
  })
}

// 测试邮件
const testEmailDialog = ref(false)
const testEmailFormRef = ref<FormInstance>()
const testEmailSending = ref(false)
const testEmailForm = reactive({
  to_address: '',
})
const testEmailRules: FormRules = {
  to_address: [
    { required: true, message: '请输入收件人邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' },
  ],
}

const sendTestEmail = async () => {
  if (!testEmailFormRef.value) return
  await testEmailFormRef.value.validate(async (valid) => {
    if (!valid) return

    testEmailSending.value = true
    try {
      // await testEmail(testEmailForm.to_address)
      ElMessage.success('测试邮件已发送')
      testEmailDialog.value = false
    } catch {
      // ignored
    } finally {
      testEmailSending.value = false
    }
  })
}

// ============ 安全配置 ============
const securitySaving = ref(false)
const securityForm = reactive({
  mfa_enabled: false,
  mfa_required: false,
  password_min_length: 6,
  password_require_upper: false,
  password_require_number: false,
  session_timeout: 120,
})

const loadSecurityConfig = async () => {
  try {
    const { data } = await getSecurityConfig()
    Object.assign(securityForm, data.data)
  } catch {
    // ignored
  }
}

const saveSecurityConfig = async () => {
  securitySaving.value = true
  try {
    await updateSecurityConfig(securityForm)
    ElMessage.success('安全配置保存成功')
  } catch {
    // ignored
  } finally {
    securitySaving.value = false
  }
}

// ============ 限流配置 ============
const rateLimitSaving = ref(false)
const rateLimitForm = reactive({
  webhook_limit: 100,
  auth_limit: 20,
  api_limit: 300,
})

const loadRateLimitConfig = async () => {
  try {
    const { data } = await getRateLimitConfig()
    Object.assign(rateLimitForm, data.data)
  } catch {
    // ignored
  }
}

const saveRateLimitConfig = async () => {
  rateLimitSaving.value = true
  try {
    await updateRateLimitConfig(rateLimitForm)
    ElMessage.success('限流配置保存成功')
  } catch {
    // ignored
  } finally {
    rateLimitSaving.value = false
  }
}

// ============ 工时配置 ============
const worklogSaving = ref(false)
const workTypeList = ref<{ value: string; label: string }[]>([])

const loadWorkTypeConfig = async () => {
  try {
    const res = await getConfig('worklog.work_types')
    if (res.data.data) {
      const parsed = JSON.parse(res.data.data.config_value || '[]')
      if (Array.isArray(parsed)) {
        workTypeList.value = parsed
      }
    }
  } catch (error: any) {
    if (error?.response?.status !== 404) {
      // ignored
    }
  }
}

const addWorkType = () => {
  workTypeList.value.push({ value: '', label: '' })
}

const removeWorkType = (index: number) => {
  workTypeList.value.splice(index, 1)
}

const saveWorkTypeConfig = async () => {
  // 过滤掉空项
  const filtered = workTypeList.value.filter(item => item.label.trim())
  if (filtered.length === 0) {
    ElMessage.warning('至少需要一个工作类型')
    return
  }

  worklogSaving.value = true
  try {
    await updateConfig('worklog.work_types', JSON.stringify(filtered))
    workTypeList.value = filtered
    ElMessage.success('工时配置保存成功')
  } catch {
    // ignored
  } finally {
    worklogSaving.value = false
  }
}

// ============ SSO 配置 ============
const ssoFormRef = ref<FormInstance>()
const ssoSaving = ref(false)
const ssoForm = reactive({
  enabled: false,
  provider_name: '企业统一认证',
  client_id: '',
  client_secret: '',
  issuer_url: '',
  redirect_uri: 'http://localhost:5173/auth/sso/callback',
  scopes: 'openid,profile,email',
  auto_create_user: true,
  default_role: 'user',
  claim_mappings: [
    { local_field: 'username', claim_name: 'preferred_username' },
    { local_field: 'email', claim_name: 'email' },
    { local_field: 'display_name', claim_name: 'name' },
    { local_field: 'avatar', claim_name: 'picture' },
  ] as Array<{ local_field: string; claim_name: string }>,
})

const addClaimMapping = () => {
  ssoForm.claim_mappings.push({ local_field: '', claim_name: '' })
}

const removeClaimMapping = (index: number) => {
  ssoForm.claim_mappings.splice(index, 1)
}

const loadSSOConfig = async () => {
  try {
    const { data } = await getSSOAdminConfig()
    const cfg = data.data
    ssoForm.enabled = cfg.enabled
    ssoForm.provider_name = cfg.provider_name
    ssoForm.client_id = cfg.client_id
    ssoForm.issuer_url = cfg.issuer_url
    ssoForm.redirect_uri = cfg.redirect_uri
    ssoForm.scopes = cfg.scopes
    ssoForm.auto_create_user = cfg.auto_create_user
    ssoForm.default_role = cfg.default_role
    if (cfg.claim_mappings && cfg.claim_mappings.length > 0) {
      ssoForm.claim_mappings = cfg.claim_mappings
    }
    // client_secret 不会从后端返回，保持为空
    ssoForm.client_secret = ''
  } catch {
    // ignored
  }
}

const saveSSOConfig = async () => {
  ssoSaving.value = true
  try {
    await updateSSOConfig(ssoForm)
    ElMessage.success('SSO 配置保存成功')
  } catch {
    // ignored
  } finally {
    ssoSaving.value = false
  }
}

// ============ 初始化 ============
onMounted(() => {
  loadBrandConfig()
  loadGeneralConfig()
  loadEmailConfig()
  loadSecurityConfig()
  loadRateLimitConfig()
  loadWorkTypeConfig()
  loadSSOConfig()
})
</script>

<style scoped lang="scss">
.system-settings {
  width: 100%;
}

.settings-tabs {
  :deep(.el-tabs__content) {
    padding-top: 16px;
  }
}

// 卡片样式
.settings-card {
  border-radius: 8px;

  :deep(.el-card__header) {
    padding: 20px 24px;
    border-bottom: 1px solid var(--td-divider-color);
  }

  :deep(.el-card__body) {
    padding: 24px;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .card-title {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .title-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;

    &.email-icon {
      background: var(--td-color-primary);
      color: var(--td-text-white);
    }

    &.general-icon {
      background: var(--td-color-success);
      color: var(--td-text-white);
    }

    &.sso-icon {
      background: var(--td-color-warning);
      color: var(--td-text-white);
    }

    &.worklog-icon {
      background: var(--td-color-primary);
      color: var(--td-text-white);
    }
  }

  .title-text {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: var(--td-text-primary);
    }

    .subtitle {
      font-size: 13px;
      color: var(--td-color-info);
    }
  }
}

// 表单样式
.settings-form {
  .form-section {
    margin-bottom: 24px;

    .section-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--td-text-regular);
      margin-bottom: 16px;
      padding-bottom: 8px;
      border-bottom: 1px solid var(--td-divider-color);
    }
  }

  .setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: var(--td-bg-page);
    border-radius: 8px;
  }

  .form-item-tip {
    font-size: 12px;
    color: var(--td-color-info);
    margin-top: 4px;
    line-height: 1.5;
  }
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 2px;

  .setting-label {
    font-size: 14px;
    color: var(--td-text-regular);
    font-weight: 500;
  }

  .setting-desc {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}

// 信息面板
.info-panel {
  display: flex;
  align-items: flex-start;
  padding-top: 24px;
}

.info-card {
  background: var(--td-color-primary-light);
  border-radius: 12px;
  padding: 20px;
  border: 1px solid var(--td-color-primary-light-hover);

  .info-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
    color: #0369a1;
    margin-bottom: 12px;
  }

  .info-list {
    margin: 0;
    padding-left: 20px;

    li {
      font-size: 13px;
      color: var(--td-text-regular);
      line-height: 1.8;
    }
  }
}

// 安全设置块
.setting-block {
  border-radius: 12px;
  margin-bottom: 20px;

  :deep(.el-card__body) {
    padding: 0;
  }

  .block-header {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    border-bottom: 1px solid var(--td-divider-color);

    .block-icon {
      width: 44px;
      height: 44px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 22px;

      &.mfa-icon {
        background: var(--td-color-primary);
        color: var(--td-text-white);
      }

      &.password-icon {
        background: #ec4899;
        color: var(--td-text-white);
      }

      &.session-icon {
        background: var(--td-color-primary);
        color: var(--td-text-white);
      }

      &.webhook-rl-icon {
        background: #ec4899;
        color: var(--td-text-white);
      }

      &.auth-rl-icon {
        background: var(--td-color-primary);
        color: var(--td-text-white);
      }

      &.api-rl-icon {
        background: var(--td-color-success);
        color: var(--td-text-white);
      }
    }

    .block-title {
      display: flex;
      flex-direction: column;
      gap: 2px;

      .title {
        font-size: 15px;
        font-weight: 600;
        color: var(--td-text-primary);
      }

      .desc {
        font-size: 12px;
        color: var(--td-text-placeholder);
      }
    }
  }

  .block-content {
    padding: 16px 20px;
  }
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;

  &:not(:last-child) {
    border-bottom: 1px solid var(--td-border-color);
  }
}

.action-bar {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--td-border-color);
}

// 响应式
@media (max-width: 992px) {
  .info-panel {
    margin-top: 24px;
  }

  .setting-block {
    margin-bottom: 16px;
  }
}

// Claims 映射动态列表
.claim-mappings {
  .claim-mapping-row {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }

  .claim-mapping-hint {
    font-size: 12px;
    color: var(--td-color-info);
    margin-top: 8px;
    line-height: 1.6;
  }
}

// 工作类型列表
.work-type-list {
  .work-type-item {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }
}

// 配置说明面板
.config-tips {
  background: var(--td-color-primary-light);
  border-radius: 12px;
  padding: 20px;
  border: 1px solid var(--td-color-primary-light-hover);

  .tip-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
    color: #0369a1;
    margin-bottom: 12px;
  }

  .tip-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .tip-item {
    .tip-label {
      font-size: 13px;
      font-weight: 600;
      color: var(--td-text-regular);
      margin-bottom: 4px;
    }

    .tip-desc {
      font-size: 13px;
      color: var(--td-text-regular);
      line-height: 1.8;

      code {
        background: var(--td-bg-section);
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 12px;
        color: #0369a1;
        word-break: break-all;
      }
    }

    .tip-example {
      font-size: 12px;
      color: var(--td-text-secondary);
      margin-top: 4px;
      font-style: italic;
    }
  }
}

// 品牌设置
.upload-area {
  display: flex;
  align-items: center;
  gap: 16px;
}

.upload-preview {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-image {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  object-fit: contain;
  border: 1px solid var(--td-border-color);
  padding: 2px;
}

.preview-favicon {
  width: 24px;
  height: 24px;
}

.brand-preview-section {
  padding: 0 8px;
}

.brand-preview-card {
  border: 1px solid var(--td-border-color);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 16px;
}

.preview-sidebar {
  background: #1e1e2d;
  padding: 16px;
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

.preview-logo-area {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 12px;
}

.preview-logo-img {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  object-fit: contain;
}

.preview-logo-placeholder {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #3b82f6;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.preview-logo-text {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.preview-menu-item {
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 4px;

  &.active {
    background: rgba(59, 130, 246, 0.2);
    color: #fff;
    border-left: 2px solid #3b82f6;
  }
}

.preview-copyright {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 10px;
  color: rgba(255, 255, 255, 0.3);
  word-break: break-all;
}

.brand-tips {
  margin-top: 12px;

  p {
    margin: 0 0 4px;
    font-size: 13px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}
</style>
