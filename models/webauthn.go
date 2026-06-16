package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnCredential stores a user's passkey credentials
type WebAuthnCredential struct {
	Base
	UserID string `json:"user_id" gorm:"index"`
	
	// WebAuthn specific fields
	ID              []byte `json:"credential_id" gorm:"uniqueIndex"`
	PublicKey       []byte `json:"public_key"`
	AttestationType string `json:"attestation_type"`
	Transport       string `json:"transport"` // JSON encoded array of strings
	SignCount       uint32 `json:"sign_count"`
	CloneWarning    bool   `json:"clone_warning"`
}

// ToWebAuthnCredential converts our DB model to the library's struct
func (c *WebAuthnCredential) ToWebAuthnCredential() webauthn.Credential {
	return webauthn.Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Authenticator: webauthn.Authenticator{
			SignCount: c.SignCount,
			CloneWarning: c.CloneWarning,
		},
	}
}

// UpdateUser - helper to satisfy webauthn.User interface on our User model
// We can't easily add methods to the User model in identity.go if we want to keep it clean,
// but we can use a wrapper or just extend the User model.
// I'll extend the User model in identity.go or use a wrapper here.

type WebAuthnUserWrapper struct {
	User        *User
	Credentials []webauthn.Credential
}

func (u *WebAuthnUserWrapper) WebAuthnID() []byte {
	return []byte(u.User.ID)
}

func (u *WebAuthnUserWrapper) WebAuthnName() string {
	return u.User.Email
}

func (u *WebAuthnUserWrapper) WebAuthnDisplayName() string {
	return u.User.Name
}

func (u *WebAuthnUserWrapper) WebAuthnIcon() string {
	return u.User.AvatarURL
}

func (u *WebAuthnUserWrapper) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
