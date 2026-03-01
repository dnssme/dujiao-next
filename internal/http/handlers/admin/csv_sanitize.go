package admin

import "strings"

// csvFormulaInjectionPrefixes 可触发电子表格公式执行的前缀字符。
var csvFormulaInjectionPrefixes = []string{"=", "+", "-", "@", "\t", "\r", "\n"}

// sanitizeCSVField 净化将写入 CSV 的字段值，防止 CSV 注入（PCI-DSS 6.5.1）。
func sanitizeCSVField(value string) string {
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
