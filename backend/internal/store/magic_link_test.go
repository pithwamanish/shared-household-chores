package store

import (
	"testing"
	"time"
)

func TestMemoryStoreMagicLinkSingleUseAndExpiration(t *testing.T) {
	s := NewMemoryStore()

	// 1. Create a magic link
	token, mem, hh, err := s.CreateMagicLink("sarah@example.com")
	if err != nil {
		t.Fatalf("failed to create magic link: %v", err)
	}
	if token == "" || mem == nil || hh == nil {
		t.Fatalf("expected valid token, member, and household")
	}

	// Verify expiration is set to ~15 minutes from now
	s.mu.RLock()
	entry, ok := s.magicLinks[token]
	s.mu.RUnlock()
	if !ok {
		t.Fatalf("expected token %s to exist in store", token)
	}
	if entry.expiresAt.Sub(entry.createdAt) != 15*time.Minute {
		t.Fatalf("expected expiresAt to be 15 minutes after createdAt, got %v", entry.expiresAt.Sub(entry.createdAt))
	}

	// 2. First verification must succeed
	verifiedMem, verifiedHh, err := s.VerifyMagicLink(token)
	if err != nil {
		t.Fatalf("first verification failed: %v", err)
	}
	if verifiedMem.ID != "m-sarah" || verifiedHh.ID != "h-roommates" {
		t.Fatalf("unexpected member or household on verify: %s, %s", verifiedMem.ID, verifiedHh.ID)
	}

	// 3. Second verification of the exact same token MUST fail (single-use)
	_, _, err = s.VerifyMagicLink(token)
	if err == nil {
		t.Fatal("expected error on reusing single-use magic link token, got nil")
	}
	if err.Error() != "INVALID_OR_EXPIRED_TOKEN" {
		t.Fatalf("expected INVALID_OR_EXPIRED_TOKEN, got %v", err)
	}

	// 4. Test expired token
	now := time.Now().UTC()
	expiredToken := "MAGIC-EXPIRED-TEST"
	s.mu.Lock()
	s.magicLinks[expiredToken] = magicLinkEntry{
		memberID:  "m-sarah",
		createdAt: now.Add(-20 * time.Minute),
		expiresAt: now.Add(-5 * time.Minute),
	}
	s.mu.Unlock()

	_, _, err = s.VerifyMagicLink(expiredToken)
	if err == nil {
		t.Fatal("expected error when verifying expired magic link, got nil")
	}
	if err.Error() != "INVALID_OR_EXPIRED_TOKEN" {
		t.Fatalf("expected INVALID_OR_EXPIRED_TOKEN for expired link, got %v", err)
	}

	// Verify expired token was purged from store
	s.mu.RLock()
	_, exists := s.magicLinks[expiredToken]
	s.mu.RUnlock()
	if exists {
		t.Fatal("expected expired token to be purged from magicLinks map")
	}
}
