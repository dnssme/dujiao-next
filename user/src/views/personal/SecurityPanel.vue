<template>
  <div class="space-y-6">
    <div class="theme-personal-card">
      <div class="mb-6 flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="text-xl font-bold theme-text-primary">{{ t('personalCenter.security.title') }}</h2>
          <p class="mt-1 text-sm theme-text-muted">
            {{ requiresOldEmailCode ? t('personalCenter.security.subtitle') : t('personalCenter.security.subtitleBindOnly') }}
          </p>
        </div>
        <span class="theme-badge theme-badge-accent px-3 py-1 text-xs font-semibold">
          {{ t('personalCenter.tabs.security') }}
        </span>
      </div>

      <div v-if="securityAlert" class="mb-5 rounded-xl border px-4 py-3 text-sm shadow-sm" :class="pageAlertClass(securityAlert.level)">
        {{ securityAlert.message }}
      </div>

      <div class="mb-6 rounded-2xl border border-gray-200/70 bg-gray-50/70 p-4 dark:border-white/10 dark:bg-white/5">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h3 class="text-base font-semibold theme-text-primary">{{ t('personalCenter.security.telegramTitle') }}</h3>
            <p class="mt-1 text-xs theme-text-muted">
              {{ telegramEnabled ? t('personalCenter.security.telegramSubtitle') : t('personalCenter.security.telegramDisabledTip') }}
            </p>
          </div>
          <span class="theme-badge px-3 py-1 text-xs font-semibold" :class="telegramBound ? 'theme-badge-success' : 'theme-badge-warning'">
            {{ telegramBound ? t('personalCenter.security.telegramBound') : t('personalCenter.security.telegramUnbound') }}
          </span>
        </div>

        <div v-if="userProfileStore.loadingTelegramBinding" class="mt-4 rounded-xl border border-dashed border-gray-200/80 px-4 py-4 text-sm theme-text-muted dark:border-white/10">
          {{ t('personalCenter.security.telegramLoading') }}
        </div>

        <div v-else-if="telegramBound" class="mt-4 space-y-4 rounded-xl border border-gray-200/80 bg-white/80 p-4 dark:border-white/10 dark:bg-white/10">
          <div class="flex items-center gap-3">
            <img
              v-if="userProfileStore.telegramBinding?.avatar_url"
              :src="userProfileStore.telegramBinding?.avatar_url"
              alt="Telegram Avatar"
              class="h-11 w-11 rounded-full border border-gray-200 object-cover dark:border-white/10"
            />
            <div>
              <p class="text-sm font-semibold theme-text-primary">{{ telegramDisplayName }}</p>
              <p class="text-xs theme-text-muted">{{ t('personalCenter.security.telegramBindID', { id: userProfileStore.telegramBinding?.provider_user_id || '-' }) }}</p>
            </div>
          </div>
          <p class="text-xs theme-text-muted">
            {{ t('personalCenter.security.telegramBindTime', { time: formatDate(userProfileStore.telegramBinding?.auth_at) || '-' }) }}
          </p>
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-xl border theme-btn-secondary px-4 py-2.5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="userProfileStore.unbindingTelegram || !canUnbindTelegram"
            @click="handleUnbindTelegram"
          >
            {{ userProfileStore.unbindingTelegram ? t('personalCenter.security.telegramUnbinding') : t('personalCenter.security.telegramUnbind') }}
          </button>
          <p v-if="!canUnbindTelegram" class="text-xs theme-text-muted">
            {{ t('personalCenter.security.telegramUnbindDisabledTip') }}
          </p>
        </div>

        <div v-else class="mt-4 space-y-3">
          <p class="text-xs theme-text-muted">
            {{ telegramEnabled ? t('personalCenter.security.telegramUnboundTip') : t('personalCenter.security.telegramDisabledTip') }}
          </p>
          <div v-if="telegramEnabled" ref="telegramWidgetRef" class="flex justify-start"></div>
        </div>
      </div>

      <form class="space-y-6" @submit.prevent="handleChangeEmail">
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.currentEmailLabel') }}</label>
          <input
            :value="currentEmailDisplay"
            disabled
            class="w-full rounded-xl border border-gray-200 bg-gray-100 px-4 py-3 text-gray-500 dark:border-white/10 dark:bg-white/5"
          />
          <p v-if="!requiresOldEmailCode" class="mt-2 text-xs theme-text-muted">
            {{ t('personalCenter.security.bindOnlyTip') }}
          </p>
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.newEmailLabel') }}</label>
          <input
            v-model="securityForm.newEmail"
            type="email"
            :placeholder="t('personalCenter.security.newEmailPlaceholder')"
            class="w-full form-input-lg"
          />
        </div>

        <div class="grid grid-cols-1 gap-4" :class="requiresOldEmailCode ? 'lg:grid-cols-2' : ''">
          <div v-if="requiresOldEmailCode">
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.oldCodeLabel') }}</label>
            <div class="flex flex-col gap-2 sm:flex-row">
              <input
                v-model="securityForm.oldCode"
                :placeholder="t('personalCenter.security.codePlaceholder')"
                class="min-w-0 flex-1 form-input-lg"
              />
              <button
                type="button"
                @click="handleSendOldCode"
                :disabled="userProfileStore.sendingCode || oldCodeCooldown > 0"
                class="whitespace-nowrap rounded-xl border theme-btn-secondary px-4 py-3 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-60"
              >
                {{ oldCodeCooldown > 0 ? t('personalCenter.security.countdown', { seconds: oldCodeCooldown }) : t('personalCenter.security.sendOldCode') }}
              </button>
            </div>
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.newCodeLabel') }}</label>
            <div class="flex flex-col gap-2 sm:flex-row">
              <input
                v-model="securityForm.newCode"
                :placeholder="t('personalCenter.security.codePlaceholder')"
                class="min-w-0 flex-1 form-input-lg"
              />
              <button
                type="button"
                @click="handleSendNewCode"
                :disabled="userProfileStore.sendingCode || newCodeCooldown > 0"
                class="whitespace-nowrap rounded-xl border theme-btn-secondary px-4 py-3 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-60"
              >
                {{ newCodeCooldown > 0 ? t('personalCenter.security.countdown', { seconds: newCodeCooldown }) : t('personalCenter.security.sendNewCode') }}
              </button>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-200/70 pt-5 dark:border-white/10">
          <button
            type="submit"
            :disabled="userProfileStore.changingEmail"
            class="inline-flex items-center justify-center rounded-xl theme-btn-primary px-6 py-3 text-sm font-bold transition-colors disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{
              userProfileStore.changingEmail
                ? t('personalCenter.security.submitting')
                : (requiresOldEmailCode ? t('personalCenter.security.submit') : t('personalCenter.security.bindSubmit'))
            }}
          </button>
        </div>
      </form>
    </div>

    <div class="theme-personal-card">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-bold theme-text-primary">{{ t('personalCenter.security.loginLogsTitle') }}</h3>
        <span class="text-xs theme-text-muted">{{ t('personalCenter.security.loginLogsTip') }}</span>
      </div>
      <div v-if="userProfileStore.loadingLoginLogs" class="rounded-xl border border-gray-200/70 px-4 py-6 text-center text-sm text-gray-500 dark:border-white/10 dark:text-gray-400">
        {{ t('personalCenter.security.loginLogsLoading') }}
      </div>
      <div v-else-if="userProfileStore.recentLoginLogs.length === 0" class="rounded-xl border border-dashed border-gray-200/80 px-4 py-6 text-center text-sm text-gray-500 dark:border-white/10 dark:text-gray-400">
        {{ t('personalCenter.security.loginLogsEmpty') }}
      </div>
      <div v-else class="overflow-x-auto rounded-xl border border-gray-200/70 dark:border-white/10">
        <table class="min-w-full divide-y divide-gray-200 text-left text-sm dark:divide-white/10">
          <thead class="bg-gray-50/80 text-xs uppercase tracking-wide text-gray-500 dark:bg-white/5 dark:text-gray-400">
            <tr>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.security.loginLogsTime') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.security.loginLogsStatus') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.security.loginLogsIp') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.security.loginLogsReason') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-white/10">
            <tr v-for="item in userProfileStore.recentLoginLogs" :key="item.id">
              <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ formatDate(item.created_at) }}</td>
              <td class="px-4 py-3">
                <span class="theme-badge px-2.5 py-1 text-xs font-semibold" :class="loginStatusClass(item.status)">
                  {{ loginStatusLabel(item.status) }}
                </span>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-gray-600 dark:text-gray-300">{{ item.client_ip || '-' }}</td>
              <td class="px-4 py-3 text-xs theme-text-muted">{{ loginReasonLabel(item.fail_reason) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="canManagePassword" class="theme-personal-card">
      <h3 class="text-lg font-bold theme-text-primary">
        {{ requiresOldPassword ? t('personalCenter.security.passwordTitle') : t('personalCenter.security.setPasswordTitle') }}
      </h3>
      <p class="mt-1 text-sm theme-text-muted">
        {{ requiresOldPassword ? t('personalCenter.security.passwordSubtitle') : t('personalCenter.security.setPasswordSubtitle') }}
      </p>

      <form class="mt-6 space-y-6" @submit.prevent="handleChangePassword">
        <div v-if="requiresOldPassword">
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.currentPasswordLabel') }}</label>
          <input
            v-model="passwordForm.oldPassword"
            type="password"
            :placeholder="t('personalCenter.security.passwordPlaceholder')"
            class="w-full form-input-lg"
          />
        </div>

        <div class="grid grid-cols-1 gap-5 md:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.newPasswordLabel') }}</label>
            <input
              v-model="passwordForm.newPassword"
              type="password"
              :placeholder="t('personalCenter.security.passwordPlaceholder')"
              class="w-full form-input-lg"
            />
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.security.confirmPasswordLabel') }}</label>
            <input
              v-model="passwordForm.confirmPassword"
              type="password"
              :placeholder="t('personalCenter.security.passwordPlaceholder')"
              class="w-full form-input-lg"
            />
          </div>
        </div>

        <div class="border-t border-gray-200/70 pt-5 dark:border-white/10">
          <button
            type="submit"
            :disabled="userProfileStore.changingPassword"
            class="inline-flex items-center justify-center rounded-xl border theme-btn-secondary px-6 py-3 text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{
              userProfileStore.changingPassword
                ? (requiresOldPassword ? t('personalCenter.security.changePasswordSubmitting') : t('personalCenter.security.setPasswordSubmitting'))
                : (requiresOldPassword ? t('personalCenter.security.changePassword') : t('personalCenter.security.setPassword'))
            }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { pageAlertClass, type PageAlert } from '../../utils/alerts'
import type { TelegramAuthPayload } from '../../api'
import { useAppStore } from '../../stores/app'
import { useUserProfileStore } from '../../stores/userProfile'
import { useUserAuthStore } from '../../stores/userAuth'

const { t } = useI18n()
const appStore = useAppStore()
const userProfileStore = useUserProfileStore()
const userAuthStore = useUserAuthStore()

const securityForm = reactive({
  newEmail: '',
  oldCode: '',
  newCode: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const securityAlert = ref<PageAlert | null>(null)
const oldCodeCooldown = ref(0)
const newCodeCooldown = ref(0)
const telegramWidgetRef = ref<HTMLDivElement | null>(null)
let cooldownTimer: number | null = null
const telegramCallbackName = '__dujiaoSecurityTelegramBind'

const telegramConfig = computed(() => appStore.config?.telegram_auth || null)
const telegramBotUsername = computed(() => String(telegramConfig.value?.bot_username || '').trim())
const telegramEnabled = computed(() => !!telegramConfig.value?.enabled && telegramBotUsername.value !== '')
const telegramBound = computed(() => !!userProfileStore.telegramBinding?.bound)
const emailChangeMode = computed(() => userProfileStore.profile?.email_change_mode || 'change_with_old_and_new')
const requiresOldEmailCode = computed(() => emailChangeMode.value !== 'bind_only')
const canManagePassword = computed(() => requiresOldEmailCode.value)
const passwordChangeMode = computed(() => userProfileStore.profile?.password_change_mode || 'change_with_old')
const requiresOldPassword = computed(() => passwordChangeMode.value !== 'set_without_old')
const canUnbindTelegram = computed(() => requiresOldEmailCode.value)
const currentEmailDisplay = computed(() => {
  if (!requiresOldEmailCode.value) {
    return t('personalCenter.security.bindOnlyEmailDisplay')
  }
  return userProfileStore.profile?.email || ''
})
const telegramDisplayName = computed(() => {
  if (userProfileStore.telegramBinding?.username) {
    return `@${userProfileStore.telegramBinding.username}`
  }
  return t('personalCenter.security.telegramDisplayFallback')
})

const startCooldown = (kind: 'old' | 'new') => {
  if (kind === 'old') {
    oldCodeCooldown.value = 60
  } else {
    newCodeCooldown.value = 60
  }
  if (cooldownTimer !== null) return
  cooldownTimer = window.setInterval(() => {
    if (oldCodeCooldown.value > 0) {
      oldCodeCooldown.value -= 1
    }
    if (newCodeCooldown.value > 0) {
      newCodeCooldown.value -= 1
    }
    if (oldCodeCooldown.value === 0 && newCodeCooldown.value === 0 && cooldownTimer !== null) {
      window.clearInterval(cooldownTimer)
      cooldownTimer = null
    }
  }, 1000)
}

const handleSendOldCode = async () => {
  securityAlert.value = null
  if (!requiresOldEmailCode.value) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.bindOnlyOldCodeDisabled'),
    }
    return
  }
  const ok = await userProfileStore.sendChangeEmailCode({ kind: 'old' })
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.sendCodeFailed'),
    }
    return
  }
  startCooldown('old')
  securityAlert.value = {
    level: 'success',
    message: t('personalCenter.security.sendOldCodeSuccess'),
  }
}

