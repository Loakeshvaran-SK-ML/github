package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Creating constant variable
const SOURCE = "github"

// Parser function get the whole data from the queue and gives the required data
func Parser(wg *sync.WaitGroup, header_interface interface{}, data_interface interface{}) error {
	defer wg.Done()

	defer func() {
		if r := recover(); r != nil {
			log.WithField("error", r).Error("Panic occurred in Parser function")
		}
	}()

	if header_interface == nil {
		log.Error("Header is nil")
		return errors.New("Error occured: Header is nil !")
	}

	if data_interface == nil {
		log.Error("Data is nil")
		return errors.New("Error occured: Data is nil !")
	}

	header, ok := header_interface.(string)
	if !ok {
		log.Error("Header is not a string")
		return errors.New("Error occured: Header is not a string !")
	}
	body, ok := data_interface.(string)
	if !ok {
		log.Error("Data is not a string")
		return errors.New("Error occured: Data is not a string !")
	}

	if header == "" {
		log.Warn("Received empty header message")
		return errors.New("Error occured: Received empty header message!")
	}

	if body == "" {
		log.Warn("Received empty message body")
		return errors.New("Error occured: Received empty body message!")
	}

	//Storing the header content
	var headers map[string]interface{}
	if err := json.Unmarshal([]byte(header), &headers); err != nil {
		log.WithError(err).Error("Failed to unmarshal header")
		return errors.New("header is not a string")
	}

	//Storing the data(body) content
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		log.WithError(err).Error("Failed to unmarshal data")
		return errors.New("body is not a string")
	}

	eventType, ok := headers["X-Github-Event"].([]interface{})[0].(string)
	if !ok {
		log.Error("Failed to parse event type")
		return errors.New("Error occured!")
	}

	signature, ok := headers["X-Hub-Signature"].([]interface{})[0].(string)
	if !ok {
		log.Error("Failed to parse signature")
		return errors.New("Error occured!")
	}

	metadata := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &metadata); err != nil {
		log.WithError(err).Error("error metadata")
		return errors.New("Error occured!")
	}

	var time_created string
	var id string

	//Creating switch case for each of the event type
	switch eventType {
	case "push":
		time_created = metadata["head_commit"].(map[string]interface{})["timestamp"].(string)
		id = metadata["head_commit"].(map[string]interface{})["id"].(string)

	case "pull_request":
		time_created = metadata["pull_request"].(map[string]interface{})["updated_at"].(string)
		id = metadata["repository"].(map[string]interface{})["name"].(string) + "/" + strconv.Itoa(int(metadata["number"].(float64)))

	case "pull_request_review":
		time_created = metadata["review"].(map[string]interface{})["submitted_at"].(string)
		id = metadata["review"].(map[string]interface{})["id"].(string)

	case "pull_request_review_comment":
		time_created = metadata["comment"].(map[string]interface{})["updated_at"].(string)
		id = metadata["comment"].(map[string]interface{})["id"].(string)

	case "issues":
		time_created = metadata["issue"].(map[string]interface{})["updated_at"].(string)
		id = metadata["repository"].(map[string]interface{})["name"].(string) + "/" + strconv.Itoa(int(metadata["issue"].(map[string]interface{})["number"].(float64)))

	case "issue_comment":
		time_created = metadata["comment"].(map[string]interface{})["updated_at"].(string)
		id = metadata["comment"].(map[string]interface{})["id"].(string)

	case "check_run":
		time_created = metadata["check_run"].(map[string]interface{})["completed_at"].(string)
		if time_created == "" {
			time_created = metadata["check_run"].(map[string]interface{})["started_at"].(string)
		}
		id = metadata["check_run"].(map[string]interface{})["id"].(string)

	case "check_suite":
		time_created = metadata["check_suite"].(map[string]interface{})["updated_at"].(string)
		if time_created == "" {
			time_created = metadata["check_suite"].(map[string]interface{})["created_at"].(string)
		}
		id = metadata["check_suite"].(map[string]interface{})["id"].(string)

	case "deployment_status":
		time_created = metadata["deployment_status"].(map[string]interface{})["updated_at"].(string)
		id = metadata["deployment_status"].(map[string]interface{})["id"].(string)

	case "status":
		time_created = metadata["updated_at"].(string)
		id = metadata["id"].(string)

	case "release":
		time_created = metadata["release"].(map[string]interface{})["published_at"].(string)
		if time_created == "" {
			time_created = metadata["release"].(map[string]interface{})["created_at"].(string)
		}
		id = metadata["release"].(map[string]interface{})["id"].(string)

	default:
		log.Info("Received unknown event type: ", eventType)
		return errors.New("Error occured!")
	}

	//Storing the required data taken from the body
	githubEvent := map[string]interface{}{
		"event_type":   eventType,
		"id":           id,
		"metadata":     metadata,
		"time_created": time_created,
		"signature":    signature,
		"source":       SOURCE,
	}

	// Convert the map to a JSON string
	jsonData, err := json.Marshal(githubEvent)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to marshal event to JSON")
		return errors.New("Error occured!")
	}
	if jsonData == nil {
		return fmt.Errorf("json data is nil")
	}
	// print the output string
	output := string(jsonData)
	if output == "" {
		return errors.New("empty output string")
	}
	fmt.Println(output)
	return errors.New("Error occured!")

}
