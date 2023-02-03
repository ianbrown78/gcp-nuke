package gcp

import (
	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/bigquery/v2"
	"log"
	"strings"
	"sync"
)

// BigQuery -
type BigQueryDatasets struct {
	serviceClient *bigquery.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	bigqueryService, err := bigquery.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	bigqueryResource := BigQueryDatasets{
		serviceClient: bigqueryService,
	}
	register(&bigqueryResource)
}

// Name - Name of the resourceLister for BigQueryDatasets
func (c *BigQueryDatasets) Name() string {
	return "BigQueryDatasets"
}

// ToSlice - Name of the resourceLister for BigQueryDatasets
func (c *BigQueryDatasets) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)

}

// Setup - populates the struct
func (c *BigQueryDatasets) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all BigQueryDatasets
func (c *BigQueryDatasets) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	// List all buckets in a project
	datasetsListCall := c.serviceClient.Datasets.List(c.base.config.Project)
	datasetsList, err := datasetsListCall.Do()
	if err != nil {
		// check if the API is enabled/
		if strings.Contains(err.Error(), "API has not been used in project") {
			log.Println("BigQuery API not enabled. Skipping.")
			return c.ToSlice()
		} else {
			// Otherwise, throw an error.
			log.Fatal(err)
		}
	}

	for _, dataset := range datasetsList.Datasets {
		instanceResource := DefaultResourceProperties{}
		c.resourceMap.Store(dataset.DatasetReference, instanceResource)
	}

	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *BigQueryDatasets) Dependencies() []string {
	return []string{}
}

// Remove -
func (c *BigQueryDatasets) Remove() error {

	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		datasetID := key.(string)

		// Parallel instance deletion
		errs.Go(func() error {

			// Delete the dataset
			datasetContentsDeleteCall := c.serviceClient.Datasets.Delete(c.base.config.Project, datasetID)
			datasetContentsDeleteCall.DeleteContents(true)
			err := datasetContentsDeleteCall.Do()
			if err != nil {
				return err
			}

			c.resourceMap.Delete(datasetID)

			log.Printf("[Info] BigQuery dataset deleted %v [type: %v project: %v]", datasetID, c.Name(), c.base.config.Project)
			return nil
		})

		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
