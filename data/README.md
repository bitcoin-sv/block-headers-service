# Data

This contains all I/O operations where data should be read or written to an external service such as a database, api or grpc endpoint for example.

Each data type will be split into its own package as shown here, we have:

* grpc
* postgres

## GRPC

This would wrap a grpc client used to call out to another service to send or receive data.

This would exist behind a store interface defined in the service root, meaning we don't explicitly tie the service into the proto definitions etc.

## Postgres

This would be a data store with functions in it to read and write from this data store. Again, as with GRPC, this would simply implement store functions defined at the service route.