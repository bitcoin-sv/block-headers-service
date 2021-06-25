# Transports

When we expose data outside of our application in rest services, as a message producer or grpc server for example, the handlers will be added here.

They will take a reference to a service interface and their only concern should be with parsing the request and responses to and from the service models. No business logic should be here.