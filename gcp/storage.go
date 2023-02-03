package gcp

import (
	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/storage/v1"
	"log"
	"strings"
	"sync"
)

// StorageBuckets -
type StorageBuckets struct {
	serviceClient *storage.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	storageService, err := storage.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	storageResource := StorageBuckets{
		serviceClient: storageService,
	}
	register(&storageResource)
}

// Name - Name of the resourceLister for StorageBuckets
func (c *StorageBuckets) Name() string {
	return "StorageBuckets"
}

// ToSlice - Name of the resourceLister for StorageBuckets
func (c *StorageBuckets) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)

}

// Setup - populates the struct
func (c *StorageBuckets) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all StorageBuckets
func (c *StorageBuckets) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	// List all buckets in a project
	bucketsListCall := c.serviceClient.Buckets.List(c.base.config.Project)
	bucketsList, err := bucketsListCall.Do()
	if err != nil {
		// check if the API is enabled/
		if !strings.Contains(err.Error(), "API has not been used in project") {
			// Otherwise, throw an error.
			log.Fatal(err)
		} else {
			log.Println("Storage API not enabled. Skipping.")
			return c.ToSlice()
		}
	}

	for _, instance := range bucketsList.Items {
		instanceResource := DefaultResourceProperties{}
		c.resourceMap.Store(instance.Name, instanceResource)
	}

	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *StorageBuckets) Dependencies() []string {
	return []string{}
}

// Remove -
func (c *StorageBuckets) Remove() error {

	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		bucketID := key.(string)
		zone := value.(DefaultResourceProperties).zone

		// Check if there is a retention period or lock on the bucket
		bucketCall := c.serviceClient.Buckets.Get(bucketID)
		bucket, _ := bucketCall.Do()
		policy := bucket.RetentionPolicy

		log.Printf("Removing bucket %v", bucketID)
		if policy != nil && policy.IsLocked == true {
			// throw an error about the bucket being locked.
			log.Printf("Bucket %v has a bucket policy that is currently locked.", bucketID)
			return true
		}
		if policy != nil && policy.RetentionPeriod > 0 {
			// throw an error about the retention policy being not zero.
			log.Printf("Bucket %v has a bucket policy retention period of %v seconds.", bucketID, policy.RetentionPeriod)
			log.Printf("Bucket %v retention policy will be updated to 0 seconds.", bucketID)

			bucket, _ := c.serviceClient.Buckets.Get(bucketID).Do()
			bucket.RetentionPolicy.RetentionPeriod = 0
			bucketUpdateCall := c.serviceClient.Buckets.Patch(bucketID, bucket)
			_, err := bucketUpdateCall.Do()
			if err != nil {
				return false
			}
		}

		// Parallel instance deletion
		errs.Go(func() error {
			// Get objects
			objectsListCall := c.serviceClient.Objects.List(bucketID)
			objectsList, _ := objectsListCall.Do()

			// Delete objects
			for _, object := range objectsList.Items {
				objectName := object.Name
				objectDeleteCall := c.serviceClient.Objects.Delete(bucketID, objectName)
				err := objectDeleteCall.Do()
				if err != nil {
					return err
				}
			}

			// Now delete the bucket
			deleteCall := c.serviceClient.Buckets.Delete(bucketID)
			err := deleteCall.Do()
			if err != nil {
				return err
			}

			c.resourceMap.Delete(bucketID)

			log.Printf("[Info] Bucket deleted %v [type: %v project: %v zone: %v]", bucketID, c.Name(), c.base.config.Project, zone)
			return nil
		})

		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
