# FabricBackend
A hyperledger fabric based blockchain backend storage.

## Dependencies

+ Golang 1.22.5+
  + In linux, please consult [go version manager](https://github.com/moovweb/gvm) to manage go version. 
+ jq
+ Docker
  + docker-compose (for older version docker)
  + In linux, add current user to the docker group to avoid privilege issue: ```sudo usermod -a -G docker $USER```


### On Ubuntu:
```
sudo apt-get install git curl jq -y
```
Start docker with 
```
sudo systemctl start docker
```

### On MacOS
```
brew install git curl jq go
```
Start docker with [Docker Desktop](https://www.docker.com/).


## To make the server work
1. Clone this repo to the destination of your choice, make sure you have make and docker/docker compose installed.
> The following commands are assumed to initiated from the FabricBackend/ directory.
2. Run `make check-prerequisite` to check prerequisites. 
3. Run `make install` to install the necessary components, including fabric, docker images, and chaincode.
4. Use `make drp_couchdb_deploy` to bring up the test network and deploy the contract.
> This may fail at contract installation in rare occasions, simply trying the command again should resolve the problem.
5. Use `make api_server` to bring up the api service

## To bring it down
1. Use `make down` to turn the network down
2. Use `make clean` to clean up the downloaded repos and docker images.

By default all the data will be cleaned when the network is down.

## APIs

The default API server address is http://localhost:8001

> `GET` /records/all

Return all records in standard JSON

> `POST` /records/query

Query records using a CouchDB Mango selector

> `POST` /records/update

Update a record

> `POST` /records/history

Get the history of a record

> `POST` /create-record (Authenticated)

Create a record with the JSON. The JSON needs to be formatted as the sample data.

<!-- 
In construction
> `GET` api/record/{selectorString} 

Return records with selector specified by {selectorString} -->

## Ports and Service
|Name|Type|Ports|
|----|----|-----|
|API server|Go|8001|
|orderer|Docker|7050|
|peer0.org1|Docker|7051|
|peer0.org2|Docker|9051|
|dev-peer.org1|Docker|-
|dev-peer.org2|Docker|-
|CouchDB0|Docker|5984|
|CouchDB1|Docker|7984|

For live demo, please check the [instructions](README_TEST.md).