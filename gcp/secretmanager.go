package gcp

import (
	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/secretmanager/v1"
	"log"
	"strings"
	"sync"
)

// SecretManager -
type SecretManagerSecrets struct {
	serviceClient *secretmanager.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	secretmanagerService, err := secretmanager.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	secretmanagerResource := SecretManagerSecrets{
		serviceClient: secretmanagerService,
	}
	register(&secretmanagerResource)
}

// Name - Name of the resourceLister for SecretManagerSecrets
func (c *SecretManagerSecrets) Name() string {
	return "SecretManagerSecrets"
}

// ToSlice - Name of the resourceLister for SecretManagerSecrets
func (c *SecretManagerSecrets) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)

}

// Setup - populates the struct
func (c *SecretManagerSecrets) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all SecretManagerSecrets
func (c *SecretManagerSecrets) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	// List all buckets in a project
	secretsListCall := c.serviceClient.Projects.Secrets.List("projects/" + c.base.config.Project)
	secretsList, err := secretsListCall.Do()
	if err != nil {
		// check if the API is enabled/
		if strings.Contains(err.Error(), "API has not been used in project") ||
			strings.Contains(err.Error(), "got HTTP response code 404") {
			log.Println("SecretManager API not enabled. Skipping.")
			return c.ToSlice()
		} else {
			// Otherwise, throw an error.
			log.Fatal(err)
		}
	}

	for _, secret := range secretsList.Secrets {
		instanceResource := DefaultResourceProperties{}
		c.resourceMap.Store(secret.Name, instanceResource)
	}

	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *SecretManagerSecrets) Dependencies() []string {
	return []string{}
}

// Remove -
func (c *SecretManagerSecrets) Remove() error {

	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		secretID := key.(string)

		// Parallel instance deletion
		errs.Go(func() error {

			// Delete the dataset
			secretDeleteCall := c.serviceClient.Projects.Secrets.Delete(secretID)
			_, err := secretDeleteCall.Do()
			if err != nil {
				return err
			}

			c.resourceMap.Delete(secretID)

			log.Printf("[Info] SecretManager secret deleted %v [type: %v project: %v]", secretID, c.Name(), c.base.config.Project)
			return nil
		})

		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
