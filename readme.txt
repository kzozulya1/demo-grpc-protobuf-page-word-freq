Small gRCP / Protobuf application 
that manage to create/update/delete/fetch data for website page url

Transport unit structure looks like
{
    "page_url": "https://www.nature.com/",
    "page_title": "Global Nature",
    "words":[
        {"value":"wild","count":213},
        {"value":"bear","count":10},
        {"value":"duck","count":32}
    ]
}

1. Run server 
$ docker-compose up golang-grpc-service

2. Run client
$ docker exec -ti golang-grpc-service go run cli/client.go
Client sends 2 protobuf messages to service for create sample data (file golang-grpc-service\src\SampleData.json)
Then fetches data with filter `tu`
At the end it removes all created data.
Otherwise next time client will be invoked, the data on server would be just updated.

Here is screenshot of client execution:
https://monosnap.com/file/WCwdEsHtwgUIZsTdFYJTXwrM8GlhvN