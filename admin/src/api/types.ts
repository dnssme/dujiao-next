export interface ApiResponse<T = unknown> {
  status_code: number
  msg: string
  data: T
  pagination?: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}
