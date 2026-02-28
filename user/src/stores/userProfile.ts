import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { userOrderAPI, userProfileAPI } from '../api'
import type {
    ChangeEmailPayload,
    ChangeUserPasswordPayload,
    SendChangeEmailCodePayload,
    TelegramAuthPayload,
    TelegramBindingData,
    UpdateUserProfilePayload,
    UserLoginLogItem,
    UserProfileData,
} from '../api'
import { useUserAuthStore } from './userAuth'

interface PersonalOrderSummary {
    id: number
    order_no: string
    total_amount?: string
    currency?: string
    status?: string
    created_at?: string
}

const normalizeErrorMessage = (error: unknown, fallback: string) => {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }
    return fallback
}

export const useUserProfileStore = defineStore('user-profile', () => {
    const userAuthStore = useUserAuthStore()

    const profile = ref<UserProfileData | null>(null)
    const recentOrders = ref<PersonalOrderSummary[]>([])
    const recentLoginLogs = ref<UserLoginLogItem[]>([])
    const telegramBinding = ref<TelegramBindingData | null>(null)

    const loadingProfile = ref(false)
    const savingProfile = ref(false)
    const loadingOrders = ref(false)
    const loadingLoginLogs = ref(false)
    const loadingTelegramBinding = ref(false)
    const bindingTelegram = ref(false)
    const unbindingTelegram = ref(false)
    const sendingCode = ref(false)
    const changingEmail = ref(false)
    const changingPassword = ref(false)

    const profileError = ref('')
    const securityError = ref('')

    const displayName = computed(() => {
        if (profile.value?.nickname && profile.value.nickname.trim() !== '') {
            return profile.value.nickname
        }
        return profile.value?.email || '-'
    })

    const clearProfileError = () => {
        profileError.value = ''
    }

    const clearSecurityError = () => {
        securityError.value = ''
    }

    const loadProfile = async () => {
        loadingProfile.value = true
        clearProfileError()
        try {
            const response = await userProfileAPI.current()
            const data = response.data.data
            profile.value = data
            userAuthStore.syncUserProfile(data)
            return true
        } catch (error) {
            profile.value = null
            profileError.value = normalizeErrorMessage(error, '加载个人资料失败')
            return false
        } finally {
            loadingProfile.value = false
        }
    }

    const saveProfile = async (payload: UpdateUserProfilePayload) => {
        savingProfile.value = true
        clearProfileError()
        try {
            const response = await userProfileAPI.updateProfile(payload)
            const data = response.data.data
            profile.value = data
            userAuthStore.syncUserProfile(data)
            return true
        } catch (error) {
            profileError.value = normalizeErrorMessage(error, '保存个人资料失败')
            return false
        } finally {
            savingProfile.value = false
        }
    }

    const sendChangeEmailCode = async (payload: SendChangeEmailCodePayload) => {
        sendingCode.value = true
        clearSecurityError()
        try {
            await userProfileAPI.sendChangeEmailCode(payload)
            return true
        } catch (error) {
            securityError.value = normalizeErrorMessage(error, '发送验证码失败')
            return false
        } finally {
            sendingCode.value = false
        }
    }

    const changeEmail = async (payload: ChangeEmailPayload) => {
        changingEmail.value = true
        clearSecurityError()
        try {
            const response = await userProfileAPI.changeEmail(payload)
            const data = response.data.data
            profile.value = data
            userAuthStore.syncUserProfile(data)
            return true
        } catch (error) {
            securityError.value = normalizeErrorMessage(error, '更换邮箱失败')
            return false
        } finally {
            changingEmail.value = false
        }
    }

    const changePassword = async (payload: ChangeUserPasswordPayload) => {
        changingPassword.value = true
        clearSecurityError()
        try {
            await userProfileAPI.changePassword(payload)
            return true
        } catch (error) {
            securityError.value = normalizeErrorMessage(error, '修改密码失败')
            return false
        } finally {
            changingPassword.value = false
        }
    }

    const loadRecentOrders = async (limit = 5) => {
        loadingOrders.value = true
        try {
            const response = await userOrderAPI.list({ page: 1, page_size: limit })
            const data = response.data.data
            recentOrders.value = Array.isArray(data) ? (data as PersonalOrderSummary[]) : []
            return true
        } catch {
            recentOrders.value = []
            return false
        } finally {
            loadingOrders.value = false
        }
    }


    const loadRecentLoginLogs = async (limit = 5) => {
        loadingLoginLogs.value = true
        try {
            const response = await userProfileAPI.loginLogs({ page: 1, page_size: limit })
            const data = response.data.data
            recentLoginLogs.value = Array.isArray(data) ? (data as UserLoginLogItem[]) : []
            return true
        } catch {
            recentLoginLogs.value = []
            return false
        } finally {
            loadingLoginLogs.value = false
        }
    }

    const loadTelegramBinding = async () => {
        loadingTelegramBinding.value = true
        clearSecurityError()
        try {
            const response = await userProfileAPI.getTelegramBinding()
            telegramBinding.value = response.data.data || { bound: false }
            return true
        } catch (error) {
            telegramBinding.value = null
            securityError.value = normalizeErrorMessage(error, '加载 Telegram 绑定信息失败')
            return false
        } finally {
            loadingTelegramBinding.value = false
        }
    }

    const bindTelegram = async (payload: TelegramAuthPayload) => {
        bindingTelegram.value = true
        clearSecurityError()
        try {
            const response = await userProfileAPI.bindTelegram(payload)
            telegramBinding.value = response.data.data || { bound: true }
            return true
        } catch (error) {
            securityError.value = normalizeErrorMessage(error, '绑定 Telegram 失败')
            return false
        } finally {
            bindingTelegram.value = false
        }
    }

    const unbindTelegram = async () => {
        unbindingTelegram.value = true
        clearSecurityError()
        try {
            await userProfileAPI.unbindTelegram()
            telegramBinding.value = { bound: false }
            return true
        } catch (error) {
            securityError.value = normalizeErrorMessage(error, '解绑 Telegram 失败')
            return false
        } finally {
            unbindingTelegram.value = false
        }
    }

    return {
        profile,
        recentOrders,
        recentLoginLogs,
        telegramBinding,
        loadingProfile,
        savingProfile,
        loadingOrders,
        loadingLoginLogs,
        loadingTelegramBinding,
        bindingTelegram,
        unbindingTelegram,
        sendingCode,
        changingEmail,
        changingPassword,
        profileError,
        securityError,
        displayName,
        clearProfileError,
        clearSecurityError,
        loadProfile,
        saveProfile,
        sendChangeEmailCode,
        changeEmail,
        changePassword,
        loadRecentOrders,
        loadRecentLoginLogs,
        loadTelegramBinding,
        bindTelegram,
        unbindTelegram,
    }
})
