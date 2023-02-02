package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

// Person a simple struct representing a person
type Person struct {
	Name string
}

// main Sets up our application server and gets it running.
func main() {
	fmt.Println("Starting application")

	// creates a new instance of a mux router
	myRouter := mux.NewRouter()

	// add our routes
	myRouter.HandleFunc("/test", test)
	myRouter.HandleFunc("/server-side-get", serverSideGet)

	// start the server running at http://localhost:8080
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

// test The endpoint, http://locallhost:8080/get, to call to test out the context functionality.
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get was called")

	ctx := r.Context()
	/* todo: try these options instead:
	// this can never be cancelled or exceed it runtime amount
	ctx := context.Background()

	// this returns a context and a function we called cancel
	ctx, cancel := context.WithCancel(r.Context())
	// calling cancel to trigger the done signal
	cancel()

	// this will trigger a timeout after 2 seconds
	ctx, cancel := context.WithTimeout(r.Context(), time.Second * 2)
	// from the Golang doc: Even though ctx will be expired, it is good practice to call its cancellation function in any case. Failure to do so may keep the context and its parent alive longer than necessary.
	defer cancel()

	// this will trigger a timeout after a time time that is 2 seconds in the future
	ctx, cancel := context.WithDeadline(r.Context(), time.Now().Add(time.Second * 2))
	defer cancel()
	*/

	// create a slice/array to hold the person list
	var people []Person

	// lookup a person from the database
	person, err := databaseCall(ctx)
	if err != nil {
		// check if the context has been cancelled or has exceeded it runtime amount and sent the done signal
		if isDone(ctx) {
			// just return since we have no further work to do
			return
		}

		// an error occurred: log it and return a 500
		fmt.Println("Error retrieving database person:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// append this person from the database to the slice of people results
	people = append(people, person)

	// lookup a person by a server side rest call
	person, err = restCall(ctx)
	if err != nil {
		// check if the context has been cancelled or has exceeded it runtime amount and sent the done signal
		if isDone(ctx) {
			// just return since we have no further work to do
			return
		}

		// an error occurred: log it and return a 500
		fmt.Println("Error retrieving rest person:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// append this person from the rest call to the slice of people results
	people = append(people, person)

	// respond with the slice of people rendered as json
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(people)
	if err != nil {
		// an error occurred: log it and return a 500
		fmt.Println("Error building the people response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Get has finished")
}

// isDone A utility function that checks to see if a context has been cancelled or has exceeded it runtime amount and
// sent the done signal.
func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		// if the context done then log the reason
		err := ctx.Err()
		if errors.Is(err, context.Canceled) {
			fmt.Println("The get context was canceled")
		} else if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("The get context has timed out")
		} else {
			fmt.Println("The get context had an unexpected error")
		}
		return true
	default:
		// the context is not done so return false
		return false
	}
}

// serverSideGet The endpoint used to simulate a making a server rest call.
func serverSideGet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Server side get was called")
	ctx := r.Context()

	// pause for a bit to allow the context to be cancelled
	err := pause(ctx)
	if err != nil {
		isDone(ctx)
		return
	}

	// return the person named paul as json
	person := Person{Name: "Paul"}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(person)
	if err != nil {
		// an error occurred: log it and return a 500
		fmt.Println("Error building the server side get response:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Println("Server side get has finished")
}

// databaseCall Looks up a person from the database.
func databaseCall(ctx context.Context) (Person, error) {
	fmt.Println("Making the database call")
	var person Person

	// pause for a bit to allow the context to be cancelled
	err := pause(ctx)
	if err != nil {
		return person, err
	}

	// the popular pgx postgres database package requires a context to be set in most operations
	connection, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/postgres")
	if err != nil {
		// in addition to the usual errors if the pgx package notices the context is done it will return an error
		return person, err
	}
	defer connection.Close(ctx)

	// query the database for a person and populate their struct values
	err = connection.QueryRow(ctx, "select name from people").Scan(&person.Name)
	return person, err
}

// restCall Looks up a person by making a rest call.
func restCall(ctx context.Context) (Person, error) {
	fmt.Println("Making the rest call")
	var person Person

	// create the get request to the server side endpoint
	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/server-side-get", nil)
	/*
		todo: If we don't pass the context along the request will not be cancelled when a done signal occurs. The
		request will be fully processed wasting resources.
		request, err := http.NewRequest("GET", "http://localhost:8080/server-side-get", nil)
	*/

	// make the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		//todo
		return person, err
	}

	// read the full response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return person, err
	}

	// close the response body
	err = response.Body.Close()
	if err != nil {
		return person, err
	}

	// unmarshal the response body contents to a person struct
	err = json.Unmarshal(body, &person)

	return person, err
}

// pause Wait for five seconds unless the context is done.
func pause(ctx context.Context) error {
	// select and return whichever case occurs first
	select {
	case <-ctx.Done():
		// the context is done so return the specific error with the reason
		return ctx.Err()
	case <-time.After(5 * time.Second):
		// five seconds have elapsed so return with no error
		return nil
	}
}
