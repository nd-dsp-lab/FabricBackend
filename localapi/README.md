# Local API Documentation

## Overview

This project provides a local RESTful API service written in Go for certificate management, using CouchDB as the backend database. The service is containerized and managed via Docker Compose for easy local development and testing.

---

## Start the Service

Make sure you have [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed.

From the `localapi` directory, run:

```bash
docker compose up --build
```
This command will build the Docker images and start the containers defined in the `docker-compose.yml` file. 

The API will be available at `http://localhost:8080`, you can get a visulized view of API at `http://localhost:8080/docs#`

The CouchDB service will be at `http://localhost:5984`, you can access the CouchDB web interface at `http://localhost:5984/_utils/` (username: `admin`, password: `admin123`).


## API
The Certificate structure is defined in [./src/utils/certificate.go](localapi/src/utils/certificate.go). :


The API provides the following endpoints:
- `POST /certificates/create`: Create a new certificate.
- `POST /certificates/query`: Query for certificates with a Mongo selector:
```json
{
  "selector": {
    "pilot_id": "Pilot00"
  }
}
```
The selector is a JSON object that specifies the criteria for the query. The API will return all certificates that match the selector.
Currently the API only supports the following fields in the selector:
- `pilot_id`: The ID of the pilot.
- `drone_id`: The ID of the drone.
- `CertificateID`: The ID of the certificate, which is randomly generated.

