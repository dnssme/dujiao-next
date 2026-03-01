package service

import "strings"

// csvFormulaInjectionPrefixes 可触发电子表格公式执行的前缀字符。
// PCI-DSS 6.5.1 — 防止 CSV 注入（Formula Injection / DDE）。
var csvFormulaInjectionPrefixes = []string{"=", "+", "-", "@", "\t", "\r", "\n"}

// SanitizeCSVField 净化将写入 CSV 的字段值。
// 若值以可触发公式执行的字符开头，前置单引号使电子表格将其作为纯文本处理。
func SanitizeCSVField(value string) string {
	if value == "" {
		return value
	}
	for _, prefix := range csvFormulaInjectionPrefixes {
		if strings.HasPrefix(value, prefix) {
			return "'" + value
		}
	}
	return value
}
