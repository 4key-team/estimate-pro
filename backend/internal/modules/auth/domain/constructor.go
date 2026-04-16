// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxNameLen = 255

// NewUser constructs a User enforcing domain invariants: non-empty trimmed
// email and name (1..255 chars). PasswordHash may be empty for OAuth users.
// AvatarURL is optional. ID is auto-generated; PreferredLocale defaults to "ru".
func NewUser(email, passwordHash, name, avatarURL string) (*User, error) {
	trimmedEmail := strings.TrimSpace(email)
	if trimmedEmail == "" {
		return nil, ErrInvalidEmail
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > maxNameLen {
		return nil, ErrInvalidName
	}
	now := time.Now()
	return &User{
		ID:              uuid.New().String(),
		Email:           trimmedEmail,
		PasswordHash:    passwordHash,
		Name:            trimmedName,
		AvatarURL:       avatarURL,
		PreferredLocale: "ru",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// UpdateProfile applies partial updates. Empty strings mean "keep current".
// A non-empty but invalid name aborts without mutation. UpdatedAt advances
// only when the update succeeds.
func (u *User) UpdateProfile(name, avatarURL, telegramChatID, notificationEmail string) error {
	var newName string
	if name != "" {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" || len(trimmed) > maxNameLen {
			return ErrInvalidName
		}
		newName = trimmed
	}
	if newName != "" {
		u.Name = newName
	}
	if avatarURL != "" {
		u.AvatarURL = avatarURL
	}
	if telegramChatID != "" {
		u.TelegramChatID = telegramChatID
	}
	if notificationEmail != "" {
		u.NotificationEmail = notificationEmail
	}
	u.UpdatedAt = time.Now()
	return nil
}