const handleSendNewCode = async () => {
  securityAlert.value = null
  const newEmail = securityForm.newEmail.trim()
  if (!newEmail) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.newEmailRequired'),
    }
    return
  }
  const ok = await userProfileStore.sendChangeEmailCode({ kind: 'new', new_email: newEmail })
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.sendCodeFailed'),
    }
    return
  }
  startCooldown('new')
  securityAlert.value = {
    level: 'success',
    message: t('personalCenter.security.sendNewCodeSuccess'),
  }
}

const handleChangeEmail = async () => {
  securityAlert.value = null
  const requiresOldCode = requiresOldEmailCode.value
  const oldCode = securityForm.oldCode.trim()
  const payload = {
    new_email: securityForm.newEmail.trim(),
    new_code: securityForm.newCode.trim(),
    ...(requiresOldCode ? { old_code: oldCode } : {}),
  }
  if (!payload.new_email || !payload.new_code || (requiresOldCode && !oldCode)) {
    securityAlert.value = {
      level: 'warning',
      message: requiresOldCode
        ? t('personalCenter.security.changeEmailRequired')
        : t('personalCenter.security.bindEmailRequired'),
    }
    return
  }

  const ok = await userProfileStore.changeEmail(payload)
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.changeEmailFailed'),
    }
    return
  }

  securityForm.newEmail = ''
  securityForm.oldCode = ''
  securityForm.newCode = ''
  oldCodeCooldown.value = 0
  newCodeCooldown.value = 0
  securityAlert.value = {
    level: 'success',
    message: requiresOldCode
      ? t('personalCenter.security.changeEmailSuccess')
      : t('personalCenter.security.bindEmailSuccess'),
  }
}

