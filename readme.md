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
The request to test gets routed to the [test](./main.go#L44) function in main.go.
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
You will now see a `The get context was canceled` message in the logs and notice all processing that had not yet occurred was skipped.
The application just returns. It doesn't even need to return a http error code nor any JSON response.

Inside the test method you will see a commented out block of [code](./main.go#L50) showing all the possible context configuration option. 
The code is well commented. 
Reading through it and trying out the options should further help understanding how the context can function.
```
// This is using the requests context meaning if you were to cancel your request while this application is
// processing it the done signal will be triggered. Depending on how the app is configured it maybe able to skip
// processing that hasn't occurred yet and improve performance.
ctx := request.Context()

/* thy this: Try these other options instead:
// this can never be cancelled or exceed it runtime amount
ctx := context.Background()

// this returns a context and a function we called cancel
ctx, cancel := context.WithCancel(request.Context())
// calling cancel will trigger the done signal
cancel()

// this will trigger a timeout after 2 seconds
ctx, cancel := context.WithTimeout(request.Context(), time.Second * 2)
// from the Golang doc: Even though ctx will be expired, it is good practice to call its cancellation function in any case. Failure to do so may keep the context and its parent alive longer than necessary.
defer cancel()

// this will trigger a timeout after a time that is 2 seconds in the future
ctx, cancel := context.WithDeadline(request.Context(), time.Now().Add(time.Second * 2))
defer cancel()
*/
```

Here are few things to remember if you want the context to cancel or timeout. 
First be sure to pass the context along as [sometimes](./main.go#L216) it is optional. 
When errors occur [check](./main.go#L86) to see if the context is done and cease processing.
Finally, when creating your own potentially long running processing [logic](./main.go#L250) be sure to check for context done signals and return the error.

The last thing to show is how you can use the context to store request-scoped values. 
Since the context gets passed around all the time it provides a way to share these values.
I have previously used this for logging common values, like a request id. 
This has been [set up](./main.go#L70) and [used](./main.go#L265) in this example as well.

## Running the database
This application depends on a Postgres database. 
There is a docker compose [file](./docker-compose.yml) to create it for you.
Upon start up it will [automatically](./db/init.sql) create and populate a person table.

```
start up: docker-compose up -d
shut down: docker-compose down
```