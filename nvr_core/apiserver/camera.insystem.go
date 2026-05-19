package apiserver

import (
	"context"
	"iter"
	"nvr_core/db/models"
	"nvr_core/onvif"
	"nvr_core/onvif/discovery"
)


// Source - https://stackoverflow.com/a/71624929
// Posted by jub0bs, modified by community. See post 'Timeline' for change history
// Retrieved 2026-05-12, License - CC BY-SA 4.0
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
    return func(yield func(U) bool) {
        for a := range seq {
            if !yield(f(a)) {
                return
            }
        }
    }
}


func (s *APIServer) ApplyCameraInSystemCheck(list *[]models.Camera, result *discovery.VerifyResult) *discovery.VerifyResult {
	if result == nil {
		return nil
	}

	for _, cam := range *list {
		if cam.IPAddress == result.IP {
			result.InSystem = true
			break
		}
	}

	return result
}

func (s *APIServer) ApplyCamerasInSystemCheck(ctx context.Context, results []discovery.VerifyResult) []discovery.VerifyResult {

	if(results == nil) {
		return make([]discovery.VerifyResult, 0)
	}
	if len(results) == 0 {
		return results
	}

	list := s.CameraInSystemCheck(ctx)

	existingIPs := make(map[string]bool)
	for _, cam := range list {
		if cam != nil {
			existingIPs[cam.IPAddress] = true
		}
	}

	for i := range results {
		if existingIPs[results[i].IP] {
			results[i].InSystem = true
		}
	}

	return results

}

func (s *APIServer) ApplyCamerasOnvifRecordInSystemCheck(ctx context.Context, results []*onvif.OnvifRecord) []*onvif.OnvifRecord {

	if(results == nil) {
		return make([]*onvif.OnvifRecord, 0)
	}
	if len(results) == 0 {
		return results
	}

	list := s.CameraInSystemCheck(ctx)

	existingIPs := make(map[string]bool)
	for _, cam := range list {
		if cam != nil {
			existingIPs[cam.IPAddress] = true
		}
	}

	for i := range results {
		if existingIPs[results[i].IP] {
			results[i].InSystem = true
		}
	}

	return results

}

func (s *APIServer) CameraInSystemCheck(ctx context.Context) []*models.Camera {

	var ll = s.logger.Lin("[CameraInSystemCheck]")

	list, err := s.Services.Camera.GetAllForInSystemCheck(ctx)
	if err != nil {
		ll.Error("Failed to get database check data: %v", err);
	}

	return list

}
