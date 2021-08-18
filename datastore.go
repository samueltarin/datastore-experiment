package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/appengine"
)

// Global Client connection
var datastoreClient *datastore.Client

// Models
type UserPracticeTest struct {
	AdminYear int `datastore:"admin_year,noindex,omitempty"`
	// NOTE: Simplified for test

	// NOTE: this is indexed
	Kaid string `datastore:"kaid"`
}

type UserPracticeTestRepeated struct {
	UserPracticeTests []*UserPracticeTest `datastore:"user_practice_tests,noindex,omitempty"`
	// NOTE: Simplified for test
}

// Constructor
func MakeUserPracticeTestRepeated(kaid string, num_tests int) *UserPracticeTestRepeated {
	upts := make([]*UserPracticeTest, num_tests)

	for i := 0; i < num_tests; i++ {
		upts[i] = &UserPracticeTest{
			AdminYear: 2000 + i,
			Kaid: kaid,
		}
	}

	return &UserPracticeTestRepeated{
		UserPracticeTests: upts,
	}
}

func main() {
	ctx := context.Background()

	// Set this in app.yaml when running in production.
	projectID := os.Getenv("GCLOUD_DATASET_ID")

	var err error
	datastoreClient, err = datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer datastoreClient.Close()

	http.HandleFunc("/", handle)
	appengine.Main()
}

