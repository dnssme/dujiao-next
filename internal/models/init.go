package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

// InitDefaultAdmin 初始化默认管理员账号
func InitDefaultAdmin(username, password string) error {
	var count int64
	DB.Model(&Admin{}).Count(&count)

	// 如果已有管理员，确保默认 admin 拥有超级管理员权限
	if count > 0 {
		if err := DB.Model(&Admin{}).Where("username = ?", "admin").Update("is_super", true).Error; err != nil {
			logger.Warnw("ensure_default_admin_super_failed", "error", err)
		}
		return nil
	}

	// 创建默认管理员
	if username == "" {
		username = "admin"
	}
	generated := false
	if password == "" {
		var err error
		password, err = generateRandomPassword(24)
		if err != nil {
			return fmt.Errorf("generate random password failed: %w", err)
		}
		generated = true
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := Admin{
		Username:     username,
		PasswordHash: string(hash),
		IsSuper:      strings.EqualFold(strings.TrimSpace(username), "admin"),
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	if generated {
		logger.Warnw("default_admin_created_with_random_password", "username", username, "password", password)
		logger.Warnw("default_admin_password_change_required", "username", username)
	} else {
		logger.Warnw("default_admin_created", "username", username, "password_hidden", true)
	}

	return nil
}

func generateRandomPassword(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf)[:length], nil
}
