
# PURCHASES-MULTI-COUNTRY

###### To generate the test coverage, run the command below in your terminal, in the root folder of the project. You can also use my custom ./skaffold.sh which have many common commands to run for this project.
- `$ go test -v -coverpkg=./... -coverprofile=profile.cov ./...; go tool cover -func profile.cov; rm profile.cov`

Right now I have **59.9%** of test coverage since I had only 5 days to do this project. If we follow the test patterns we can go to 100% in a few days.

## Pré Requirements

- **Golang 1.19+**
- **Docker & Docker-Compose**: To run the full project in containers (with mysql as well)
- **Linux**: I strongly recommend running the project in a **linux** environment. If you are going to use Windows, you will need to deal with some issues involving **SWL2** and port mirroring. If you are using **MacOS** it will also work, but you will also need to deal with port mirroring if you want to run the complete environment (docker-compose). Windows like MacOS, uses a virtualization structure from a native Linux environment (filesystem) to instantiate docker images and therefore problems such as _"I want to directly access a port of the running container"_ become more complex to resolve in these environments. With a good understanding, it is possible to run smoothly on both, but the support of this documentation is **_exclusively on Linux_**.
- **curl**: To run manual requests, but feel free to use **wget** if you prefer.

## Topology

I used a custom framework created by myself. This framework uses multiple patters following the SOLID principles.

Following the **SOLID** principles, the code follows a layered structure using **Inversion of Control**(Inspired by the old Spring's ApplicationContext...) and **"NoOps"**(**_No Operation_**, standard _"imported"_ from civil and mechanical engineering) implemented. We will look at these two patterns in more depth throughout this documentation but in principle, using these patterns, any "layer" (service/component) can be accessed from anywhere in the code without any problem. This topology gives this code almost infinite flexibility. In the "infinite" sense regarding _performance_, all external calls (**https://swapi.dev**) are made in a **_concurrent/parallel_** and in a **_non-blocking_** manner, drastically reducing the latency of requests experienced by the end user during Http requests. In the "infinite" sense regarding the _speed_ of adding new features, with the ServiceManager managing the life cycle of any component, we simply add the new feature as a new component, implement its NoOps, add the ServiceManager and then it can be used by any other component in a free and mainly decoupled way.

### Inversion of Control

The _Inversion of Control_ pattern consists of allowing another entity to be in charge of managing the life cycle of **ALL** dependencies (objects/instances) of a specific component. For example, the _ExchangeService_ component needs the _TreasuryAccessService_ component and the _PersistenceService_ component to perform its functions. In an application that does not use _Inversion of Control_, the _ExchangeService_ would be responsible for **instanciating** the dependencies and thus it ends up being responsible for controlling the entire life cycle of these components.

In our case, the **ServiceManager** is the entity _responsible for controlling the life cycle of ALL other components of the application and this is **the only responsibility of this entity (SOLID)**_. This way, if the _ExchangeService_ needs to use the _TreasuryAccessService_ to make the request to the public API, with a single line of code the _ExchangeService_ receives an instance of the _TreasuryAccessService_ from the _ServiceManager_. The same behavior exists in the _"communication"_ between ALL the _services/components_ of the system, **isolating and decoupling them**.

The ServiceManager uses a 'fluent API' to enable easy use of all lifecycle functions, as exemplified in the application's cmd/main/main.go.

### NoOps (No Operation)

_No Operation_ is a little-known name in the software industry, however, it is widely used. Inspired by civil construction, a famous example of the pattern is the existence of _"balancing steel balls"_ used in the construction of very large buildings in places where there is a lot of wind. With wind pressure, all very tall buildings naturally bend and unbuck. In the center of these buildings there is ALWAYS a large steel ball attached by a steel rope to the ceiling and hanging at a certain height (normally half the building) suspended in the air. This ball swings as the building _"tilts"_, playing the role of adjusting the building's center of balance.

The No Operation concept of the "steel ball" case comes from the fact that this ball requires **_NO_** maintenance. Any engineering modification made to the building (apart from changing its height) does not result in changes in the position of the ball. The ball will just exist there, suspended in the center of the building, doing its job. Requiring no maintenance, this is what we call the NoOps (No Operation) pattern.

Bringing the example to our software world, the NoOps standard consists of code structures that are necessary for the system to function as it was designed but without requiring any maintenance :). For example, the struct **noOpsExchangeService** is exactly the **_NoOps_** implementation of the ExchangeService interface. This struct implements the interface and provides basic behaviors used in **ALL** unit tests in this project. **Basically, the _NoOps_ instances play the entire role of returning "mocks"** used in unit tests so that much less code is needed to write a unit test that requires Mocks from the **ExchangeService** interface for example, which is exactly the case with unit tests for the **HttpService** entity. In turn, the **HttpService** interface also has its **NoOps** implementation, which allows any other component that needs HttpService mocks to use this other implementation without bureaucracy.

