package shared

// maxPageNumber 分页最大页码，防止超大偏移量导致数据库性能问题。
const maxPageNumber = 10000

// NormalizePagination 归一化分页参数。
func NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if page > maxPageNumber {
		page = maxPageNumber
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
