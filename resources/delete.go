package resources

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ianbrown78/gcp-nuke/config"
	"github.com/ianbrown78/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/cloudresourcemanager/v3"
)

// RemoveProjectResources  -
func RemoveProjectResources(config config.Config) {
	helpers.SetupCloseHandler()
	resourceMap := GetResourceMap(config)

	// Parallel deletion
	errs, _ := errgroup.WithContext(config.Context)

	for _, resource := range resourceMap {
		resource := resource
		errs.Go(func() error {
			log.Println("[Info] Retrieving list of resources for", resource.Name())
			resource.List(true)
			if config.NoDryRun {
				err := parallelResourceDeletion(resourceMap, resource, config)

				if err != nil {
					return err
				}
				return nil
			}

			parallelDryRun(resourceMap, resource, config)

			if config.NoKeepProject {
				err := deleteProject(config)

				if err != nil {
					return err
				}
			}

			return nil
		})
	}

	// Wait for all deletions to complete, and check for errors
	if err := errs.Wait(); err != nil {
		log.Fatal(err)
	}

	log.Printf("-- Deletion complete for project %v (dry-run: %v) (keep-project: %v) --\n", config.Project, config.NoDryRun, config.NoKeepProject)
}

func parallelResourceDeletion(resourceMap map[string]Resource, resource Resource, config config.Config) error {
	refreshCache := false
	if len(resource.List(false)) == 0 {
		log.Println("[Skipping] No", resource.Name(), "items to delete")
		return nil
	}

	timeOut := config.Timeout
	pollTime := config.PollTime
	seconds := 0

	// Wait for dependencies to delete
	for _, dependencyResourceName := range resource.Dependencies() {
		if seconds > timeOut {
			return fmt.Errorf("[Error] Resource %v timed out whilst waiting for dependency %v to delete. (%v seconds)", resource.Name(), dependencyResourceName, timeOut)
		}
		dependencyResource := resourceMap[dependencyResourceName]
		for len(dependencyResource.List(false)) != 0 {
			refreshCache = true
			time.Sleep(time.Duration(pollTime) * time.Second)
			seconds += pollTime
			log.Printf("[Waiting] Resource %v waiting for dependency %v to delete. (%v seconds)\n", resource.Name(), dependencyResource.Name(), seconds)
		}
	}

	if refreshCache {
		resource.List(refreshCache)
	}

	log.Println("[Remove] Removing", resource.Name(), "items:", resource.List(false))
	seconds = 0
	err := resource.Remove()

	// Unfortunately the API seems inconsistent with timings, so retry until any dependent resources delete
	for apiErrorCheck(err) {
		resource.List(true)

		if seconds > timeOut {
			return fmt.Errorf("[Error] Resource %v timed out whilst trying to delete. (%v seconds). Details of error below:\n %v", resource.Name(), timeOut, err.Error())
		}

		log.Printf("[Remove] In use Resource: %v. Items: %v. Waiting before retrying delete. (%v seconds)", resource.Name(), resource.List(false), seconds)
		time.Sleep(time.Duration(pollTime) * time.Second)
		seconds += pollTime
		err = resource.Remove()
	}

	// Add some info to the error
	if err != nil {
		detailedError := fmt.Errorf("[Error] Resource: %v. Items: %v. Details of error below:\n %v", resource.Name(), resource.List(false), err.Error())
		err = detailedError
	}

	return err
}

func deleteProject(config config.Config) error {
	ctx := config.Context
	client, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}

	deleteProjectCall := client.Projects.Delete(config.Project)
	deleteProject, err := deleteProjectCall.Do()

	if err != nil {
		return err
	}

	var updateOpStatus string
	seconds := 0
	for updateOpStatus != "DONE" {
		log.Printf("[Info] Removing project %v (%v seconds)", config.Project, seconds)

		operationCall := client.Operations.Get(deleteProject.Name)
		checkOpp, err := operationCall.Do()
		if err != nil {
			return err
		}

		if checkOpp.Done == true {
			updateOpStatus = "DONE"
		} else {
			updateOpStatus = "RUNNING"
		}

		time.Sleep(time.Duration(config.PollTime) * time.Second)
		seconds += config.PollTime
		if seconds > config.Timeout {
			return fmt.Errorf("[Error] Project removal timed out for %v (%v seconds)", config.Project, config.Timeout)
		}
	}
	log.Printf("[Info] Project removal completed for %v (%v seconds)", config.Project, seconds)

	return nil
}

// apiErrorCheck - Not proud of this workaround for the inconsistent api timings, suggestions welcome
func apiErrorCheck(err error) bool {
	if err == nil {
		return false
	}
	errorDescriptors := []string{
		"resourceInUseByAnotherResource",
		"resourceNotReady",
		// Interestingly in the case of instancegroups managed by GKE, listing them after deletion can often give back a ghost list
		"googleapi: Error 404",
	}
	for _, errorDesc := range errorDescriptors {
		if strings.Contains(err.Error(), errorDesc) {
			return true
		}
	}
	return false
}
