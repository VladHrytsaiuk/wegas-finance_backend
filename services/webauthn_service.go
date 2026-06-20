package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type WebAuthnService interface {
	BeginRegistration(user *models.User) (*protocol.CredentialCreation, string, error)
	FinishRegistration(user *models.User, sessionID string, responseData json.RawMessage, origin string) error
	BeginLogin(email string) (*protocol.CredentialAssertion, string, error)
	FinishLogin(sessionID string, responseData json.RawMessage, origin string) (*models.User, error)
}

type webAuthnService struct {
	web      *webauthn.WebAuthn
	repo     repositories.WebAuthnRepository
	userRepo repositories.UserRepository
	
	// In-memory session store for challenges
	sessions   map[string]webauthn.SessionData
	sessionUsers map[string]string // sessionID -> userID
	sessionMu sync.RWMutex
}

func NewWebAuthnService(rpID, rpOrigin string, repo repositories.WebAuthnRepository, userRepo repositories.UserRepository) (WebAuthnService, error) {
	w, err := webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: "WeGaS Finance",
		RPOrigins:     []string{rpOrigin},
	})
	if err != nil {
		return nil, err
	}

	return &webAuthnService{
		web:          w,
		repo:         repo,
		userRepo:     userRepo,
		sessions:     make(map[string]webauthn.SessionData),
		sessionUsers: make(map[string]string),
	}, nil
}

func (s *webAuthnService) getWebAuthnUser(user *models.User) (*models.WebAuthnUserWrapper, error) {
	creds, err := s.repo.GetCredentialsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	webAuthnCreds := make([]webauthn.Credential, len(creds))
	for i, c := range creds {
		webAuthnCreds[i] = c.ToWebAuthnCredential()
	}

	return &models.WebAuthnUserWrapper{
		User:        user,
		Credentials: webAuthnCreds,
	}, nil
}

func (s *webAuthnService) BeginRegistration(user *models.User) (*protocol.CredentialCreation, string, error) {
	waUser, err := s.getWebAuthnUser(user)
	if err != nil {
		return nil, "", err
	}

	options, sessionData, err := s.web.BeginRegistration(waUser)
	if err != nil {
		return nil, "", err
	}

	sessionID := uuid.NewString()
	s.sessionMu.Lock()
	s.sessions[sessionID] = *sessionData
	s.sessionUsers[sessionID] = user.ID
	s.sessionMu.Unlock()

	// Auto-cleanup session after 5 minutes
	go func() {
		time.Sleep(5 * time.Minute)
		s.sessionMu.Lock()
		delete(s.sessions, sessionID)
		delete(s.sessionUsers, sessionID)
		s.sessionMu.Unlock()
	}()

	return options, sessionID, nil
}

func (s *webAuthnService) FinishRegistration(user *models.User, sessionID string, responseData json.RawMessage, origin string) error {
	s.sessionMu.RLock()
	sessionData, ok := s.sessions[sessionID]
	s.sessionMu.RUnlock()

	if !ok {
		return errors.New("session expired or not found")
	}

	waUser, err := s.getWebAuthnUser(user)
	if err != nil {
		return err
	}

	// Створюємо запит з правильним Origin для перевірки бібліотекою
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(responseData))
	req.Header.Set("Content-Type", "application/json")
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	// Host має бути саме доменним ім'ям (RPID), а не повним URL
	req.Host = s.web.Config.RPID

	credential, err := s.web.FinishRegistration(waUser, sessionData, req)
	if err != nil {
		return err
	}

	// Save credential to DB
	newCred := &models.WebAuthnCredential{
		Base:            models.Base{ID: uuid.NewString()},
		UserID:          user.ID,
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		SignCount:       credential.Authenticator.SignCount,
		CloneWarning:    credential.Authenticator.CloneWarning,
	}
	
	return s.repo.CreateCredential(newCred)
}

func (s *webAuthnService) BeginLogin(email string) (*protocol.CredentialAssertion, string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", errors.New("user not found")
	}

	waUser, err := s.getWebAuthnUser(user)
	if err != nil {
		return nil, "", err
	}

	options, sessionData, err := s.web.BeginLogin(waUser)
	if err != nil {
		return nil, "", err
	}

	sessionID := uuid.NewString()
	s.sessionMu.Lock()
	s.sessions[sessionID] = *sessionData
	s.sessionUsers[sessionID] = user.ID
	s.sessionMu.Unlock()

	return options, sessionID, nil
}

func (s *webAuthnService) FinishLogin(sessionID string, responseData json.RawMessage, origin string) (*models.User, error) {
	s.sessionMu.RLock()
	sessionData, ok := s.sessions[sessionID]
	userID, userOk := s.sessionUsers[sessionID]
	s.sessionMu.RUnlock()

	if !ok || !userOk {
		return nil, errors.New("session expired or not found")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	waUser, err := s.getWebAuthnUser(user)
	if err != nil {
		return nil, err
	}

	// Створюємо запит з правильним Origin
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(responseData))
	req.Header.Set("Content-Type", "application/json")
	if origin != "" {
		req.Header.Set("Origin", origin)
		req.Host = origin
	}

	credential, err := s.web.FinishLogin(waUser, sessionData, req)
	if err != nil {
		return nil, err
	}

	// Update sign count
	dbCred, err := s.repo.GetCredentialByID(credential.ID)
	if err == nil {
		dbCred.SignCount = credential.Authenticator.SignCount
		s.repo.UpdateCredential(dbCred)
	}

	return user, nil
}


