# SendX-Backend-Assignment

Given Assessment :
 - Write a server endpoint which takes the URL of a webpage. 
 - After getting the URL it fetches the webpage and downloads it as a file on the local file system.
 - The server accepts a retry limit as a parameter. It retries maximum upto 10 times or retry limit, whichever is lower, before either successfully downloading the webpage or marking the page as a failure.
 - If the webpage has already been requested in the last 24 hours then it should be served from the local cache 
 - Make a pool of workers that do the work of downloading the requested webpage. So basically if there are 10000’s of concurrent requests to download web pages, the number of requests to actually download them are still limited (based on number of active workers)


JavaScript : 
Implemented a server endpoint using Node.js and Express that takes a URL and a retry limit as parameters, fetches the webpage, and downloads it as a .html file on the local file system. 
 - The server endpoint creates a pool of workers that download the webpage in parallel. 
 - If the number of active workers exceeds the maximum, the oldest worker is removed from the pool to make room for the new one. 
 - If the webpage has already been requested in the last 24 hours, it is served from the cache. 
 - If the webpage has not been requested recently, it is downloaded and added to the cache. 
 - If the download fails, the worker retries up to the minimum of retry limit given or 10 times before giving up and returning an error.


Golang :
Implementd a server endpoint in Go using concurrency techniques that takes a URL and a retry limit as parameters, fetches the webpage, and downloads a file on the local file system:

 - main: It creates a cache folder to store downloaded files if it doesn't exist and a the worker pool, then starts the worker pipeline by calling workerPipeline. It also sets up an HTTP handler for the "/download" using the downloadHandler function. Finally, it starts the HTTP server.

 - downloadHandler: This function is HTTP handler that is called when a request is made to the /download. It parses the query string to get the URL and retry limit, and checks if the webpage has been requested in the last 24 hours. If it has, it serves the webpage from the cache. Otherwise, it sends the Worker to the work queue.

 - workerPipeline: This function listens to the work queue and sends incoming workers to the worker pool. It tries to find an available worker channel in the pool and sends the worker to that channel. If it does not find an available worker channel, it creates a new one and starts a new worker.

 - workerFunc: This function runs in a loop, waiting for the next worker to be sent to the worker channel. When it receives a worker, it downloads the webpage, adds it to the cache, and saves it to the local file system. If any of these steps fail, it retries up to the specified retry limit i.e., minimum of given retry limit and 10, before returning an error.

 - downloadWebpage: This function takes a URL and a retry limit as parameters and downloads the webpage at that URL. If the download fails, it retries up to the specified retry limit before giving up and returning an error.

 - saveWebpage: This function takes an HTML string and a URL as parameters and saves the HTML to a file with the name of the URL parsed according to file naming conventions and a .html extension in cache folder. If the file cannot be created or written to, it returns an error.

This implementation allows us to limit the number of concurrent requests to download webpages, even if there are many requests coming in at the same time. The number of requests that are actually sent to the server is limited by the MaxWorkers variable.
