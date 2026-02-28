<template>
  <div class="space-y-6">
    <div class="theme-personal-card">
      <div class="mb-6 flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="text-xl font-bold theme-text-primary">{{ t('personalCenter.wallet.title') }}</h2>
          <p class="mt-1 text-sm theme-text-muted">{{ t('personalCenter.wallet.subtitle') }}</p>
        </div>
        <span class="theme-badge theme-badge-accent px-3 py-1 text-xs font-semibold">
          {{ t('personalCenter.tabs.wallet') }}
        </span>
      </div>

      <div v-if="walletAlert" class="mb-5 rounded-xl border px-4 py-3 text-sm shadow-sm" :class="pageAlertClass(walletAlert.level)">
        {{ walletAlert.message }}
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.balanceLabel') }}</div>
          <div class="mt-2 text-lg font-bold theme-text-primary">{{ balanceDisplay }}</div>
        </div>
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.transactionsLabel') }}</div>
          <div class="mt-2 text-lg font-bold theme-text-primary">{{ pagination.total }}</div>
        </div>
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.currentPageLabel') }}</div>
          <div class="mt-2 text-lg font-bold theme-text-primary">
            {{ t('orders.pageInfo', { page: pagination.page, total: pagination.total_page }) }}
          </div>
        </div>
      </div>
    </div>

    <div class="theme-personal-card">
      <h3 class="text-lg font-bold theme-text-primary">{{ t('personalCenter.wallet.rechargeTitle') }}</h3>
      <p class="mt-1 text-sm theme-text-muted">{{ t('personalCenter.wallet.rechargeSubtitle') }}</p>
      <form class="mt-6 grid grid-cols-1 gap-4 md:grid-cols-[1fr_1fr_2fr_auto]" @submit.prevent="handleRecharge">
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.wallet.amountLabel') }}</label>
          <input
            v-model.trim="rechargeForm.amount"
            type="text"
            inputmode="decimal"
            :placeholder="t('personalCenter.wallet.amountPlaceholder')"
            class="w-full form-input-lg"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.wallet.channelLabel') }}</label>
          <select
            v-model.number="rechargeForm.channelId"
            class="w-full form-input-lg"
            :disabled="!hasChannels || recharging"
          >
            <option :value="0">{{ t('personalCenter.wallet.channelPlaceholder') }}</option>
            <option v-for="channel in channels" :key="channel.id" :value="channel.id">
              {{ channel.name }}
            </option>
          </select>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('personalCenter.wallet.remarkLabel') }}</label>
          <input
            v-model.trim="rechargeForm.remark"
            type="text"
            :placeholder="t('personalCenter.wallet.remarkPlaceholder')"
            class="w-full form-input-lg"
          />
        </div>
        <div class="flex items-end">
          <button
            type="submit"
            :disabled="recharging || !hasChannels"
            class="inline-flex h-11 w-full items-center justify-center rounded-xl theme-btn-primary px-5 text-sm font-bold transition-colors disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{ recharging ? t('personalCenter.wallet.recharging') : t('personalCenter.wallet.rechargeSubmit') }}
          </button>
        </div>
      </form>
      <p v-if="!hasChannels" class="mt-3 text-xs text-amber-600">
        {{ t('payment.channelEmpty') }}
      </p>
    </div>

    <div v-if="currentRecharge && currentRechargePayment" class="theme-personal-card">
      <h3 class="text-lg font-bold theme-text-primary">{{ t('personalCenter.wallet.paymentInfoTitle') }}</h3>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3 text-sm">
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.rechargeNoLabel') }}</div>
          <div class="mt-1 font-mono theme-text-primary">{{ currentRecharge.recharge_no }}</div>
        </div>
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.paymentStatusLabel') }}</div>
          <div class="mt-1">
            <span class="theme-badge px-2.5 py-1 text-xs font-semibold" :class="rechargeStatusClass(currentRecharge.status)">
              {{ rechargeStatusLabel(currentRecharge.status) }}
            </span>
          </div>
        </div>
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.payAmountLabel') }}</div>
          <div class="mt-1 font-mono theme-text-primary">{{ formatMoney(currentRecharge.payable_amount, currentRecharge.currency) }}</div>
        </div>
      </div>
      <div v-if="isRechargePending" class="mt-3 text-xs theme-text-muted">
        {{ t('personalCenter.wallet.pendingHint') }}
      </div>
      <div class="mt-5 grid grid-cols-1 gap-6 md:grid-cols-2">
        <div v-if="qrImageUrl" class="rounded-xl border theme-surface-soft p-4">
          <div class="mb-3 text-sm font-semibold theme-text-primary">{{ t('payment.qrTitle') }}</div>
          <div class="flex items-center justify-center">
            <img :src="qrImageUrl" alt="Recharge QR" class="h-52 w-52 object-contain" />
          </div>
          <div v-if="qrUsingPayLinkFallback" class="mt-3 text-xs theme-text-muted">
            {{ t('payment.qrFallbackHint') }}
          </div>
          <div class="mt-3 text-xs theme-text-muted break-all">{{ qrDisplayContent }}</div>
        </div>
        <div class="rounded-xl border theme-surface-soft p-4">
          <div class="text-xs theme-text-muted">{{ t('personalCenter.wallet.paymentChannelLabel') }}</div>
          <div class="mt-1 text-sm font-semibold theme-text-primary">{{ currentChannelName }}</div>
          <div class="mt-4 flex flex-wrap items-center gap-3">
            <a
              v-if="payLink"
              :href="payLink"
              target="_blank"
              rel="noopener"
              class="inline-flex items-center rounded-lg border theme-btn-secondary px-3 py-1.5 text-xs font-semibold"
            >
              {{ t('payment.openPayLink') }}
            </a>
            <button
              type="button"
              class="inline-flex items-center rounded-lg border theme-btn-secondary px-3 py-1.5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="checkingPayment"
              @click="checkRechargePayment"
            >
              {{ checkingPayment ? t('personalCenter.wallet.checkingPayStatus') : t('personalCenter.wallet.checkPayStatus') }}
            </button>
          </div>
          <div v-if="payLink" class="mt-3 text-xs theme-text-muted break-all">
            {{ t('payment.payLinkLabel') }}：{{ payLink }}
          </div>
        </div>
      </div>
    </div>

    <div class="theme-personal-card">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-bold theme-text-primary">{{ t('personalCenter.wallet.detailTitle') }}</h3>
        <button
          type="button"
          class="inline-flex items-center rounded-lg border theme-btn-secondary px-3 py-1.5 text-xs font-semibold"
          @click="refreshCurrentPage"
        >
          {{ t('orders.filters.refresh') }}
        </button>
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="idx in 3" :key="idx" class="h-16 animate-pulse rounded-xl border theme-surface-muted"></div>
      </div>
      <div v-else-if="transactions.length === 0" class="rounded-xl border border-dashed theme-surface-soft px-4 py-6 text-sm theme-text-muted">
        {{ t('personalCenter.wallet.empty') }}
      </div>
      <div v-else class="overflow-x-auto rounded-xl border border-gray-200/70 dark:border-white/10">
        <table class="min-w-full divide-y divide-gray-200 text-left text-sm dark:divide-white/10">
          <thead class="bg-gray-50/80 text-xs uppercase tracking-wide text-gray-500 dark:bg-white/5 dark:text-gray-400">
            <tr>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.wallet.table.createdAt') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.wallet.table.type') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.wallet.table.direction') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.wallet.table.amount') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.wallet.table.balanceAfter') }}</th>
              <th class="px-4 py-3 font-semibold">{{ t('personalCenter.wallet.table.remark') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-white/10">
            <tr v-for="item in transactions" :key="item.id">
              <td class="px-4 py-3 text-xs theme-text-muted">{{ formatDate(item.created_at) }}</td>
              <td class="px-4 py-3 text-xs theme-text-secondary">{{ transactionTypeLabel(item.type) }}</td>
              <td class="px-4 py-3 text-xs">
                <span class="theme-badge px-2.5 py-1 text-xs font-semibold" :class="directionClass(item.direction)">
                  {{ directionLabel(item.direction) }}
                </span>
              </td>
              <td class="px-4 py-3 font-mono text-sm" :class="item.direction === 'in' ? 'text-emerald-500' : 'text-rose-500'">
                {{ signedAmount(item.direction, item.amount, item.currency) }}
              </td>
              <td class="px-4 py-3 font-mono text-sm theme-text-primary">
                {{ formatMoney(item.balance_after, item.currency) }}
              </td>
              <td class="px-4 py-3 text-xs theme-text-muted">{{ item.remark || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="pagination.total_page > 1" class="mt-6 flex flex-wrap items-center justify-center gap-3">
        <button
          :disabled="pagination.page <= 1"
          class="rounded-lg border theme-btn-secondary px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-40"
          @click="changePage(pagination.page - 1)"
        >
          {{ t('orders.prevPage') }}
        </button>
        <span class="rounded-full border theme-pill-neutral px-4 py-2 text-sm">
          {{ t('orders.pageInfo', { page: pagination.page, total: pagination.total_page }) }}
        </span>
        <button
          :disabled="pagination.page >= pagination.total_page"
          class="rounded-lg border theme-btn-secondary px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-40"
          @click="changePage(pagination.page + 1)"
        >
          {{ t('orders.nextPage') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { walletAPI } from '../../api'
import { useAppStore } from '../../stores/app'
import { pageAlertClass, type PageAlert } from '../../utils/alerts'
import { amountToCents } from '../../utils/money'
import QRCode from 'qrcode'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const recharging = ref(false)
const checkingPayment = ref(false)
const wallet = ref<any>(null)
const transactions = ref<any[]>([])
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})
const walletAlert = ref<PageAlert | null>(null)
const pollTimer = ref<number | null>(null)

const rechargeForm = reactive({
  amount: '',
  channelId: 0,
  remark: '',
})

const currentRecharge = ref<any>(null)
const currentRechargePayment = ref<any>(null)

const channels = computed(() => {
  const list = appStore.config?.payment_channels
  if (!Array.isArray(list)) return []
  return list.filter((channel: any) => {
    const providerType = String(channel?.provider_type || '').toLowerCase()
    const channelType = String(channel?.channel_type || '').toLowerCase()
    if (providerType === 'epay') {
      return ['wechat', 'wxpay', 'alipay', 'qqpay'].includes(channelType)
    }
    return true
  }).map((channel: any) => ({
    id: Number(channel.id),
    name: String(channel.name || channel.channel_type || channel.id),
    channel_type: String(channel.channel_type || ''),
  })).filter((channel: any) => Number.isFinite(channel.id) && channel.id > 0)
})
const hasChannels = computed(() => channels.value.length > 0)

const formatMoney = (amount?: string, currency?: string) => {
  if (amount === null || amount === undefined || amount === '') return '-'
  if (currency === null || currency === undefined || currency === '') {
    return String(amount)
  }
  return `${amount} ${currency}`
}

const balanceDisplay = computed(() => formatMoney(wallet.value?.balance, String(appStore.config?.currency || 'CNY')))
const payLink = computed(() => String(currentRechargePayment.value?.pay_url || '').trim())
const interactionMode = computed(() => String(currentRechargePayment.value?.interaction_mode || '').toLowerCase())
const qrCodeContent = computed(() => String(currentRechargePayment.value?.qr_code || '').trim())
const qrFallbackContent = computed(() => {
  if (interactionMode.value === 'redirect') return ''
  if (qrCodeContent.value) return ''
  return payLink.value
})
const qrDisplayContent = computed(() => qrCodeContent.value || qrFallbackContent.value)
const qrUsingPayLinkFallback = computed(() => Boolean(!qrCodeContent.value && qrFallbackContent.value))
const qrImageUrl = ref('')
const qrRenderVersion = ref(0)

const renderQRCodeImage = async () => {
  const qr = qrDisplayContent.value
  const currentVersion = qrRenderVersion.value + 1
  qrRenderVersion.value = currentVersion
  if (!qr) {
    qrImageUrl.value = ''
    return
  }
  if (qr.startsWith('data:image/')) {
    qrImageUrl.value = qr
    return
  }
  const isImageURL = /^https?:\/\/.+\.(png|jpe?g|gif|webp|svg)(\?.*)?$/i.test(qr)
  if (isImageURL) {
    qrImageUrl.value = qr
    return
  }
  try {
    const dataURL = await QRCode.toDataURL(qr, {
      width: 240,
      margin: 1,
      errorCorrectionLevel: 'M',
    })
    if (currentVersion !== qrRenderVersion.value) return
    qrImageUrl.value = dataURL
  } catch {
    if (currentVersion !== qrRenderVersion.value) return
    qrImageUrl.value = ''
  }
}
const isRechargePending = computed(() => {
  const status = String(currentRecharge.value?.status || '').toLowerCase()
  return status === 'pending' || status === 'initiated'
})

const currentChannelName = computed(() => {
  const channelID = Number(currentRecharge.value?.channel_id || currentRechargePayment.value?.channel_id || 0)
  if (!Number.isFinite(channelID) || channelID <= 0) return '-'
  const channel = channels.value.find((item: any) => item.id === channelID)
  if (channel) return channel.name
  return String(currentRechargePayment.value?.channel_type || '-')
})

const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

const directionLabel = (direction?: string) => {
  if (direction === 'in') return t('personalCenter.wallet.directionIn')
  if (direction === 'out') return t('personalCenter.wallet.directionOut')
  return direction || '-'
}

const directionClass = (direction?: string) => {
  if (direction === 'in') return 'theme-badge-success'
  if (direction === 'out') return 'theme-badge-danger'
  return 'theme-badge-warning'
}

const transactionTypeLabel = (type?: string) => {
  const key = `personalCenter.wallet.types.${type || ''}`
  const translated = t(key)
  if (translated === key) return type || '-'
  return translated
}

const signedAmount = (direction: string, amount?: string, currency?: string) => {
  const base = formatMoney(amount, currency)
  if (base === '-') return base
  if (direction === 'out') return `-${base}`
  return `+${base}`
}

const rechargeStatusLabel = (status?: string) => {
  const normalized = String(status || '').toLowerCase()
  const key = `personalCenter.wallet.rechargeStatus.${normalized}`
  const translated = t(key)
  if (translated === key) return normalized || '-'
  return translated
}

const rechargeStatusClass = (status?: string) => {
  const normalized = String(status || '').toLowerCase()
  if (normalized === 'success') return 'theme-badge-success'
  if (normalized === 'failed' || normalized === 'expired') return 'theme-badge-danger'
  return 'theme-badge-warning'
}

const loadWallet = async () => {
  const response = await walletAPI.account()
  wallet.value = response.data.data
}

const loadTransactions = async (page = 1) => {
  loading.value = true
  try {
    const response = await walletAPI.transactions({
      page,
      page_size: pagination.value.page_size,
    })
    transactions.value = response.data.data || []
    pagination.value = response.data.pagination || pagination.value
  } catch {
    transactions.value = []
  } finally {
    loading.value = false
  }
}

const stopPolling = () => {
  if (!pollTimer.value) return
  window.clearInterval(pollTimer.value)
  pollTimer.value = null
}

const startPolling = () => {
  if (!isRechargePending.value || pollTimer.value) return
  pollTimer.value = window.setInterval(() => {
    void refreshRechargeStatus(true)
  }, 5000)
}

const syncRechargePayload = (payload: any) => {
  currentRecharge.value = payload?.recharge || currentRecharge.value
  currentRechargePayment.value = payload?.payment || currentRechargePayment.value
  if (payload?.account) {
    wallet.value = payload.account
  }
}

const initialize = async () => {
  walletAlert.value = null
  try {
    if (!appStore.config) {
      await appStore.loadConfig()
    }
    await Promise.all([
      loadWallet(),
      loadTransactions(),
    ])
  } catch (err: any) {
    walletAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.wallet.errors.loadFailed'),
    }
  }
}

const handleRecharge = async () => {
  walletAlert.value = null
  const amount = rechargeForm.amount.trim()
  const amountCents = amountToCents(amount)
  if (!amount || amountCents === null || amountCents <= 0) {
    walletAlert.value = {
      level: 'warning',
      message: t('personalCenter.wallet.errors.invalidAmount'),
    }
    return
  }
  if (!Number.isFinite(rechargeForm.channelId) || rechargeForm.channelId <= 0) {
    walletAlert.value = {
      level: 'warning',
      message: t('personalCenter.wallet.errors.channelRequired'),
    }
    return
  }

  recharging.value = true
  try {
    const response = await walletAPI.recharge({
      amount,
      channel_id: rechargeForm.channelId,
      remark: rechargeForm.remark.trim() || undefined,
    })
    const payload = response.data.data || {}
    syncRechargePayload(payload)
    rechargeForm.amount = ''
    rechargeForm.remark = ''
    walletAlert.value = {
      level: 'success',
      message: t('personalCenter.wallet.createPaymentSuccess'),
    }
    if (isRechargePending.value) {
      startPolling()
    } else {
      stopPolling()
      await Promise.all([
        loadWallet(),
        loadTransactions(1),
      ])
    }
    if (payLink.value && String(currentRechargePayment.value?.interaction_mode || '').toLowerCase() === 'redirect') {
      window.open(payLink.value, '_blank', 'noopener')
    }
  } catch (err: any) {
    walletAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.wallet.errors.rechargeFailed'),
    }
  } finally {
    recharging.value = false
  }
}

const refreshRechargeStatus = async (silent = false) => {
  const rechargeNo = String(currentRecharge.value?.recharge_no || '').trim()
  if (!rechargeNo) return
  try {
    const response = await walletAPI.rechargeDetail(rechargeNo)
    const payload = response.data.data || {}
    syncRechargePayload(payload)

    const status = String(currentRecharge.value?.status || '').toLowerCase()
    if (status === 'success') {
      stopPolling()
      await Promise.all([
        loadWallet(),
        loadTransactions(1),
      ])
      walletAlert.value = {
        level: 'success',
        message: t('personalCenter.wallet.rechargeSuccess'),
      }
      return
    }
    if (status === 'failed' || status === 'expired') {
      stopPolling()
      walletAlert.value = {
        level: 'warning',
        message: t('personalCenter.wallet.errors.rechargeStateFailed'),
      }
      return
    }
    startPolling()
  } catch (err: any) {
    if (!silent) {
      walletAlert.value = {
        level: 'error',
        message: err?.message || t('personalCenter.wallet.errors.rechargeQueryFailed'),
      }
    }
  }
}

const checkRechargePayment = async () => {
  const paymentID = Number(currentRechargePayment.value?.id || 0)
  if (!Number.isFinite(paymentID) || paymentID <= 0) return
  checkingPayment.value = true
  try {
    const response = await walletAPI.captureRechargePayment(paymentID)
    const payload = response.data.data || {}
    syncRechargePayload(payload)
    await refreshRechargeStatus(true)
  } catch (err: any) {
    walletAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.wallet.errors.checkPayStatusFailed'),
    }
  } finally {
    checkingPayment.value = false
  }
}

const changePage = (page: number) => {
  if (page < 1 || page > pagination.value.total_page) return
  loadTransactions(page)
}

const refreshCurrentPage = async () => {
  await Promise.all([
    loadWallet(),
    loadTransactions(pagination.value.page),
  ])
}

watch(
  channels,
  (list) => {
    if (list.length === 0) {
      rechargeForm.channelId = 0
      return
    }
    if (!list.some((item: any) => item.id === rechargeForm.channelId)) {
      const first = list[0]
      if (!first) {
        rechargeForm.channelId = 0
        return
      }
      rechargeForm.channelId = first.id
    }
  },
  { immediate: true }
)

watch(
  () => qrDisplayContent.value,
  () => {
    void renderQRCodeImage()
  },
  { immediate: true }
)

onMounted(() => {
  initialize()
})

onUnmounted(() => {
  stopPolling()
})
</script>