const handleChangePassword = async () => {
  securityAlert.value = null
  const oldPassword = passwordForm.oldPassword.trim()
  const newPassword = passwordForm.newPassword.trim()
  const confirmPassword = passwordForm.confirmPassword.trim()
  const needOldPassword = requiresOldPassword.value

  if (!newPassword || !confirmPassword || (needOldPassword && !oldPassword)) {
    securityAlert.value = {
      level: 'warning',
      message: needOldPassword
        ? t('personalCenter.security.changePasswordRequired')
        : t('personalCenter.security.setPasswordRequired'),
    }
    return
  }

  if (newPassword !== confirmPassword) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.passwordMismatch'),
    }
    return
  }

  const payload = {
    ...(needOldPassword ? { old_password: oldPassword } : {}),
    new_password: newPassword,
  }
  const ok = await userProfileStore.changePassword(payload)

  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.changePasswordFailed'),
    }
    return
  }

  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  securityAlert.value = {
    level: 'success',
    message: needOldPassword
      ? t('personalCenter.security.changePasswordSuccess')
      : t('personalCenter.security.setPasswordSuccess'),
  }

  userAuthStore.logout('/auth/login?reason=password_changed')
}

const buildTelegramPayload = (raw: any): TelegramAuthPayload | null => {
  const id = Number(raw?.id)
  const authDate = Number(raw?.auth_date)
  const hash = String(raw?.hash || '').trim()
  if (!Number.isFinite(id) || id <= 0 || !Number.isFinite(authDate) || authDate <= 0 || hash === '') {
    return null
  }
  return {
    id,
    first_name: String(raw?.first_name || '').trim(),
    last_name: String(raw?.last_name || '').trim(),
    username: String(raw?.username || '').trim(),
    photo_url: String(raw?.photo_url || '').trim(),
    auth_date: authDate,
    hash,
  }
}

