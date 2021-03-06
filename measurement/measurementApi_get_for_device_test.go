package measurement

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestMeasurementApi_GetForDevice_ExistingId(t *testing.T) {
	// given: A test server
	ts := buildHttpServer(200, fmt.Sprintf(measurementCollectionTemplate, measurement))
	defer ts.Close()

	// and: the api as system under test
	api := buildMeasurementApi(ts.URL)

	collection, err := api.GetForDevice(deviceId, 5)

	if err != nil {
		t.Fatalf("GetForDevice() got an unexpected error: %s", err.Error())
	}

	if collection == nil {
		t.Fatalf("GetForDevice() got no explict error but the collection was nil.")
	}

	if len(collection.Measurements) != 1 {
		t.Fatalf("GetForDevice() measurements count = %v, want %v", len(collection.Measurements), 1)
	}

	measurement := collection.Measurements[0]
	if measurement.Id != measurementId {
		t.Errorf("GetForDevice() measurement id = %v, want %v", measurement.Id, measurementId)
	}

	assertMetricsOfMeasurement(measurement.Metrics, t)
}

func TestMeasurementApi_GetForDevice_HandlesPageSize(t *testing.T) {
	tests := []struct {
		name        string
		pageSize    int
		errExpected bool
	}{
		{"Negative", -1, true},
		{"Zero", 0, true},
		{"Min", 1, false},
		{"Max", 2000, false},
		{"too large", 2001, true},
		{"in range", 10, false},
	}

	// given: A test server
	var capturedUrl string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUrl = r.URL.String()
		_, _ = w.Write([]byte(fmt.Sprintf(measurementCollectionTemplate, measurement)))
	}))
	defer ts.Close()

	// and: the api as system under test
	api := buildMeasurementApi(ts.URL)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := api.GetForDevice(deviceId, tt.pageSize)

			if tt.errExpected {
				if err == nil {
					t.Error("GetForDevice() error expected but was nil")
				}
			}

			if !tt.errExpected {
				contains := strings.Contains(capturedUrl, fmt.Sprintf("pageSize=%d", tt.pageSize))

				if !contains {
					t.Errorf("GetForDevice() expected pageSize '%d' in url. '%s' given", tt.pageSize, capturedUrl)
				}
			}
		})
	}
}

func TestMeasurementApi_GetForDevice_NotExistingId(t *testing.T) {
	// given: A test server
	ts := buildHttpServer(200, fmt.Sprintf(measurementCollectionTemplate, ""))
	defer ts.Close()

	// and: the api as system under test
	api := buildMeasurementApi(ts.URL)

	collection, err := api.GetForDevice(deviceId, 5)

	if err != nil {
		t.Fatalf("GetForDevice() got an unexpected error: %s", err.Error())
		return
	}

	if collection == nil {
		t.Fatalf("GetForDevice() got no explict error but the collection was nil.")
		return
	}

	if len(collection.Measurements) != 0 {
		t.Fatalf("GetForDevice() measurements count = %v, want %v", len(collection.Measurements), 0)
	}
}

func TestMeasurementApi_GetForDevice_MalformedResponse(t *testing.T) {
	// given: A test server
	ts := buildHttpServer(200, "{ foo ...")
	defer ts.Close()

	// and: the api as system under test
	api := buildMeasurementApi(ts.URL)

	_, err := api.GetForDevice(deviceId, 5)

	if err == nil {
		t.Errorf("GetForDevice() Expected error - non given")
		return
	}
}
