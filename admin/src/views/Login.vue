<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAdminAuthStore } from '@/stores/auth'
import { adminAPI, type CaptchaPayload } from '@/api/admin'
import ImageCaptcha from '@/components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '@/components/captcha/TurnstileCaptcha.vue'

const { t } = useI18n()
const router = useRouter()
const authStore = useAdminAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loadingCaptcha = ref(false)
const captchaConfig = ref<any>(null)
const captchaPayload = ref<CaptchaPayload>({})
const turnstileToken = ref('')
const imageCaptchaRef = ref<InstanceType<typeof ImageCaptcha> | null>(null)
const turnstileRef = ref<InstanceType<typeof TurnstileCaptcha> | null>(null)

const captchaProvider = computed(() => String(captchaConfig.value?.provider || 'none'))
const loginCaptchaEnabled = computed(() => {
  const loginScene = !!captchaConfig.value?.scenes?.login
  return loginScene && captchaProvider.value !== 'none'
})
const turnstileSiteKey = computed(() => String(captchaConfig.value?.turnstile?.site_key || ''))

const getCaptchaPayload = (): CaptchaPayload | undefined => {
  if (!loginCaptchaEnabled.value) return undefined
  if (captchaProvider.value === 'image') {
    return {
      captcha_id: captchaPayload.value.captcha_id || '',
      captcha_code: captchaPayload.value.captcha_code || '',
    }
  }
  if (captchaProvider.value === 'turnstile') {
    return {
      turnstile_token: turnstileToken.value,
    }
  }
  return undefined
}

const submit = async () => {
  error.value = ''

  if (loginCaptchaEnabled.value && captchaProvider.value === 'image') {
    if (!captchaPayload.value.captcha_id || !captchaPayload.value.captcha_code) {
      error.value = t('admin.login.captchaRequired')
      return
    }
  }

  if (loginCaptchaEnabled.value && captchaProvider.value === 'turnstile') {
    if (!turnstileToken.value) {
      error.value = t('admin.login.captchaRequired')
      return
    }
  }

  try {
    await authStore.login({
      username: username.value.trim(),
      password: password.value,
      captcha_payload: getCaptchaPayload(),
    })
    router.push('/')
  } catch (err: any) {
    error.value = err?.message || t('admin.login.errors.invalidCredentials')
    if (captchaProvider.value === 'image') {
      imageCaptchaRef.value?.refresh()
    }
    if (captchaProvider.value === 'turnstile') {
      turnstileRef.value?.reset()
      turnstileToken.value = ''
    }
  }
}

const loadCaptchaConfig = async () => {
  loadingCaptcha.value = true
  try {
    const res = await adminAPI.getPublicConfig()
    const payload = res.data?.data as any
    captchaConfig.value = payload?.captcha || null
  } catch {
    captchaConfig.value = null
  } finally {
    loadingCaptcha.value = false
  }
}

onMounted(() => {
  loadCaptchaConfig()
})
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex items-center justify-center px-6">
    <div class="w-full max-w-md">
      <Card class="border-border shadow-sm">
        <CardHeader>
          <CardTitle class="text-2xl">{{ t('admin.login.title') }}</CardTitle>
          <p class="text-sm text-muted-foreground mt-1">{{ t('admin.login.subtitle') }}</p>
        </CardHeader>
        <CardContent>
          <form class="space-y-4" @submit.prevent="submit">
            <div class="space-y-2">
              <Label for="username">{{ t('admin.login.username') }}</Label>
              <Input id="username" v-model="username" :placeholder="t('admin.login.username')" />
            </div>
            <div class="space-y-2">
              <Label for="password">{{ t('admin.login.password') }}</Label>
              <Input id="password" v-model="password" type="password" :placeholder="t('admin.login.password')" />
            </div>

            <div v-if="loginCaptchaEnabled" class="space-y-2">
              <Label>{{ t('admin.login.captchaLabel') }}</Label>
              <ImageCaptcha
                v-if="captchaProvider === 'image'"
                ref="imageCaptchaRef"
                v-model="captchaPayload"
                :disabled="authStore.loading || loadingCaptcha"
              />
              <TurnstileCaptcha
                v-else-if="captchaProvider === 'turnstile'"
                ref="turnstileRef"
                v-model="turnstileToken"
                :site-key="turnstileSiteKey"
              />
            </div>

            <div v-if="error" class="text-sm text-destructive">{{ error }}</div>
            <Button type="submit" class="w-full" :disabled="authStore.loading || loadingCaptcha">
              {{ authStore.loading ? t('admin.login.submitting') : t('admin.login.submit') }}
            </Button>
          </form>
        </CardContent>
      </Card>
      <p class="mt-4 flex items-center justify-center gap-1 text-center text-xs text-muted-foreground">
        <span>© {{ new Date().getFullYear() }} Dujiao-Next ·</span>
        <a
          href="https://github.com/dujiao-next"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-1 underline-offset-2 hover:underline"
        >
          <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M12 .5C5.648.5.5 5.648.5 12c0 5.084 3.292 9.4 7.86 10.922.575.106.784-.25.784-.556 0-.273-.01-1-.016-1.962-3.197.694-3.872-1.54-3.872-1.54-.522-1.326-1.274-1.678-1.274-1.678-1.042-.713.079-.699.079-.699 1.152.081 1.758 1.183 1.758 1.183 1.024 1.755 2.688 1.248 3.343.954.104-.742.401-1.248.73-1.535-2.552-.29-5.236-1.276-5.236-5.678 0-1.254.448-2.28 1.182-3.084-.118-.29-.512-1.457.112-3.04 0 0 .964-.308 3.158 1.178a10.98 10.98 0 0 1 2.876-.387c.976.004 1.96.132 2.878.387 2.192-1.486 3.154-1.178 3.154-1.178.626 1.583.232 2.75.114 3.04.736.804 1.18 1.83 1.18 3.084 0 4.413-2.688 5.384-5.248 5.668.412.354.78 1.052.78 2.12 0 1.53-.014 2.764-.014 3.14 0 .31.206.668.79.554C20.212 21.396 23.5 17.083 23.5 12 23.5 5.648 18.352.5 12 .5Z" />
          </svg>
          <span>https://github.com/dujiao-next</span>
        </a>
      </p>
    </div>
  </div>
</template>
