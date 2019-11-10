package asana

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go"
	asana "github.com/oms-services/asana/pkg/asana/v1"
	result "github.com/oms-services/asana/result"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Payload struct {
	EventId     string   `json:"eventID"`
	EventType   string   `json:"eventType"`
	ContentType string   `json:"contentType"`
	Data        DataArgs `json:"data"`
}

type Subscribe struct {
	Data      DataArgs `json:"data"`
	Endpoint  string   `json:"endpoint"`
	ID        string   `json:"id"`
	IsTesting bool     `json:"istesting"`
}

type DataArgs struct {
	ProjectID   string `json:"projectId"`
	WorkspaceID string `json:"workspaceId"`
	Existing    bool   `json:"existing"`
}

type AsanaArgument struct {
	TaskID    string `json:"taskId,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
}

type Message struct {
	Success    string `json:"success"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

var Listener = make(map[string]Subscribe)
var rtmStarted bool
var isExistingPrinted bool
var client *asana.Client
var oldTask *asana.Task

//CreateProject asana
func CreateProject(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param *asana.ProjectRequest
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	param.Layout = asana.ListLayout

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	project, projectErr := client.CreateProject(param)
	if projectErr != nil {
		result.WriteErrorResponseString(responseWriter, projectErr.Error())
		return
	}

	bytes, _ := json.Marshal(project)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//CreateTask asana
func CreateTask(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param *asana.TaskRequest
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	param.ProjectID = param.TempProjectID
	task, taskErr := client.CreateTask(param)
	if taskErr != nil {
		result.WriteErrorResponseString(responseWriter, taskErr.Error())
		return
	}
	bytes, _ := json.Marshal(task)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)

}

//DeleteProject asana
func DeleteProject(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param AsanaArgument
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	if err := client.DeleteProjectByID(param.ProjectID); err != nil {
		result.WriteErrorResponseString(responseWriter, err.Error())
		return
	}

	message := Message{"true", "Project deleted successfully", http.StatusOK}
	bytes, _ := json.Marshal(message)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//DeleteTask asana
func DeleteTask(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param AsanaArgument
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	if err := client.DeleteTask(param.TaskID); err != nil {
		fmt.Println("err ::", err)
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	message := Message{"true", "Task deleted successfully", http.StatusOK}
	bytes, _ := json.Marshal(message)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//ListTask asana
func ListTask(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param *asana.TaskRequest
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}
	var allTasks []*asana.Task
	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	taskPagesChan, _ := client.ListMyTasks(param)

	pageCount := 0
	for page := range taskPagesChan {
		if err := page.Err; err != nil {
			result.WriteErrorResponseString(responseWriter, err.Error())
			return
		}

		for _, task := range page.Tasks {
			taskDetails, _ := client.FindTaskByID(strconv.FormatInt(task.ID, 10))
			allTasks = append(allTasks, taskDetails)
		}
		pageCount++
	}
	bytes, _ := json.Marshal(allTasks)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//ListWorkspace asana
func ListWorkspace(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}
	workspacesChan, _ := client.ListMyWorkspaces()

	var listWorkspace []*asana.Workspace

	pageCount := 0
	for page := range workspacesChan {
		if err := page.Err; err != nil {
			continue
		}

		for _, workspace := range page.Workspaces {
			listWorkspace = append(listWorkspace, workspace)
		}
		pageCount += 1
	}
	bytes, _ := json.Marshal(listWorkspace)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)

}

//FindTask asana
func FindTask(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param AsanaArgument
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	taskDetails, err := client.FindTaskByID(param.TaskID)
	if err != nil {
		result.WriteErrorResponseString(responseWriter, err.Error())
		return
	}

	bytes, _ := json.Marshal(taskDetails)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//FindProject asana
func FindProject(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param AsanaArgument
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	projectDetails, err := client.FindProjectByID(param.ProjectID)
	if err != nil {
		result.WriteErrorResponseString(responseWriter, err.Error())
		return
	}
	bytes, _ := json.Marshal(projectDetails)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//UpdateProject asana
func UpdateProject(responseWriter http.ResponseWriter, request *http.Request) {
	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param *asana.ProjectRequest
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	update, updateErr := client.UpdateProject(param)
	if updateErr != nil {
		result.WriteErrorResponseString(responseWriter, updateErr.Error())
		return
	}

	bytes, _ := json.Marshal(update)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//ListProjectTasks asana
func ListProjectTasks(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	decoder := json.NewDecoder(request.Body)

	var param *asana.TaskRequest
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
		result.WriteErrorResponse(responseWriter, decodeErr)
		return
	}

	var allTasks []*asana.Task
	client, err := asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	param.ProjectID = param.TempProjectID
	taskPagesChan, _, _ := client.ListTasksForProject(param)

	pageCount := 0
	for page := range taskPagesChan {
		if err := page.Err; err != nil {
			result.WriteErrorResponseString(responseWriter, err.Error())
			return
		}

		for _, task := range page.Tasks {
			taskDetails, _ := client.FindTaskByID(strconv.FormatInt(task.ID, 10))
			allTasks = append(allTasks, taskDetails)
		}
		pageCount++
	}

	bytes, _ := json.Marshal(allTasks)
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//SubscribeTasks asana
func SubscribeTasks(responseWriter http.ResponseWriter, request *http.Request) {

	var accessToken = os.Getenv("ACCESS_TOKEN")

	var sub Subscribe
	decoder := json.NewDecoder(request.Body)
	decodeError := decoder.Decode(&sub)
	if decodeError != nil {
		result.WriteErrorResponse(responseWriter, decodeError)
		return
	}

	if sub.Data.ProjectID == "" && sub.Data.WorkspaceID == "" {
		message := Message{"false", "Please provide Project ID or Workspace ID", http.StatusBadRequest}
		bytes, _ := json.Marshal(message)
		result.WriteJsonResponse(responseWriter, bytes, http.StatusBadRequest)
		return
	}

	var err error
	client, err = asana.NewClient(accessToken)
	if err != nil {
		result.WriteErrorResponse(responseWriter, err)
		return
	}

	Listener[sub.Data.WorkspaceID] = sub
	if !rtmStarted {
		go RTSAsana()
		rtmStarted = true
	}

	bytes, _ := json.Marshal("Subscribed")
	result.WriteJsonResponse(responseWriter, bytes, http.StatusOK)
}

//RTSAsana function
func RTSAsana() {
	isTest := false
	for {
		if len(Listener) > 0 {
			for WorkspaceID, Sub := range Listener {
				go getMessageUpdates(WorkspaceID, Sub.Data.ProjectID, Sub, Sub.Data.Existing)
				isTest = Sub.IsTesting
			}
		} else {
			rtmStarted = false
			break
		}
		time.Sleep(5 * time.Second)
		if isTest {
			break
		}
	}
}

func getMessageUpdates(workspaceID, projectID string, sub Subscribe, existing bool) {

	var finalTask *asana.Task
	var finalTasks []*asana.Task
	var allTasks []*asana.Task

	if projectID != "" {

		taskPagesChan, _, err := client.ListTasksForProject(&asana.TaskRequest{
			ProjectID: projectID,
		})
		if err != nil {
			log.Fatal(err)
		}

		pageCount := 0
		for page := range taskPagesChan {
			if err := page.Err; err != nil {
				log.Fatalln("Error : ", err)
				return
			}

			for _, task := range page.Tasks {
				taskDetails, _ := client.FindTaskByID(strconv.FormatInt(task.ID, 10))
				allTasks = append(allTasks, taskDetails)
			}
			pageCount++
		}
	} else {
		taskPagesChan, err := client.ListMyTasks(&asana.TaskRequest{
			Workspace: workspaceID,
		})
		if err != nil {
			log.Fatal(err)
		}

		pageCount := 0
		for page := range taskPagesChan {
			if err := page.Err; err != nil {
				log.Fatalln("Error : ", err)
				return
			}

			for _, task := range page.Tasks {
				taskDetails, _ := client.FindTaskByID(strconv.FormatInt(task.ID, 10))
				allTasks = append(allTasks, taskDetails)
			}
			pageCount++
		}
	}

	if isExistingPrinted == false {
		if existing {
			finalTasks = allTasks
		} else {
			if finalTask == nil {
				finalTask = latestTask(allTasks)
			}
		}
	} else {
		finalTask = latestTask(allTasks)
	}

	contentType := "application/json"

	t, err := cloudevents.NewHTTPTransport(cloudevents.WithTarget(sub.Endpoint), cloudevents.WithStructuredEncoding())
	if err != nil {
		log.Printf("failed to create transport, %v", err)
		return
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow())
	if err != nil {
		log.Printf("failed to create client, %v", err)
		return
	}

	source, err := url.Parse(sub.Endpoint)
	event := cloudevents.Event{
		Context: cloudevents.EventContextV01{
			EventID:     sub.ID,
			EventType:   "task",
			Source:      cloudevents.URLRef{URL: *source},
			ContentType: &contentType,
		}.AsV01(),
		Data: "",
	}

	if finalTasks != nil {
		event.Data = finalTasks

	} else {
		event.Data = finalTask
	}

	if oldTask == nil && finalTask != nil {
		oldTask = finalTask
	}

	if existing && !isExistingPrinted {
		resp, evt, err := c.Send(context.Background(), event)
		if err != nil {
			log.Printf("failed to send: %v (%v)", err, evt)
		}
		fmt.Printf("Response1: \n%s\n", resp)
		finalTasks = nil
		isExistingPrinted = true

	} else if oldTask != nil && finalTask.ID != oldTask.ID {
		resp, evt, err := c.Send(context.Background(), event)
		if err != nil {
			log.Printf("failed to send: %v (%v)", err, evt)
		}
		fmt.Printf("Response2: \n%s\n", resp)
		oldTask = finalTask
		finalTask = nil
	}
}

func latestTask(tasks []*asana.Task) *asana.Task {
	if len(tasks) == 0 {
		return nil
	}
	latest := tasks[0]
	for _, task := range tasks {
		if task.ID > latest.ID {
			latest = task
		}
	}
	return latest
}
