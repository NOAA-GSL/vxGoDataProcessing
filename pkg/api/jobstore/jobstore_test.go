package jobstore

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJobStore(t *testing.T) {
	tests := []struct {
		name string
		want *JobStore
	}{
		// test cases
		{
			name: "Test New Job Store",
			want: &JobStore{
				jobs:   map[int]Job{},
				nextID: 0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NewJobStore(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewJobStore() = %v, want %v", got, test.want)
			}
		})
	}
}

// TODO - test concurrency

func TestJobStore_CreateJob(t *testing.T) {
	t.Run("Test creating a job", func(t *testing.T) {
		js := NewJobStore()
		want := Job{
			ID:     0,
			DocID:  "foo",
			Status: "created",
		}

		_, _ = js.CreateJob("foo")

		got, _ := js.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.CreateJob() = %v, want %v", got, want)
		}
	})

	t.Run("Test creating a second job", func(t *testing.T) {
		js := NewJobStore()
		want := Job{
			ID:     1,
			DocID:  "bar",
			Status: "created",
		}

		_, _ = js.CreateJob("foo")
		_, _ = js.CreateJob("bar")

		got, _ := js.GetJob(1)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.CreateJob() = %v, want %v", got, want)
		}
	})

	t.Run("Test creating an empty job", func(t *testing.T) {
		js := NewJobStore()
		want := "expected a non-empty docID"

		_, err := js.CreateJob("")
		if err == nil {
			t.Error("JobStore.CreateJob() - expected an error but got none")
		}

		if err.Error() != want {
			t.Errorf("JobStore.updateJobStatus got error '%v', wanted error '%v'", err.Error(), want)
		}
	})

	// TODO - what happens if we get the same docID submitted multiple times?
}

func TestJobStore_GetJob(t *testing.T) {
	t.Run("Test getting a nonexistant job", func(t *testing.T) {
		js := NewJobStore()

		jobID := 0
		_, err := js.GetJob(jobID)
		if err == nil {
			t.Error("JobStore.GetJob() Wanted an error but didn't get one")
			return
		}

		wantedErr := fmt.Sprintf("job with id=%v not found", jobID)
		if err != nil && err.Error() != wantedErr {
			t.Errorf("JobStore.GetJob() got error: '%v', want error: '%v'", err, wantedErr)
			return
		}
	})

	t.Run("Test getting a job", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")

		want := Job{
			ID:     0,
			DocID:  "foo",
			Status: "created",
		}
		got, _ := js.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.GetJob() = %v, want %v", got, want)
		}
	})

	t.Run("Test getting the correct job", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")
		_, _ = js.CreateJob("bar")

		want := Job{
			ID:     0,
			DocID:  "foo",
			Status: "created",
		}
		got, _ := js.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.GetJob() = %v, want %v", got, want)
		}
	})
}

func TestJobStore_GetAllJobs(t *testing.T) {
	t.Run("Test getting multiple jobs", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")
		_, _ = js.CreateJob("bar")

		want := []Job{
			{ID: 0, DocID: "foo", Status: "created"},
			{ID: 1, DocID: "bar", Status: "created"},
		}
		got := js.GetAllJobs()
		assert.ElementsMatch(t, want, got)
	})
}

func TestJobStore_GetJobsToProcess(t *testing.T) {
	t.Run("Test getting a single job from the store", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")
		_, _ = js.CreateJob("bar")

		assert.Equal(t, 0, js.nextIDToProcess)

		want := []Job{{ID: 0, DocID: "foo", Status: "created"}}
		got, _ := js.GetJobsToProcess(1)
		assert.Equal(t, want, got)
		assert.Equal(t, 1, js.nextIDToProcess)

		want2 := []Job{{ID: 1, DocID: "bar", Status: "created"}}
		got2, _ := js.GetJobsToProcess(1)
		assert.Equal(t, want2, got2)
		assert.Equal(t, 2, js.nextIDToProcess)
	})

	t.Run("Test getting multiple jobs from the store", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")
		_, _ = js.CreateJob("bar")
		_, _ = js.CreateJob("baz")
		_, _ = js.CreateJob("bas")
		_, _ = js.CreateJob("fred")

		want := []Job{{ID: 0, DocID: "foo", Status: "created"}}
		got, _ := js.GetJobsToProcess(1)
		assert.Equal(t, want, got)
		assert.Equal(t, 1, js.nextIDToProcess)

		want2 := []Job{{ID: 1, DocID: "bar", Status: "created"}}
		got2, _ := js.GetJobsToProcess(1)
		assert.Equal(t, want2, got2)
		assert.Equal(t, 2, js.nextIDToProcess)

		want3 := []Job{{ID: 2, DocID: "baz", Status: "created"}, {ID: 3, DocID: "bas", Status: "created"}}
		got3, _ := js.GetJobsToProcess(2)
		assert.Equal(t, want3, got3)
		assert.Equal(t, 4, js.nextIDToProcess)
	})

	t.Run("Test getting a job not in the store", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")

		want := []Job{{ID: 0, DocID: "foo", Status: "created"}}
		got, _ := js.GetJobsToProcess(1)
		assert.Equal(t, want, got)
		assert.Equal(t, 1, js.nextIDToProcess)

		wantErr := "No unprocessed jobs available"
		want2 := []Job{}
		got2, err := js.GetJobsToProcess(1)
		assert.Equal(t, wantErr, err.Error())
		assert.Equal(t, got2, want2)
		assert.Equal(t, 1, js.nextIDToProcess)
	})

	t.Run("Test getting a partially-filled slice of jobs", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")
		_, _ = js.CreateJob("bar")

		want := []Job{{ID: 0, DocID: "foo", Status: "created"}}
		got, _ := js.GetJobsToProcess(1)
		assert.Equal(t, want, got)
		assert.Equal(t, 1, js.nextIDToProcess)

		want2 := []Job{{ID: 1, DocID: "bar", Status: "created"}}
		got2, err := js.GetJobsToProcess(2)
		if err != nil {
			t.Errorf("Unexpected error %v", err.Error())
		}
		assert.Equal(t, want2, got2)
		assert.Equal(t, 2, js.nextIDToProcess)
	})
}

func TestJobStore_updateJobStatus(t *testing.T) {
	t.Run("Set to random string", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")

		want := Job{ID: 0, DocID: "foo", Status: "mystatus"}
		err := js.updateJobStatus(0, "mystatus")
		if err != nil {
			t.Errorf("JobStore.updateJobStatus() got an unexpected error: %v", err.Error())
			return
		}

		got, _ := js.GetJob(0)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("JobStore.updateJobStatus() = %v, want %v", got, want)
		}
	})

	t.Run("Errors as expected", func(t *testing.T) {
		js := NewJobStore()
		_, _ = js.CreateJob("foo")

		want := "job with id=1 not found"
		err := js.updateJobStatus(1, "mystatus")
		if err == nil {
			t.Error("JobStore.updateJobStatus() didn't error as expected")
			return
		}

		if err.Error() != want {
			t.Errorf("JobStore.updateJobStatus got error '%v', wanted error '%v'", err.Error(), want)
		}

	})

}
