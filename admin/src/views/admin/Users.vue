<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import IdCell from '@/components/IdCell.vue'
import { userStatusClass, userStatusLabel } from '@/utils/status'
import { formatDate, formatMoney } from '@/utils/format'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogHeader, DialogScrollContent, DialogTitle } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { confirmAction } from '@/utils/confirm'

const { t } = useI18n()
const loading = ref(true)
const users = ref<any[]>([])
const selectedIds = ref<number[]>([])
const adminPath = import.meta.env.VITE_ADMIN_PATH || ''
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})
const jumpPage = ref('')
const filters = reactive({
  keyword: '',
  status: '__all__',
  createdFrom: '',
  createdTo: '',
  lastLoginFrom: '',
  lastLoginTo: '',
})
const normalizeFilterValue = (value: string) => (value === '__all__' ? '' : value)

const showModal = ref(false)
const error = ref('')
const editingId = ref<number | null>(null)
const siteCurrency = ref('CNY')
const form = reactive({
  email: '',
  nickname: '',
  password: '',
  locale: 'zh-CN',
  status: 'active',
})

const fetchUsers = async (page = 1) => {
  loading.value = true
  try {
    const response = await adminAPI.getUsers({
      page,
      page_size: pagination.value.page_size,
      keyword: filters.keyword || undefined,
      status: normalizeFilterValue(filters.status) || undefined,
      created_from: filters.createdFrom || undefined,
      created_to: filters.createdTo || undefined,
      last_login_from: filters.lastLoginFrom || undefined,
      last_login_to: filters.lastLoginTo || undefined,
    })
    users.value = (response.data.data as any[]) || []
    pagination.value = response.data.pagination || pagination.value
    selectedIds.value = []
  } catch {
    users.value = []
  } finally {
    loading.value = false
  }
}

const fetchSiteCurrency = async () => {
  try {
    const response = await adminAPI.getSettings({ key: 'site_config' })
    const data = response.data?.data as any
    const raw = String(data?.currency || 'CNY').trim().toUpperCase()
    siteCurrency.value = /^[A-Z]{3}$/.test(raw) ? raw : 'CNY'
  } catch {
    siteCurrency.value = 'CNY'
  }
}

const handleSearch = () => {
  fetchUsers(1)
}

const refresh = () => {
  fetchUsers(pagination.value.page)
}

const resetFilters = () => {
  filters.keyword = ''
  filters.status = '__all__'
  filters.createdFrom = ''
  filters.createdTo = ''
  filters.lastLoginFrom = ''
  filters.lastLoginTo = ''
  fetchUsers(1)
}

const userDetailLink = (userId: number) => `${adminPath}/users/${userId}`

const allSelected = computed(() => {
  if (users.value.length === 0) return false
  return users.value.every((item) => selectedIds.value.includes(item.id))
})

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedIds.value = []
    return
  }
  selectedIds.value = users.value.map((item) => item.id)
}

const batchUpdateStatus = async (status: string) => {
  if (selectedIds.value.length === 0) return
  const confirmed = await confirmAction(t('admin.users.batch.confirm', { count: selectedIds.value.length }))
  if (!confirmed) return
  try {
    await adminAPI.batchUpdateUserStatus({
      user_ids: selectedIds.value,
      status,
    })
    fetchUsers(pagination.value.page)
  } catch (err: any) {
    error.value = err?.message || t('admin.users.errors.updateFailed')
  }
}

const changePage = (page: number) => {
  if (page < 1 || page > pagination.value.total_page) return
  fetchUsers(page)
}

const jumpToPage = () => {
  if (!jumpPage.value) return
  const raw = Number(jumpPage.value)
  if (Number.isNaN(raw)) return
  const target = Math.min(Math.max(Math.floor(raw), 1), pagination.value.total_page)
  if (target === pagination.value.page) return
  changePage(target)
}

const openEditModal = (user: any) => {
  editingId.value = user.id
  form.email = user.email || ''
  form.nickname = user.display_name || ''
  form.password = ''
  form.locale = user.locale || 'zh-CN'
  form.status = user.status || 'active'
  error.value = ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingId.value = null
  error.value = ''
}

const handleSubmit = async () => {
  if (!editingId.value) return
  try {
    await adminAPI.updateUser(editingId.value, {
      email: form.email,
      nickname: form.nickname,
      password: form.password || undefined,
      locale: form.locale,
      status: form.status,
    })
    closeModal()
    fetchUsers(pagination.value.page)
  } catch (err: any) {
    error.value = err?.message || t('admin.users.errors.updateFailed')
  }
}

const statusClass = (status: string) => userStatusClass(status)
const statusLabel = (status: string) => userStatusLabel(t, status)

