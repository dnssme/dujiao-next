<template>
  <div class="min-h-screen theme-page pt-20 pb-16">
    <div class="container mx-auto px-4">
      <div class="max-w-3xl mx-auto mt-12 theme-panel backdrop-blur-xl border rounded-3xl p-8 md:p-12 text-center">
        <p class="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold theme-surface-soft border theme-border theme-text-muted">
          {{ brandSiteName }}
        </p>
        <p class="text-7xl md:text-8xl font-black tracking-tight theme-text-primary">404</p>
        <h1 class="mt-4 text-2xl md:text-4xl font-extrabold theme-text-primary">
          {{ t('notFoundPage.title') }}
        </h1>
        <p class="mt-4 text-base md:text-lg theme-text-secondary leading-relaxed">
          {{ t('notFoundPage.description', { site: brandSiteName }) }}
        </p>

        <div class="mt-8 theme-panel-soft border theme-border rounded-2xl p-4 text-left">
          <p class="text-xs uppercase tracking-widest theme-text-muted mb-3">
            {{ t('notFoundPage.quickLinksTitle') }}
          </p>
          <div class="flex flex-wrap gap-2">
            <RouterLink
              to="/products"
              class="inline-flex items-center px-3 py-2 rounded-lg theme-btn-secondary border text-sm font-medium">
              {{ t('nav.products') }}
            </RouterLink>
            <RouterLink
              to="/notice"
              class="inline-flex items-center px-3 py-2 rounded-lg theme-btn-secondary border text-sm font-medium">
              {{ t('nav.notice') }}
            </RouterLink>
            <RouterLink
              to="/blog"
              class="inline-flex items-center px-3 py-2 rounded-lg theme-btn-secondary border text-sm font-medium">
              {{ t('nav.blog') }}
            </RouterLink>
          </div>
        </div>

        <div class="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
          <RouterLink
            to="/"
            class="inline-flex items-center justify-center px-6 py-3 rounded-xl theme-btn-primary border font-semibold min-w-[180px]">
            {{ t('notFoundPage.backHome') }}
          </RouterLink>
          <button
            type="button"
            class="inline-flex items-center justify-center px-6 py-3 rounded-xl theme-btn-secondary border font-semibold min-w-[180px]"
            @click="goBack">
            {{ t('notFoundPage.backPrevious') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'

const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()

const brandSiteName = computed(() => {
  const siteName = String(appStore.config?.brand?.site_name || '').trim()
  return siteName !== '' ? siteName : 'Dujiao-Next'
})

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push('/')
}
</script>
