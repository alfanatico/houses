# API client challenge

The project was built with Go programming language  [https://go.dev/](https://go.dev/).

Compiled with version 1.21rc1.

This is my first experience with Go,  but I selected it over spring boot for two reasons, the first one is that I like learning new frameworks and tools, and it was fun and a bit addictive.
The other reason is that if I am accepted in the team I will be far more comfortable during my first weeks as I am already quite familiar with the language and their tools (GoLand IDE).

## Design considerations

Since we are dealing with a flaky API, the code was developed using a custom retryable logic, but that layer could be easily removed in the future as it is encapsulated in a function that could be bypassed.
Another alternative could be to use a third party standard library like:

https://pkg.go.dev/github.com/hashicorp/go-retryablehttp

https://pkg.go.dev/github.com/sethvargo/go-retry

https://github.com/avast/retry-go

This project uses a channel and workers to download images concurrently, but this  feature is not tightly coupled into the main logic, and it could be easily removed by changing the constant enableConcurrentOptimization value. This could be helpful for troubleshooting or debugging, or just comparing the performance.

Testability, several methods were extracted to allow for better testability, and API struct was created to allow mocking an http server to test error handling and last page detection logic.

There are two main processes, the first step is to fetch house data, and the second step is to download images, however we start downloading earlier, as soon as the first page of data is obtained, later the fetch of new pages and downloads runs in parallel to speed up the process.
## Future improvements

Move constants to configuration file or system properties that can be passed when running from command line.
Move logic to packages, following the single responsibility principle.

Increase code coverage.

Implement retryable logic for download endpoint as well (not done for time constraints and because it was reliable)

## Building the application

Open a shell terminal at the project root.
```shell
go build
```
This will create an executable file named "hello.sh"

## Running the application
From the project root
```shell
go run main.go
```

Alternatively, you can run the executable created after the build.
From the project root
```shell
./hello
```

In the shell console you will see some messages showing the progress:

```shell
process starts
Processing page  1
2023/06/20 15:39:44 Worker 1 started  download ID= 1
2023/06/20 15:39:44 Worker 3 started  download ID= 2
2023/06/20 15:39:44 Worker 2 started  download ID= 0
2023/06/20 15:39:44 Worker 5 started  download ID= 4
2023/06/20 15:39:44 Worker 4 started  download ID= 3
2023/06/20 15:39:44 Worker 4 finished download ID= 3
2023/06/20 15:39:44 Worker 4 started  download ID= 5
2023/06/20 15:39:44 Worker 1 finished download ID= 1
2023/06/20 15:39:44 Worker 1 started  download ID= 6
process step 1 completed: fetch completed
2023/06/20 15:39:47 Worker 3 finished download ID= 99
2023/06/20 15:39:47 Worker 1 finished download ID= 90
2023/06/20 15:39:47 Worker 4 finished download ID= 98
process step 2 completed: images downloaded
```


After running the application the images will be downloaded to the output folder relative to the project root folder.



## Running tests

From the project root
```shell
go test -v
```

## About the author
Santiago Font is an experienced Java programmer who is doing his first steps in the Go world.

Feel free to reach me at 
[santiagofont@yahoo.es](mailto:santiagofont@yahoo.es)

