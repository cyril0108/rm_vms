package models

import "time"

type PartialUpdateInterfaces map[string]interface{}

func (pui PartialUpdateInterfaces) ApplyUpdatedAt() {
	pui["updated_at"] = time.Now().Unix()
}