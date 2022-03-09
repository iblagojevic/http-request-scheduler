# Golang HTTP Request Scheduler
\
Service provides http interface for scheduling RESTful requests to targeted endpoints. \
\
It can be treated as a blackbox which listens on port 9292 and accepts input in predefined format 

```
POST http://localhost:9292/
Accept: application/json
{
    "action": string,      // "POST" or "GET" are supported for now
    "url": string          // full url of targeted endpoint
    "delay": integer
    "payload": {}          // JSON object
}
```
and makes `POST` or `GET` request to `url` after `delay` seconds, with body `payload` in case of POST request.\
\
Service can be used as a testing tool or as internal service for scheduling delayed messages via REST APIs.\

Current version is basic, but the plan is to add handling for other protocols 
targeting message destinations (AMQP, etc) and support for `PUT` and `DELETE` actions on input.


