export const toRFC3339 = (raw?: string) => {
  if (!raw) return undefined
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return undefined
  return date.toISOString()
}

export const formatDate = (raw?: string) => {
  if (!raw) return ''
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

export const formatMoney = (amount?: string | number, currency?: string) => {
  if (amount === null || amount === undefined || amount === '') return '-'
  if (!currency) return String(amount)
  return `${amount} ${currency}`
}

export const hasPositiveAmount = (amount?: string | number) => {
  if (amount === null || amount === undefined || amount === '') return false
  const value = Number(amount)
  return !Number.isNaN(value) && value > 0
}

export const getLocalizedText = (jsonData: any) => {
  if (!jsonData) return ''
  if (typeof jsonData === 'string') return jsonData
  return jsonData['zh-CN'] || jsonData['zh-TW'] || jsonData['en-US'] || Object.values(jsonData)[0] || ''
}