const clearTelegramWidget = () => {
  if (telegramWidgetRef.value) {
    telegramWidgetRef.value.innerHTML = ''
  }
}

const renderTelegramWidget = () => {
  if (!telegramEnabled.value || telegramBound.value || !telegramWidgetRef.value) {
    clearTelegramWidget()
    return
  }
  clearTelegramWidget()
  const script = document.createElement('script')
  script.async = true
  script.src = 'https://telegram.org/js/telegram-widget.js?22'
  script.setAttribute('data-telegram-login', telegramBotUsername.value)
  script.setAttribute('data-size', 'large')
  script.setAttribute('data-userpic', 'false')
  script.setAttribute('data-request-access', 'write')
  script.setAttribute('data-onauth', `${telegramCallbackName}(user)`)
  script.onerror = () => {
    securityAlert.value = {
      level: 'error',
      message: t('personalCenter.security.telegramWidgetLoadFailed'),
    }
  }
  telegramWidgetRef.value.appendChild(script)
}

const handleTelegramBind = async (raw: any) => {
  securityAlert.value = null
  const payload = buildTelegramPayload(raw)
  if (!payload) {
    securityAlert.value = {
      level: 'warning',
      message: t('personalCenter.security.telegramInvalidPayload'),
    }
    return
  }
  const ok = await userProfileStore.bindTelegram(payload)
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.telegramBindFailed'),
    }
    return
  }
  securityAlert.value = {
    level: 'success',
    message: t('personalCenter.security.telegramBindSuccess'),
  }
  renderTelegramWidget()
}

