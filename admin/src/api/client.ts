import axios from 'axios'
import i18n from '@/i18n'
import type { ApiResponse } from './types'
import { notifyError } from '@/utils/notify'

export type { ApiResponse }

const t = (key: string, params?: Record<string, unknown>) =>
  (params ? i18n.global.t(key, params) : i18n.global.t(key)) as string

interface NotifiedError extends Error {
  __notified?: boolean
}

const createNotifiedError = (message: string): NotifiedError => {
  const error = new Error(message) as NotifiedError
  error.__notified = true
  return error
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''
const API_PREFIX = '/api/v1'
const ADMIN_PATH = import.meta.env.VITE_ADMIN_PATH || ''

const redirectToLogin = () => {
  localStorage.removeItem('admin_token')
  window.location.href = `${ADMIN_PATH}/login`
}

export const api = axios.create({
  baseURL: `${API_BASE_URL}${API_PREFIX}`,
  timeout: 10000,
})

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (response) => {
    const data: ApiResponse = response.data
    if (!data) {
      const message = t('common.api.responseMissing')
      notifyError(message)
      return Promise.reject(createNotifiedError(message))
    }
    if (typeof data.status_code !== 'undefined' && data.status_code !== 0) {
      const fallbackMessage = t('common.api.requestFailed')
      const message = data.msg || fallbackMessage
      if (data.status_code === 401) {
        notifyError(message)
        redirectToLogin()
        return Promise.reject(createNotifiedError(message))
      }
      notifyError(message)
      return Promise.reject(createNotifiedError(message))
    }
    return response
  },
  (error) => {
    if (error.response) {
      const status = error.response.status
      let message = t('common.api.requestFailed')
      switch (status) {
        case 401:
          message = t('common.api.unauthorized')
          redirectToLogin()
          break
        case 403:
          message = t('common.api.forbidden')
          break
        case 404:
          message = t('common.api.notFound')
          break
        case 500:
          message = t('common.api.serverError')
          break
        case 502:
          message = t('common.api.badGateway')
          break
        case 503:
          message = t('common.api.serviceUnavailable')
          break
        default:
          message = t('common.api.requestFailedStatus', { status })
      }
      notifyError(message)
      return Promise.reject(createNotifiedError(message))
    }
    if (error.request) {
      const message = t('common.api.networkError')
      notifyError(message)
      return Promise.reject(createNotifiedError(message))
    }
    if ((error as NotifiedError)?.__notified) {
      return Promise.reject(error)
    }
    if (error?.message) {
      notifyError(error.message)
      return Promise.reject(createNotifiedError(error.message))
    }
    return Promise.reject(error)
  }
)
