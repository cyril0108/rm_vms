package tw

import (
	"fmt"
	"nvr_core/db/models"
)

type I18NMessage string

type TranslateMap map[any]string

const (
	MSGCameraOffline I18NMessage = "MSGCameraOffline"
	MSGCameraBackOnline I18NMessage = "MSGCameraBackOnline"
	MSGDiskNearFullWarning I18NMessage = "MSGDiskNearFullWarning"
	MSGDiskFullWarning I18NMessage = "MSGDiskFullWarning"
)


var Events TranslateMap = TranslateMap {
	models.EventTypeCameraOffline: "設備斷線",
	models.EventTypeDiskWarning: "硬碟空間",
}

var EventsMessage TranslateMap = TranslateMap {
	MSGCameraOffline: "\"%v\"目前離線",
	MSGCameraBackOnline: "\"%v\"恢復連線",
	MSGDiskNearFullWarning: "硬碟空間即將達到上限(%v)，目前：%v",
	MSGDiskFullWarning: "硬碟空間達到回收條件，準備進行回收",
}

type Translator struct {}

func (t *Translator) EventsMessage() TranslateMap {
	return EventsMessage
}

func (t *Translator) Events() TranslateMap {
	return Events
}


func (t *Translator) Translate(k I18NMessage, vals...any) string {
	return fmt.Sprintf(EventsMessage[k], vals...)
}

func (t TranslateMap) Translate(k I18NMessage, vals...any) string {
	return fmt.Sprintf(t[k], vals...)
}