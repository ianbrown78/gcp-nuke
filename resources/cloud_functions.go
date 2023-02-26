package resources

import (
	"fmt"
	"google.golang.org/api/cloudfunctions/v2"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
)

// FunctionsInstances -
type FunctionsInstances struct {
	serviceClient *cloudfunctions.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	functionsService, err := cloudfunctions.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	functionsResource := FunctionsInstances{
		serviceClient: functionsService,
	}
	register(&functionsResource)
}

// Name - Name of the resourceLister for FunctionsInstances
func (c *FunctionsInstances) Name() string {
	return "FunctionsInstances"
}

// ToSlice - Name of the resourceLister for FunctionsInstances
func (c *FunctionsInstances) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)

}

// Setup - populates the struct
func (c *FunctionsInstances) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all FunctionsInstances
func (c *FunctionsInstances) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	// Get the list of locations for the project.
	locationsListCall := c.serviceClient.Projects.Locations.List("projects/" + c.base.config.Project)
	locationsList, err := locationsListCall.Do()
	if err != nil {
		// check if the API is enabled/
		if !strings.Contains(err.Error(), "API has not been used in project") {
			// Otherwise, throw an error.
			log.Fatal(err)
		} else {
			log.Printf("SQLAdmin API not enabled in project %v. Skipping.", c.base.config.Project)
			return c.ToSlice()
		}
	}

	// Get the list of functions by location and project.
	for _, location := range locationsList.Locations {
		functionsListCall := c.serviceClient.Projects.Locations.Functions.List(
			"projects/" + c.base.config.Project + "/locations/" + location.LocationId)
		functionsList, err := functionsListCall.Do()
		if err != nil {
			log.Fatal(err)
		}

		// Add functions to the resourceMap.
		for _, function := range functionsList.Functions {
			instanceResource := DefaultResourceProperties{
				zone: location.LocationId,
			}
			c.resourceMap.Store(function.Name, instanceResource)
		}
	}
	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *FunctionsInstances) Dependencies() []string {
	return []string{}
}

// Remove -
func (c *FunctionsInstances) Remove() error {

	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		functionID := key.(string)
		location := value.(DefaultResourceProperties).zone

		// Parallel instance deletion
		errs.Go(func() error {
			deleteCall := c.serviceClient.Projects.Locations.Functions.Delete(
				"projects/" + c.base.config.Project +
					"/locations/" + location +
					"/functions/" + functionID)
			operation, err := deleteCall.Do()
			if err != nil {
				return err
			}
			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being deleted %v [type: %v project: %v region: %v] (%v seconds)", functionID, c.Name(), c.base.config.Project, location, seconds)

				operationCall := c.serviceClient.Projects.Locations.Operations.Get(operation.Name)
				checkOpp, err := operationCall.Do()
				if err != nil {
					return err
				}
				opStatus = string(checkOpp.Response)

				time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
				seconds += c.base.config.PollTime
				if seconds > c.base.config.Timeout {
					return fmt.Errorf("[Error] Resource deletion timed out for %v [type: %v project: %v region: %v] (%v seconds)", functionID, c.Name(), c.base.config.Project, location, c.base.config.Timeout)
				}
			}
			c.resourceMap.Delete(functionID)

			log.Printf("[Info] Resource deleted %v [type: %v project: %v region: %v] (%v seconds)", functionID, c.Name(), c.base.config.Project, location, seconds)
			return nil
		})

		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