const handleUnbindTelegram = async () => {
  securityAlert.value = null
  const ok = await userProfileStore.unbindTelegram()
  if (!ok) {
    securityAlert.value = {
      level: 'error',
      message: userProfileStore.securityError || t('personalCenter.security.telegramUnbindFailed'),
    }
    return
  }
  securityAlert.value = {
    level: 'success',
    message: t('personalCenter.security.telegramUnbindSuccess'),
  }
  renderTelegramWidget()
}

const loginStatusLabel = (status?: string) => {
  const normalized = (status || '').trim() || 'failed'
  return t(`personalCenter.security.loginLogsStatusMap.${normalized}`)
}

const loginStatusClass = (status?: string) => {
  if ((status || '').trim() === 'success') {
    return 'theme-badge-success'
  }
  return 'theme-badge-danger'
}

const loginReasonLabel = (reason?: string) => {
  const normalized = (reason || '').trim()
  if (!normalized) return '-'
  const key = `personalCenter.security.loginLogsReasonMap.${normalized}`
  const translated = t(key)
  return translated === key ? normalized : translated
}

const formatDate = (raw?: string | null) => {
  if (!raw) return ''
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

onMounted(async () => {
  await Promise.all([
    appStore.loadConfig(),
    userProfileStore.loadRecentLoginLogs(10),
    userProfileStore.loadTelegramBinding(),
  ])
  const win = window as Window & Record<string, any>
  win[telegramCallbackName] = handleTelegramBind
  renderTelegramWidget()
})

onUnmounted(() => {
  const win = window as Window & Record<string, any>
  delete win[telegramCallbackName]
  clearTelegramWidget()
  if (cooldownTimer !== null) {
    window.clearInterval(cooldownTimer)
    cooldownTimer = null
  }
})

watch([telegramEnabled, telegramBotUsername, telegramBound], () => {
  renderTelegramWidget()
})
</script>
