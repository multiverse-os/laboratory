package cli

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bom-maker/bomcsv"
	"github.com/bom-maker/core"
	"github.com/bom-maker/mouser"
	"github.com/bom-maker/mouser/model"
)

// createUberParts retrieves each and every part from the Mouser API and creates a slice of UberPart
func createUberParts(mouserAPIKey string, csvParts []bomcsv.CSVPart) []core.UberPart {
	// Mouser API calls on all parts (multithreaded)
	api := mouser.NewAPI(mouserAPIKey)
	var wg sync.WaitGroup
	mutex := sync.Mutex{}
	parts := make([]core.UberPart, 0)
	for _, v := range csvParts {
		//fmt.Printf("Calling Mouser API for part(s) %s (Device=%s, MouserRef=%s)\n", v.Parts, v.Device, v.MouserRef)
		wg.Add(1)
		go func(wg *sync.WaitGroup, csvPart bomcsv.CSVPart) {
			defer wg.Done()

			// If rate limited, retry
			var err error
			var p *model.Part
			for {
				p, err = api.SearchByPartNumber(csvPart.MouserRef)
				if err != nil {
					if _, ok := err.(mouser.ErrorRateLimited); ok {
						continue // Rate limited, retrying
					}
				}

				break
			}

			if err != nil {
				// this is just a warning for that part, not an error per se
				fmt.Fprintf(os.Stderr, "An error occurred looking for MouserRef %s, err=%+v\n", csvPart.MouserRef, err)
			} else {
				mutex.Lock()
				parts = append(parts, core.UberPart{
					Part:    *p,
					CSVPart: csvPart,
				})
				mutex.Unlock()
			}

		}(&wg, v)

		// Wait to avoid overloading the API
		time.Sleep(400 * time.Millisecond)
	} // for

	wg.Wait() // Wait for all go routines to finish}
	return parts
}
