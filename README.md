# Service Example

An example of a golang service using Hexagonal Architecture with the core domain being located in the Service folder.

A brief overview follows:

## Application root

Contains the definitions for our domain models, one shown here being transaction.

Being at the root of the application does two things:

1) a developer can quickly glance and see this service is concerned with transactions
2) the dependency direction is one way, starting from this top level down to the services, stores and handlers - no circular dependencies

These domain files will contain structs defining the objects as well as the interfaces:

* Service interfaces to define our core domain layer
* Reader/Writer interfaces for interacting with a data store

## Transports

These are our methods of exposing data, these are our public interface. We could have http endpoints, grpc endpoints or message publishers here.

They should be dumb and have no business logic at all here, they simply parse requests and responses and pass onto the service layer.

## Service

Our core domain knowledge is here, this knows how to validate data, how to publish data to messaging system, what order to call data stores in etc.

They should have no knowledge of the transport layer on top, or the data layer below them.

## Data

Can have one or many data stores here, their only role being to store and retrieve data that has been passed to them by the service layer.

## Example.

There is a top to bottom example in the cmd/http-server binary.

This uses a noop data store and you can follow the requests up and down through the layers, just run

`go run cmd/http-server/main.go`