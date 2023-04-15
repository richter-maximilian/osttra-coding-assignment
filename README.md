OSTTRA coding assignment
===============
This repository contains the code for the coding assignment of OSTTRA. 
The program is a web service for sending and recieving messages.

## How to run the webservice?
The docker image can be build via docker-compose
```
docker-compose up webservice
```
This will spin up the webservice and a postgres database.

Alternatively, if you have Go installed, you can start the webservice by running
```
go mod download
go run main.go
```

The environment variable `DB_CONN` needs to be the URL to a postgres instance.

## API
### POST /messages

This endpoint submits a message.

#### Request body example

```json
{
  "recipient_user_name": "user-name",
  "content": "content"
}
```

#### Reply example

```
200 OK
```

```json
{
  "message_id": "message-id",
}
```

### Get /messages/new

This endpoint fetches all new messages. The messages are ordered by time.

#### Reply example

```
200 OK
```

```json
[
  {
    "id": "message-id",
    "recipient_user_name": "user-name",
    "content": "content",
    "sent_at": "2023-04-13T19:43:23.999145+02:00"
  }
]
```

### DELETE /messages

This endpoint deletes messages by message id. It is possible to delete one or multiple message with one call.

#### Request body example

```json
{
  "message_ids": ["message-id"]
}
```

#### Reply example

```
204 No content
```

### Get /messages

This endpoint gets all new messages, the ones that have not been fetched and the ones that have been fetched. The messages are ordered by time.

#### Query parameters

- `start_cursor`
  - message ID
  - all messages sent later than the `start_cursor` are returned (including the `start_cursor`)
  - optional
- `end_cursor`
  - message ID
  - all messages sent before the `end_cursor` are returned (including the `end_cursor`)
  - optional

#### Reply example

```
200 OK
```

```json
[
  {
    "id": "message-id",
    "recipient_user_name": "user-name",
    "content": "content",
    "sent_at": "2023-04-13T19:43:23.999145+02:00",
    "fetched_at": "2023-04-14T19:43:23.999145+02:00"
  }
]
```