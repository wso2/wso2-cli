package connect

var endpointsYamlSpec = `version: 0.1

# +required List of endpoints to create
endpoints:
  # +required Unique name for the endpoint. (This name will be used when generating the managed API)
  - name: rest-bridge
    # +required Numeric port value that gets exposed via this endpoint
    port: 8080
    # +required Type of the traffic this endpoint is accepting. Example: REST, GraphQL, etc.
    # Allowed values: REST, GraphQL, GRPC. WS
    type: REST
    # +optional Network level visibility of this endpoint. Defaults to Project
    # Accepted values: Project|Organization|Public.
    networkVisibility: Public
    # +optional Context (base path) of the API that is exposed via this endpoint.
    # This is mandatory if the endpoint type is set to REST or GraphQL.
    context: /
    schemaFilePath: openapi.yaml`

/*
var endpointsYamlSpec = `version: 0.1

    # +required List of endpoints to create
    endpoints:
      # +required Unique name for the endpoint. (This name will be used when generating the managed API)
      - name: ws-bridge
        # +required Numeric port value that gets exposed via this endpoint
        port: 8081
        # +required Type of the traffic this endpoint is accepting. Example: REST, GraphQL, etc.
        # Allowed values: REST, GraphQL, GRPC. WS
        type: WS
        # +optional Network level visibility of this endpoint. Defaults to Project
        # Accepted values: Project|Organization|Public.
        networkVisibility: Public
        # +optional Context (base path) of the API that is exposed via this endpoint.
        # This is mandatory if the endpoint type is set to REST or GraphQL.
        context: /ws
        schemaFilePath: asyncapi.yaml
      # +required Unique name for the endpoint. (This name will be used when generating the managed API)
      - name: rest-bridge
        # +required Numeric port value that gets exposed via this endpoint
        port: 8080
        # +required Type of the traffic this endpoint is accepting. Example: REST, GraphQL, etc.
        # Allowed values: REST, GraphQL, GRPC. WS
        type: REST
        # +optional Network level visibility of this endpoint. Defaults to Project
        # Accepted values: Project|Organization|Public.
        networkVisibility: Public
        # +optional Context (base path) of the API that is exposed via this endpoint.
        # This is mandatory if the endpoint type is set to REST or GraphQL.
        context: /
        schemaFilePath: openapi.yaml`
*/

var openApiSpec = `openapi: 3.0.0
info:
  title: REST API for Component Communication
  description: "REST APIs to Bridge Local Env"
  version: 1.0.0
paths:
  /health:
    get:
      summary: Health check
      responses:
        "200":
          description: Successfully running service

  /preview/{user}/{component}/{path}:
    parameters:
      - name: user
        in: path
        required: true
        schema:
          type: string
      - name: component
        in: path
        required: true
        schema:
          type: string
      - name: path
        in: path
        required: true
        schema:
          type: string
    get:
      summary: Forward a GET request to the client
      responses:
        "200":
          description: Request forwarded successfully
    post:
      summary: Forward a POST request to the client
      responses:
        "200":
          description: Request forwarded successfully
    put:
      summary: Forward a PUT request to the client
      responses:
        "200":
          description: Request forwarded successfully
    delete:
      summary: Forward a DELETE request to the client
      responses:
        "200":
          description: Request forwarded successfully

  /preview/{user}/{component}:
    parameters:
      - name: user
        in: path
        required: true
        schema:
          type: string
      - name: component
        in: path
        required: true
        schema:
          type: string
    get:
      summary: Forward a GET request to the client
      responses:
        "200":
          description: Request forwarded successfully
    post:
      summary: Forward a POST request to the client
      responses:
        "200":
          description: Request forwarded successfully
    put:
      summary: Forward a PUT request to the client
      responses:
        "200":
          description: Request forwarded successfully
    delete:
      summary: Forward a DELETE request to the client
      responses:
        "200":
          description: Request forwarded successfully

  /handle-client-req:
    post:
      summary: Accept request details from client and make the request within the dataplane
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
        "200":
          description: Request handled successfully`

var asyncApiSpec = `asyncapi: "2.6.0"
info:
  title: WebSocket API for Component Communication
  version: "1.0.0"
  description: |
    This API allows clients to connect to the server via WebSocket and handle bidirectional communication.
    Clients can receive requests and send responses back to the server.

servers:
  production:
    url: ws://localhost:8080
    protocol: ws
    description: Production WebSocket server

channels:
  /{user}/{component}:
    parameters:
      user:
        description: ID of the user connecting to the WebSocket.
        schema:
          type: string
      component:
        description: The handle of the component connecting to the WebSocket.
        schema:
          type: string
    subscribe:
      operationId: handleWebSocket
      summary: Subscribe to WebSocket messages
      description: |
        Clients can connect to this WebSocket endpoint to receive requests and send responses.
      message:
        oneOf:
          - $ref: "#/components/messages/RequestMessage"
          - $ref: "#/components/messages/ResponseMessage"

components:
  messages:
    RequestMessage:
      name: RequestMessage
      summary: A request sent from the server to the client.
      payload:
        type: object
        properties:
          request_id:
            type: string
            description: Unique ID for the request.
          method:
            type: string
            description: HTTP method (e.g., GET, POST).
          path:
            type: string
            description: Request path (e.g., /foo/bar).
          headers:
            type: object
            description: Request headers.
            additionalProperties:
              type: string
          query:
            type: object
            description: Query parameters.
            additionalProperties:
              type: string
          body:
            type: string
            description: Request body.
          port:
            type: string
            description: Optional port (e.g., "8080").

    ResponseMessage:
      name: ResponseMessage
      summary: A response sent from the client to the server.
      payload:
        type: object
        properties:
          request_id:
            type: string
            description: Unique ID for the request.
          status_code:
            type: integer
            description: HTTP status code (e.g., 200, 404).
          headers:
            type: object
            description: Response headers.
            additionalProperties:
              type: string
          body:
            type: string
            description: Response body.`
