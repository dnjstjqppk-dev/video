package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
)

// GetFinanceSummary aggregates the double-entry ledger and payment
// transactions into a finance review snapshot. Debits must equal credits;
// any imbalance blocks batch approval.
func (h *Handler) GetFinanceSummary(w http.ResponseWriter, r *http.Request) {
	var debitTotal, creditTotal int64
	debitByAccount := map[string]int64{}
	creditByAccount := map[string]int64{}
	for _, entry := range h.Store.Ledger {
		switch entry.Direction {
		case "DEBIT":
			debitTotal += entry.AmountCents
			debitByAccount[entry.AccountType] += entry.AmountCents
		case "CREDIT":
			creditTotal += entry.AmountCents
			creditByAccount[entry.AccountType] += entry.AmountCents
		}
	}
	txByStatus := map[string]int{}
	var succeededCents int64
	for _, tx := range h.Store.Transactions {
		txByStatus[tx.Status]++
		if tx.Status == "succeeded" {
			succeededCents += tx.AmountCents
		}
	}
	var creatorPayable int64
	for _, entry := range h.Store.Ledger {
		if entry.AccountType == "creator_payable" && entry.Direction == "CREDIT" {
			creatorPayable += entry.AmountCents
		}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"balanced":             debitTotal == creditTotal,
		"debitTotalCents":      debitTotal,
		"creditTotalCents":     creditTotal,
		"debitByAccount":       debitByAccount,
		"creditByAccount":      creditByAccount,
		"transactionsByStatus": txByStatus,
		"succeededCents":       succeededCents,
		"creatorPayableCents":  creatorPayable,
		"ledgerEntries":        len(h.Store.Ledger),
		"batchId":              "BATCH-" + time.Now().Format("2006-01"),
		"generatedAt":          time.Now().Format(time.RFC3339),
	})
}

// ApproveBatch records a step-up authenticated settlement batch approval.
// The caller must present a valid TOTP code enrolled on their own account,
// and the ledger must balance. The approval is appended to the audit log.
func (h *Handler) ApproveBatch(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	var payload struct {
		BatchID  string `json:"batchId"`
		TOTPCode string `json:"totpCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	if !user.MFAEnabled || user.TOTPSecret == "" {
		identity.WriteError(w, r, http.StatusPreconditionFailed, "MFA_REQUIRED", "Enroll TOTP via /v1/me/mfa/setup before approving batches")
		return
	}
	if !security.ValidateTOTP(user.TOTPSecret, payload.TOTPCode) {
		identity.WriteError(w, r, http.StatusUnauthorized, "INVALID_TOTP", "Step-up code is invalid or expired")
		return
	}
	var debitTotal, creditTotal int64
	for _, entry := range h.Store.Ledger {
		switch entry.Direction {
		case "DEBIT":
			debitTotal += entry.AmountCents
		case "CREDIT":
			creditTotal += entry.AmountCents
		}
	}
	if debitTotal != creditTotal {
		identity.WriteError(w, r, http.StatusConflict, "LEDGER_IMBALANCED", "Ledger debits do not equal credits; batch blocked")
		return
	}
	batchID := security.SanitizeUserText(payload.BatchID, 64)
	if batchID == "" {
		batchID = "BATCH-" + time.Now().Format("2006-01")
	}
	approvalID := "apr_" + time.Now().Format("20060102150405")
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        approvalID,
		"actorId":   user.ID,
		"action":    "APPROVE_SETTLEMENT_BATCH",
		"target":    batchID,
		"ip":        r.RemoteAddr,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"approvalId":  approvalID,
		"batchId":     batchID,
		"approvedBy":  user.ID,
		"debitTotal":  debitTotal,
		"creditTotal": creditTotal,
		"status":      "approved",
	})
}
