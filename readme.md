# The Golang Context

In Golang applications you will see a context getting passed around a lot.
This context carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.
This example project was created to better understand how the Golang context deadlines and cancellation signals work.
A deadline can occur if we specify a timeout or deadline values in the context.
A cancellation can occur for many reasons, but we are going to test by clicking cancel as the request is being processed.
When deadline or cancellation occurs the context send done signal that application can respond to.
We will be examining the lifetime of a request and how the various context configurations effect what occurs.

Make a request to http://localhost:8080/test to start this process. 
This is a flat project with all the functionality contained in the main.go file.
The request to test gets routed to the test function in main.go.
```
func test(response http.ResponseWriter, request *http.Request) ...
```
This request does two different tasks: load a person from the database and get a person from a server side rest call.
It will take at least ten seconds to process as both tasks have a five second pause in them.
The application logs what is occurring in the console for you to follow along.

The initial configuration we are going to examine is `ctx := request.Context()`.
This is using the requests context meaning if you were to cancel your request while this application is processing it the done signal will be triggered. 
In the initial stage the app is configured to any skip processing that hasn't occurred yet and improve performance.

Let's make our first request, http://localhost:8080/test, and just let it completely process.
You should see a json response of two people: Paul and Amy.

Make a second request see what happens when you click cancel while it is being processed.
You will now see a `The get context was canceled` message in the logs and notice all processing that had not yet occured was skipped.
todo: add more starting here

## Running the database
This application depends on a Postgres database. 
There is a docker compose file to create it for you.
Upon start up it will automatically create and populate a person table.

```
start up: docker-compose up -d
shut down: docker-compose down
```