package apiserver

import (
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/hardware"
	"nvr_core/utils"
)

// HandleAdminInitConfigure
//
func (s *APIServer) HandleGetMachineInfo(w http.ResponseWriter, r *http.Request) {

	info := dto.SystemMachineInfo{
		MachineID: hardware.GetPersistentMachineID(),
	}


	utils.RespondJSON(w, info, "success")
	return
}