### Logs

I used Uber's Zap lib (https://go.uber.org/zap) to create structured logs containing the following information:
```
{
    "level":"info", # log level (info, warn, error, debug)
    "ts":1669308163.428072, #timestamp
    "caller":"exchangeservice/exchangeservice.go:29", #"package/filename:line"
    "msg":"Exchange Service Started!", #log message text
    "AppEnv":"PROD", #environment
    "service":"swapiapp", #project name
    "version":"1.0" # version
}
```

### Running this project

You can run the code with a simple `>$ go mod tidy; go run cmd/main/main.go` however, without an instance of mysql up and running, listening to the host **_db:3306_** you will receive errors. For this reason, one of the prerequisites is the use of Docker and Docker Compose to run the project.

To make everyone's life easier, I created a shell script called `scaffold.sh` which has the following commands:
```
swapi/$ ./scaffold.sh full-rebuild -prune -runtests #compile, test and run the project with docker volumes clean.
swapi/$ ./scaffold.sh runtests #run all tests and shows the test coverage.
swapi/$ ./scaffold.sh build #just compile the code and the docker image.
swapi/$ ./scaffold.sh down #destroy all docker running images.
swapi/$ ./scaffold.sh up #run all already compiled containers from the docker-compose.yml.
swapi/$ ./scaffold.sh logs -app #appends shell session to the app's log tail.
swapi/$ ./scaffold.sh logs -db #appends shell session to the mysql container log tail.
```
ps.: The -prune and -runtests commands are optional.

With these commands:

- execute `./scaffold.sh full-rebuild -prune -runtests` for the first build, test and run.
- Wait for the last command to finish.
- Make sure the database container is running and finished creating the datafiles and is listening to the 3306 port: **[Server] /usr/sbin/mysqld: ready for connections. Version: '8.0.31'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server - GPL.** using `./scaffold.sh logs -db`.
- Press **CTRL + C** to go back to the terminal.
- Execute `./scaffold.sh restartapp` or `docker compose restart purchases-multi-country` to restart the app's container.
- Follow the steps below to access the endpoints.

### POST -H 'Countrycurrency: Brazil-Real' /purchases

Insert a new purchase and follow the **Idempotency** pattern in the way if you insert the same purchase later, the endpoint will just answer 200 and no change will be made to the database. This endpoint will verify if that transaction already exists and if it does not exists it will persist it. After the persistence, a goroutine will be triggered to async load ALL the exchange rates from the Treasury Access API(external service). With this flow, the user will get a quick response and the load of the exchages will happen in the "background".

Ex:
```
curl -X POST -H 'Content-Type: application/json' -d "{\"id\": \"$(echo $RANDOM | md5sum | head -c 10)\", \"description\": \"Some purchase\", \"amount\": \"20.13\", \"date\": \"2023-10-29\"}" http://localhost:8080/purchases
```

### GET /purchases/:id

return a specific purchase from the **:id**(string) informed, calculated using the informed "Countrycurrency" header. The header is a requirement.
Ex:
```
curl -X GET -H 'Content-Type: application/json' -H "Countrycurrency: Brazil-Real" http://localhost:8080/purchases/$SOME_ID
```

### GET /purchases

Return every purchase from the database wiht the amount converted based on the "Countrycurrency" header. The header is a requirement.

Ex:
```
curl -X GET -H 'Content-Type: application/json' -H "Countrycurrency: Brazil-Real" http://localhost:8080/purchases
```

### Shuttinh down

just call  `$ docker compose down`, `docker system prune -f` and `docker volume prune -f`.