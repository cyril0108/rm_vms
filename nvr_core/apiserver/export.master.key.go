package apiserver

import (
	"encoding/hex"
	"net/http"
)

// =======================================
// The exported string can be converted back to raw bytes and save to the protected file path
// # Convert Hex back to raw bytes and save to the protected file path
// echo -n "INSERT_BACKUP_HEX_STRING_HERE" | xxd -r -p > /etc/nvr/master.key

// # Lock down the permissions
// chmod 0400 /etc/nvr/master.key
// =======================================

// HandleExportMasterKey serves the encryption key for backup purposes.
func (s *APIServer) HandleExportMasterKey(w http.ResponseWriter, r *http.Request) {
	// SECURITY CRITICAL: 
	// This endpoint MUST be placed behind your strictest Admin Authentication middleware.
	// If an unauthorized user hits this, your entire encryption scheme is compromised.

	// Assuming GlobalMasterKey is accessible in this package
	keyHex := hex.EncodeToString(s.CFG.Server.MasterKey())

	response := map[string]string{
		"master_key_hex": keyHex,
		"format":         "hexadecimal",
		"instructions":   "Store this key in a secure password manager. If the server hardware fails, you will need this exact string to decrypt camera passwords.",
	}

	if err := RespondJSON(w, response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}