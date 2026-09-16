package handlers

import (
	"github.com/choresync/backend/internal/store"
)

// ApprovalHandler is an alias for CompletionHandler to support approval workflow semantics.
type ApprovalHandler = CompletionHandler

// NewApprovalHandler creates an ApprovalHandler backed by the store.
func NewApprovalHandler(s store.Store) *ApprovalHandler {
	return NewCompletionHandler(s)
}
