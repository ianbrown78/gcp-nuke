package gcp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/sqladmin/v1beta4"
)

// SQLInstances -
type SQLInstances struct {
	serviceClient *sqladmin.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	sqlService, err := sqladmin.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	sqlResource := SQLInstances{
		serviceClient: sqlService,
	}
	register(&sqlResource)
}

// Name - Name of the resourceLister for SqlInstances
func (c *SQLInstances) Name() string {
	return "SqlInstances"
}

// ToSlice - Name of the resourceLister for SqlInstances
func (c *SQLInstances) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)

}

// Setup - populates the struct
func (c *SQLInstances) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all SqlInstances
func (c *SQLInstances) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	instanceListCall := c.serviceClient.Instances.List(c.base.config.Project)
	instanceList, err := instanceListCall.Do()
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

	for _, instance := range instanceList.Items {

		instanceResource := DefaultResourceProperties{
			protected: instance.Settings.DeletionProtectionEnabled,
		}
		c.resourceMap.Store(instance.Name, instanceResource)
	}
	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *SQLInstances) Dependencies() []string {
	return []string{}
}

// Remove -
func (c *SQLInstances) Remove() error {

	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		instanceID := key.(string)
		zone := value.(DefaultResourceProperties).zone
		protected := value.(DefaultResourceProperties).protected

		// Parallel instance deletion
		errs.Go(func() error {
			// Check if instance protection is enabled. If so, disable it.
			if protected == true {
				log.Printf("SQL instance %v has deletion protection enabled. Disabling", instanceID)
				instanceCall := c.serviceClient.Instances.Get(c.base.config.Project, instanceID)
				instance, err := instanceCall.Do()
				if err != nil {
					log.Fatalf("Could not get CloudSQL instance %v", instanceID)
				}

				instance.Settings.DeletionProtectionEnabled = false
				instanceUpdateCall := c.serviceClient.Instances.Update(c.base.config.Project, instance.Name, instance)
				updateOp, err := instanceUpdateCall.Do()
				if err != nil {
					log.Fatal(err)
				}
				var updateOpStatus string
				seconds := 0
				for updateOpStatus != "DONE" {
					log.Printf("[Info] Removing deletion protection for %v (%v seconds)", instanceID, seconds)

					operationCall := c.serviceClient.Operations.Get(c.base.config.Project, updateOp.Name)
					checkOpp, err := operationCall.Do()
					if err != nil {
						return err
					}
					updateOpStatus = checkOpp.Status

					time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
					seconds += c.base.config.PollTime
					if seconds > c.base.config.Timeout {
						return fmt.Errorf("[Error] Resource deletionprotection removal timed out for %v (%v seconds)", instanceID, c.base.config.Timeout)
					}
				}
				log.Printf("[Info] Deletion protection removal completed for %v (%v seconds)", instanceID, seconds)
			}

			deleteCall := c.serviceClient.Instances.Delete(c.base.config.Project, instanceID)
			operation, err := deleteCall.Do()
			if err != nil {
				return err
			}
			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being deleted %v [type: %v project: %v zone: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, zone, seconds)

				operationCall := c.serviceClient.Operations.Get(c.base.config.Project, operation.Name)
				checkOpp, err := operationCall.Do()
				if err != nil {
					return err
				}
				opStatus = checkOpp.Status

				time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
				seconds += c.base.config.PollTime
				if seconds > c.base.config.Timeout {
					return fmt.Errorf("[Error] Resource deletion timed out for %v [type: %v project: %v zone: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, zone, c.base.config.Timeout)
				}
			}
			c.resourceMap.Delete(instanceID)

			log.Printf("[Info] Resource deleted %v [type: %v project: %v zone: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, zone, seconds)
			return nil
		})

		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
