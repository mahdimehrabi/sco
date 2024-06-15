![Scale Ops](./scalesops-logo-v1.png)

# Setup 
` cp env.example .env` 
then set your database connection in the .env file
Make sure you have a valid postgresql connection in .env

then to run migrations `make db-migrate-up` 


I implement an ability to download images with proxy too but
please download images without proxy which is preferred because program will scrap free proxies and those free proxies connection speed and quality are awful

#### Without using docker
<br>``` go run ./cmd/main.go```<br>

#### using docker

```go mod vendor```

``` docker build -t image-downloader . ```

``` docker run -it --network=host image-downloader ```

## Unit Test
I only wrote unit test for service layer because lack of time but this will give you the idea of how I write unit tests

```cd ./domain/service```

run with

``` go test ./... ```

and benchmark with
``` go test -bench . -benchmem  ```

# Key Features

- **Used worker group pattern** to optimize the process of downloading and saving as file and in the db
- **The program automatically fetch proxies** from internet and use them in our software (but free proxies have awful speed and you must enable your vpn if you are in iran so I recommend to dont use this option)
- **Proxy fetching** won't work with iran ip so please make sure golang or docker using your system vpn
- **Used batch insertion** to increase database tps
- **Used rate limiter** to handle the rate of download to prevent google rate limit
- **Handling problems with image and network** and prevent any effect to performance by network and image encoding problem
- **Extend the program to use 2 search engines (google,bing)** and implement an interface for implementing any other search engine very easily, this way we can accumulate more images and we can have better performance and better handling search engines rate limits
- **It downloads exactly the number that user entered even with very large numbers like over 100k** without even one race condition problem
- **Read with a chunk pattern** to increase performance for reading images from database

##### Created with ❤ by mahdi mehrabi
thank you it was an interesting topic 
