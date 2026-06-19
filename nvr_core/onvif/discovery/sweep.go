package discovery

import (
	"context"
	"fmt"
	"sync"
	"nvr_core/network"
)

// SweepSubnet sends a unicast probe to all 254 IPs in a /24 subnet simultaneously.
// Example baseIP: "192.168.1."
func (v *Verifier) SweepSubnet(ctx context.Context, baseIP string) []VerifyResult {
	return v.SweepSubnetRange(ctx, baseIP, 1, 254)
}

func (v *Verifier) SweepSubnetIPRange(ctx context.Context, fromip string, toip string) ([]VerifyResult, error) {

	baseIP, from, err1 := network.GetSubnetParts(fromip)
	if err1 != nil {
		return nil, err1
	}

	_, to, err2 := network.GetSubnetParts(toip)
	if err2 != nil {
		return nil, err2
	}

	return v.SweepSubnetRange(ctx, baseIP, from, to), nil

}

func (v *Verifier) SweepSubnetRange(ctx context.Context, baseIP string, from int, to int) []VerifyResult {
	var wg sync.WaitGroup
	resultsChan := make(chan VerifyResult, 255)

	// Sweep IPs .1 through .254
	for i := from; i <= to; i++ {
		targetIP := fmt.Sprintf("%s%d", baseIP, i)

		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			// Use the UnicastProbe method we built earlier
			isONVIF, rawData := v.unicastProbe(ip)
			if isONVIF {
				resultsChan <- VerifyResult{
					IP: ip,
					IsValid:   true,
					Protocol:  "onvif",
					PortFound: 3702,
					RawData:   rawData, // Contains the XAddrs
				}
			}
		}(targetIP)
	}

	// Wait for all probes to finish or timeout
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var foundCameras []VerifyResult
	for result := range resultsChan {
		foundCameras = append(foundCameras, result)
	}

	return foundCameras
}