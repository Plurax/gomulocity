package gomulocity_event

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func NewEventsApi(client Client) Events {
	return &events{client, "/event/events"}
}

type Events interface {
	CreateEvent(event *CreateEvent) error
	UpdateEvent(eventId string, event *UpdateEvent) error
	DeleteEvent(eventId string) error

	Get(eventId string) (*Event, error)
	GetForDevice(source string) (*EventCollection, error)
	Find(query EventQuery) (*EventCollection, error)
	NextPage(c *EventCollection) (*EventCollection, error)
	PreviousPage(c *EventCollection) (*EventCollection, error)
}

type EventQuery struct {
	DateFrom     *time.Time
	DateTo       *time.Time
	FragmentType string
	Type         string
	Source       string
}

func (q EventQuery) QueryParams() string {
	params := url.Values{}

	if q.DateFrom != nil {
		params.Add("dateFrom", q.DateFrom.Format(time.RFC3339))
	}

	if q.DateTo != nil {
		params.Add("dateTo", q.DateTo.Format(time.RFC3339))
	}

	if len(q.FragmentType) > 0 {
		params.Add("fragmentType", q.FragmentType)
	}

	if len(q.Type) > 0 {
		params.Add("type", q.Type)
	}

	if len(q.Source) > 0 {
		params.Add("source", q.Source)
	}

	return params.Encode()
}

type events struct {
	client   Client
	basePath string
}

func (e *events) CreateEvent(event *CreateEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error while marhalling the event: %s", err.Error())
	}

	body, status, err := e.client.post(e.basePath, bytes)
	if err != nil {
		log.Printf("Error while posting a new event: %s", err.Error())
		return err
	}
	if status != http.StatusCreated {
		return createErrorFromResponse(body)
	}

	return nil
}

func (e *events) UpdateEvent(eventId string, event *UpdateEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error while marhalling the update event: %s", err.Error())
	}
	path := fmt.Sprintf("%s/%s", e.basePath, url.QueryEscape(eventId))

	body, status, err := e.client.put(path, bytes)
	if status != http.StatusOK {
		return createErrorFromResponse(body)
	}

	return err
}

func (e *events) DeleteEvent(eventId string) error {
	body, status, err := e.client.delete(fmt.Sprintf("%s/%s", e.basePath, url.QueryEscape(eventId)))

	if status != http.StatusNoContent {
		return createErrorFromResponse(body)
	}

	return err
}

func (e *events) Get(eventId string) (*Event, error) {
	body, status, err := e.client.get(fmt.Sprintf("%s/%s", e.basePath, url.QueryEscape(eventId)))

	if status != http.StatusOK {
		log.Printf("Event with id %s was not found", eventId)
		return nil, nil
	}

	var result Event
	if len(body) > 0 {
		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Printf("Error while parsing response JSON: %s", err.Error())
			return nil, err
		}
	} else {
		return nil, errors.New("GetEvent: response body was empty")
	}

	return &result, nil
}

func (e *events) GetForDevice(source string) (*EventCollection, error) {
	return e.Find(EventQuery{Source: source})
}

func (e *events) Find(query EventQuery) (*EventCollection, error) {
	return e.getCommon(fmt.Sprintf("%s?%s", e.basePath, query.QueryParams()))
}

func (e *events) NextPage(c *EventCollection) (*EventCollection, error) {
	return e.getPage(c.Next)
}

func (e *events) PreviousPage(c *EventCollection) (*EventCollection, error) {
	return e.getPage(c.Prev)
}

func (e *events) getPage(reference string) (*EventCollection, error) {
	if reference == "" {
		log.Print("No page reference given. Returning nil.")
		return nil, nil
	}

	nextUrl, err := url.Parse(reference)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unparsable URL given for page reference: '%s'", reference))
	}

	return e.getCommon(fmt.Sprintf("%s?%s", nextUrl.Path, nextUrl.RawQuery))
}

// -- internal

func (e *events) getCommon(path string) (*EventCollection, error) {
	body, status, err := e.client.get(path)

	if status != http.StatusOK {
		return nil, createErrorFromResponse(body)
	}

	var result EventCollection
	if len(body) > 0 {
		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Printf("Error while parsing response JSON: %s", err.Error())
			return nil, err
		}
	} else {
		return nil, errors.New("GetCollection: response body was empty")
	}

	return &result, nil
}

func createErrorFromResponse(responseBody []byte) error {
	var msg map[string]interface{}
	_ = json.Unmarshal(responseBody, &msg)
	return errors.New(fmt.Sprintf("Request failed. Server returns error: {%s: %s}", msg["error"], msg["message"]))
}