const formatLocale = (raw?: string) => {
  if (!raw) return '-'
  const map: Record<string, string> = {
    'zh-CN': t('admin.common.lang.zhCN'),
    'zh-TW': t('admin.common.lang.zhTW'),
    'en-US': t('admin.common.lang.enUS'),
  }
  return map[raw] || raw
}

onMounted(() => {
  fetchSiteCurrency()
  fetchUsers()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-semibold">{{ t('admin.users.title') }}</h1>
    </div>

    <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
      <div class="flex flex-wrap items-center gap-3">
        <div class="w-full md:w-64">
          <Input v-model="filters.keyword" :placeholder="t('admin.users.filterKeyword')" @update:modelValue="handleSearch" />
        </div>
        <div class="w-full md:w-40">
          <Select v-model="filters.status" @update:modelValue="handleSearch">
            <SelectTrigger class="h-9 w-full">
            <SelectValue :placeholder="t('admin.users.filterStatusAll')" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__all__">{{ t('admin.users.filterStatusAll') }}</SelectItem>
            <SelectItem value="active">{{ t('admin.users.status.active') }}</SelectItem>
            <SelectItem value="disabled">{{ t('admin.users.status.disabled') }}</SelectItem>
          </SelectContent>
        </Select>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-xs text-muted-foreground whitespace-nowrap">{{ t('admin.users.filterCreatedRange') }}</span>
          <Input
            v-model="filters.createdFrom"
            type="date"
            class="h-9 w-full md:w-auto"
            :placeholder="t('admin.users.filterCreatedFrom')"
            @update:modelValue="handleSearch"
          />
          <span class="text-muted-foreground">-</span>
          <Input
            v-model="filters.createdTo"
            type="date"
            class="h-9 w-full md:w-auto"
            :placeholder="t('admin.users.filterCreatedTo')"
            @update:modelValue="handleSearch"
          />
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-xs text-muted-foreground whitespace-nowrap">{{ t('admin.users.filterLastLoginRange') }}</span>
          <Input
            v-model="filters.lastLoginFrom"
            type="date"
            class="h-9 w-full md:w-auto"
            :placeholder="t('admin.users.filterLastLoginFrom')"
            @update:modelValue="handleSearch"
          />
          <span class="text-muted-foreground">-</span>
          <Input
            v-model="filters.lastLoginTo"
            type="date"
            class="h-9 w-full md:w-auto"
            :placeholder="t('admin.users.filterLastLoginTo')"
            @update:modelValue="handleSearch"
          />
        </div>
        <div class="flex-1"></div>
        <Button size="sm" variant="outline" @click="resetFilters">{{ t('admin.common.reset') }}</Button>
        <Button size="sm" @click="refresh">{{ t('admin.common.refresh') }}</Button>
      </div>
    </div>

    <div class="rounded-xl border border-border bg-card">
      <Table>
        <TableHeader class="border-b border-border bg-muted/40 text-xs uppercase text-muted-foreground">
          <TableRow>
            <TableHead class="px-6 py-3">
              <input type="checkbox" :checked="allSelected" class="h-4 w-4 accent-primary" @change="toggleSelectAll" />
            </TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.id') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.email') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.nickname') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.status') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.locale') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.walletBalance') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.createdAt') }}</TableHead>
            <TableHead class="px-6 py-3">{{ t('admin.users.table.lastLoginAt') }}</TableHead>
            <TableHead class="px-6 py-3 text-right">{{ t('admin.users.table.action') }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody class="divide-y divide-border">
          <TableRow v-if="loading">
            <TableCell colspan="10" class="px-6 py-8 text-center text-muted-foreground">{{ t('admin.common.loading') }}</TableCell>
          </TableRow>
          <TableRow v-else-if="users.length === 0">
            <TableCell colspan="10" class="px-6 py-8 text-center text-muted-foreground">{{ t('admin.users.empty') }}</TableCell>
          </TableRow>
          <TableRow v-for="user in users" :key="user.id" class="hover:bg-muted/30">
            <TableCell class="px-6 py-4">
              <input type="checkbox" :value="user.id" v-model="selectedIds" class="h-4 w-4 accent-primary" />
            </TableCell>
            <TableCell class="px-6 py-4">
              <IdCell :value="user.id" />
            </TableCell>
            <TableCell class="px-6 py-4 text-foreground">{{ user.email }}</TableCell>
            <TableCell class="px-6 py-4 text-muted-foreground">{{ user.display_name || '-' }}</TableCell>
            <TableCell class="px-6 py-4 text-xs">
              <span class="inline-flex rounded-full border px-2.5 py-1 text-xs" :class="statusClass(user.status)">
                {{ statusLabel(user.status) }}
              </span>
            </TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ formatLocale(user.locale) }}</TableCell>
            <TableCell class="px-6 py-4 text-xs font-mono text-foreground">{{ formatMoney(user.wallet_balance, siteCurrency) }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ formatDate(user.created_at) }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ formatDate(user.last_login_at) }}</TableCell>
            <TableCell class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2">
                <Button as-child size="sm" variant="outline">
                  <router-link :to="userDetailLink(user.id)">{{ t('admin.users.actions.detail') }}</router-link>
                </Button>
                <Button size="sm" variant="outline" @click="openEditModal(user)">
                  {{ t('admin.users.actions.edit') }}
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <div
        v-if="pagination.total_page > 1"
        class="flex flex-wrap items-center justify-between gap-3 border-t border-border px-6 py-4"
      >
        <div class="flex items-center gap-3">
          <span class="text-xs text-muted-foreground">
            {{ t('admin.common.pageInfo', { total: pagination.total, page: pagination.page, totalPage: pagination.total_page }) }}
          </span>
          <div v-if="selectedIds.length > 0" class="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              class="border-emerald-200 text-emerald-700 hover:bg-emerald-50"
              @click="batchUpdateStatus('active')"
            >
              {{ t('admin.users.batch.enable') }}
            </Button>
            <Button
              size="sm"
              variant="outline"
              class="border-destructive/40 text-destructive hover:bg-destructive/10"
              @click="batchUpdateStatus('disabled')"
            >
              {{ t('admin.users.batch.disable') }}
            </Button>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex items-center gap-2">
            <Input
              v-model="jumpPage"
              type="number"
              min="1"
              :max="pagination.total_page"
              class="h-8 w-20"
              :placeholder="t('admin.common.jumpPlaceholder')"
            />
            <Button variant="outline" size="sm" class="h-8" @click="jumpToPage">
              {{ t('admin.common.jumpTo') }}
            </Button>
          </div>
          <div class="flex items-center gap-2">
            <Button variant="outline" size="sm" class="h-8" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
              {{ t('admin.common.prevPage') }}
            </Button>
            <Button
              variant="outline"
              size="sm"
              class="h-8"
              :disabled="pagination.page >= pagination.total_page"
              @click="changePage(pagination.page + 1)"
            >
              {{ t('admin.common.nextPage') }}
            </Button>
          </div>
        </div>
      </div>
    </div>

    <Dialog v-model:open="showModal" @update:open="(value) => { if (!value) closeModal() }">
      <DialogScrollContent class="w-full max-w-xl">
        <DialogHeader>
          <DialogTitle>{{ t('admin.users.modal.editTitle') }}</DialogTitle>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="handleSubmit">
          <div class="grid grid-cols-1 gap-4">
            <div>
              <label class="block text-xs font-medium text-muted-foreground mb-1.5">{{ t('admin.users.form.email') }}</label>
              <Input v-model="form.email" :placeholder="t('admin.users.form.emailPlaceholder')" />
            </div>
            <div>
              <label class="block text-xs font-medium text-muted-foreground mb-1.5">{{ t('admin.users.form.nickname') }}</label>
              <Input v-model="form.nickname" :placeholder="t('admin.users.form.nicknamePlaceholder')" />
            </div>
            <div>
              <label class="block text-xs font-medium text-muted-foreground mb-1.5">{{ t('admin.users.form.password') }}</label>
              <Input v-model="form.password" type="password" :placeholder="t('admin.users.form.passwordPlaceholder')" />
              <p class="mt-1 text-xs text-muted-foreground">{{ t('admin.users.form.passwordTip') }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-muted-foreground mb-1.5">{{ t('admin.users.form.locale') }}</label>
              <Select v-model="form.locale">
                <SelectTrigger class="h-9 w-full">
                  <SelectValue :placeholder="t('admin.common.lang.zhCN')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="zh-CN">{{ t('admin.common.lang.zhCN') }}</SelectItem>
                  <SelectItem value="zh-TW">{{ t('admin.common.lang.zhTW') }}</SelectItem>
                  <SelectItem value="en-US">{{ t('admin.common.lang.enUS') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <label class="block text-xs font-medium text-muted-foreground mb-1.5">{{ t('admin.users.form.status') }}</label>
              <Select v-model="form.status">
                <SelectTrigger class="h-9 w-full">
                  <SelectValue :placeholder="t('admin.users.status.active')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">{{ t('admin.users.status.active') }}</SelectItem>
                  <SelectItem value="disabled">{{ t('admin.users.status.disabled') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div v-if="error" class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
            {{ error }}
          </div>

          <div class="flex justify-end gap-3">
            <Button type="button" variant="outline" @click="closeModal">{{ t('admin.common.cancel') }}</Button>
            <Button type="submit">{{ t('admin.common.save') }}</Button>
          </div>
        </form>
      </DialogScrollContent>
    </Dialog>
  </div>
</template>
