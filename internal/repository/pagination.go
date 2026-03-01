package repository

import (
	"strings"

	"gorm.io/gorm"
)

// maxPageSize 仓储层分页上限 — 作为防御性兜底，即使上层未做归一化也不会
// 产生超大 LIMIT（CIS Control 4.8 / PCI-DSS 6.5.6）。
const maxPageSize = 500

// applyPagination 应用分页参数，统一处理非法页码与偏移量。
func applyPagination(query *gorm.DB, page, pageSize int) *gorm.DB {
	if query == nil || pageSize <= 0 {
		return query
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return query.Limit(pageSize).Offset(offset)
}

// escapeLikePattern 转义 SQL LIKE 模式中的通配符，防止搜索绕过。
// PCI-DSS 6.5.1 — 去除 % 多字符通配符与反斜杠，阻止大面积模糊匹配。
// 注: 不去除 _ (单字符通配符) — 其攻击面极小（仅匹配一个字符），
// 且跨数据库(PostgreSQL/MySQL/SQLite)统一转义 _ 需要额外 ESCAPE 子句，
// 复杂度与收益不成比例。
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "%", "")
	return s
}