func maybeHandleDatastoreError(w *http.ResponseWriter, err error, operation, kind string) {
	if err != nil {
		msg := fmt.Sprintf("Could not %s %s to datastore: %v", operation, kind, err)
		http.Error(*w, msg, http.StatusInternalServerError)
	}
	// fmt.Fprintln(*w, operation, kind, "successful")
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := context.Background()

	DO_WRITE := true
	// Disable so nothing happens
	// NUM_USERS := 0
	NUM_USERS := 10
	NUM_ENTITIES_PER_USER := 10

	seededRand := rand.New(rand.NewSource(time.Now().Unix()))
	randomOrder := seededRand.Perm(NUM_USERS)

	// Repeated Structured Properties
	kind := "UserPracticeTestRepeated"

	// Write Path
	keys := make([]*datastore.Key, NUM_USERS)
	for i := 0; i < NUM_USERS; i++ {
		key := datastore.NameKey(kind, fmt.Sprintf("TestRepeatedKey%04d", i), nil)
		keys[i] = key
		if DO_WRITE {
			upts := MakeUserPracticeTestRepeated(fmt.Sprintf("kaid_%04d", i), NUM_ENTITIES_PER_USER)
			_, err := datastoreClient.Put(ctx, key, upts)
			maybeHandleDatastoreError(&w, err, "Put", kind)
		}
	}

	// Read Path
	start := time.Now()
	for _, i := range randomOrder {
		uptrAgain := &UserPracticeTestRepeated{}
		err := datastoreClient.Get(ctx, keys[i], uptrAgain)
		maybeHandleDatastoreError(&w, err, "Get", kind)
		if len(uptrAgain.UserPracticeTests) != NUM_ENTITIES_PER_USER {
			fmt.Fprintln(w, "Error reading all entities for key", keys[i])
		}
	}
	end := time.Now()

	fmt.Fprintln(w, "Repeated Structured Properties Read Time:", end.Sub(start).Seconds()/float64(NUM_USERS)*1000)

	// Entity Groups
	kind = "UserPracticeTestGroup"

	// Write Path
	parentKeys := make([]*datastore.Key, NUM_USERS)
	for i := 0; i < NUM_USERS; i++ {
		parentKey := datastore.NameKey("UserPracticeTestParent", fmt.Sprintf("ParentKey%04d", i), nil)
		parentKeys[i] = parentKey
		if DO_WRITE {
			for j := 0; j < NUM_ENTITIES_PER_USER; j++ {
				// Put entities in a group
				key := datastore.NameKey(kind, fmt.Sprintf("TestKey%04d", j), parentKey)
				upt := &UserPracticeTest{
					AdminYear: 2000 + 10*i + j,
					Kaid: fmt.Sprintf("kaid_%04d", i),
				}
				_, err := datastoreClient.Put(ctx, key, upt)
				maybeHandleDatastoreError(&w, err, "Put", kind)
			}
		}
	}

	// Read Path
	start = time.Now()
	for _, i := range randomOrder {
		uptQueryResults := make([]*UserPracticeTest, 0)
		// Query for all entities for the user
		q := datastore.NewQuery(kind).Ancestor(parentKeys[i])
		_, err := datastoreClient.GetAll(ctx, q, &uptQueryResults)
		maybeHandleDatastoreError(&w, err, "Query", kind)
		if len(uptQueryResults) != NUM_ENTITIES_PER_USER {
			fmt.Fprintln(w, "Error reading all entities for key", parentKeys[i])
		}
	}
	end = time.Now()

	fmt.Fprintln(w, "Entity Group Read Time:", end.Sub(start).Seconds()/float64(NUM_USERS)*1000)

	// Entity (Indexed & Group)
	kind = "UserPracticeTestIndexedGroup"

	// Write Path
	kaids := make([]string, NUM_USERS)
	parentKeys = make([]*datastore.Key, NUM_USERS)
	for i := 0; i < NUM_USERS; i++ {
		parentKeys[i] = datastore.NameKey(
			"UserPracticeTestParentIndexed",
			fmt.Sprintf("ParentKey%0d", i),
			nil,
		)
		kaids[i] = fmt.Sprintf("kaid_%04d", i)
		if DO_WRITE {
			for j := 0; j < NUM_ENTITIES_PER_USER; j++ {
				// Put entities (all with same parent & kaid)
				key := datastore.NameKey(kind, fmt.Sprintf("TestKey%0d", i*10 + j), parentKeys[i])
				upt := &UserPracticeTest{
					AdminYear: 2000 + 10*i + j,
					Kaid: kaids[i],
				}
				_, err := datastoreClient.Put(ctx, key, upt)
				maybeHandleDatastoreError(&w, err, "Put", kind)
			}
		}
	}

	// Read Path
	start = time.Now()
	for _, i := range randomOrder {
		uptQueryResults := make([]*UserPracticeTest, 0)
		// Query for all entities for the user (by parent & kaid)
		q := datastore.NewQuery(kind).Ancestor(parentKeys[i]).Filter("kaid =", kaids[i])
		_, err := datastoreClient.GetAll(ctx, q, &uptQueryResults)
		maybeHandleDatastoreError(&w, err, "Query", kind)
		if len(uptQueryResults) != NUM_ENTITIES_PER_USER {
			fmt.Fprintln(w, "Error reading all entities for key", parentKeys[i])
		}
	}
	end = time.Now()

	fmt.Fprintln(w, "Entity (Indexed Group) Read Time:", end.Sub(start).Seconds()/float64(NUM_USERS)*1000)

	// Entity (Indexed)
	kind = "UserPracticeTestIndexed"

	// Write Path
	kaids = make([]string, NUM_USERS)
	for i := 0; i < NUM_USERS; i++ {
		kaids[i] = fmt.Sprintf("kaid_%04d", i)
		if DO_WRITE {
			for j := 0; j < NUM_ENTITIES_PER_USER; j++ {
				// Put entities (all with same kaid)
				key := datastore.NameKey(kind, fmt.Sprintf("TestKey%0d", i*10 + j), nil)
				upt := &UserPracticeTest{
					AdminYear: 2000 + 10*i + j,
					Kaid: kaids[i],
				}
				_, err := datastoreClient.Put(ctx, key, upt)
				maybeHandleDatastoreError(&w, err, "Put", kind)
			}
		}
	}

	// Read Path
	start = time.Now()
	for _, i := range randomOrder {
		uptQueryResults := make([]*UserPracticeTest, 0)
		// Query for all entities for the user (by kaid)
		q := datastore.NewQuery(kind).Filter("kaid =", kaids[i])
		_, err := datastoreClient.GetAll(ctx, q, &uptQueryResults)
		maybeHandleDatastoreError(&w, err, "Query", kind)
		if len(uptQueryResults) != NUM_ENTITIES_PER_USER {
			fmt.Fprintln(w, "Error reading all entities for kaid", kaids[i])
		}
	}
	end = time.Now()

	fmt.Fprintln(w, "Entity (Indexed) Read Time:", end.Sub(start).Seconds()/float64(NUM_USERS)*1000)

}
