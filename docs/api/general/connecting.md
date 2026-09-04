# Connecting to the Tiingo API Documentation

> 출처: https://www.tiingo.com/documentation/general/connecting · 생성: 2026-09-04 (tools/gendocs)

## 1.2 Connecting

## 1.2.1 Connecting to the REST API

Using our REST API is super easy to do. Just use your favorite programming language's web request package and the data will be returned via JSON or CSV.

To use the REST API, you must let our server know you have an account. You can do this by passing your API Token.

There are two (2) ways to pass your API token using the REST API.

1. Pass the token directly within the request URL.
2. Pass the token in your request headers.

**1. Pass the token directly within the request URL.**

You can pass the token directly in the request URL by passing the **token** parameter. For example you would query the https://api.tiingo.com/api/test/ endpoint with the token in URL by adding ?token=. Check out the copy & paste example below.

```python
import requests

headers = {
        'Content-Type': 'application/json'
        }
requestResponse = requests.get("https://api.tiingo.com/api/test?token=<TOKEN>",
                                    headers=headers)
print(requestResponse.json())
```

**Response:**

```json
{'message': 'You successfully sent a request'}
```

**2. Pass the token in your request headers.**

You can also pass the token in the request headers. Note you can do this by passing "Token " + your API token to the "Authorization" header. Check out the copy & paste examples below.

```python
import requests

headers = {
        'Content-Type': 'application/json',
        'Authorization' : 'Token <TOKEN>'
        }
requestResponse = requests.get("https://api.tiingo.com/api/test/",
                                    headers=headers)
print(requestResponse.json())
```

**Response:**

```json
{'message': 'You successfully sent a request'}
```

## 1.2.2 Connecting to the Websocket API

Websockets allow for two-way communication, allowing us to data to you as soon as it's available. If you want real-time data, this is the fastest way to get it.

This can seem complicated if it's your first experience with websockets, but don't worry - it's just as easy with a RESTful interface and even more efficient. Websockets are both faster and uses less data than RESTful requests.

The web socket API functions a bit differently as you "subscribe" and "unsubscribe" to data sources. From there, you will receive all updates as soon as they're received without having to make requests.

Additionally, when new data comes in there will be a "messageType" which can be

- "A" for new data
- "U" for updating existing data
- "D" for deleing existing data
- "I" for informational/meta data
- "E" for error messages
- "H" for Heartbeats (can be ignored for most cases)

This lets us pass on notices that data has updated or some data is no longer considered valid Each request made to the websocket server contains a JSON object that follows the format:

**Response:**

```json
{
    'eventName': 'subscribe',
    'eventData': {
                    'authToken': '<TOKEN>',
                    'service':'test'
                }
}
```

Notice how we have to pass an authorization token just like a REST request. Also notice the HeartBeat message. This is sent every 30 seconds to keep the connection alive.

Once we get our first data request back on a successful connection, the "data" attribute will contain a **subscriptionId**. This will be used for managing the connection. For example with IEX data, if we want to add or remove tickers in our subscription, we can send updates using the **subscriptionId**.

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/test")

subscribe = {
                'eventName':'subscribe',
                'eventData': {
                            'authToken': '<TOKEN>'
                            }
                }

ws.send(json.dumps(subscribe))
while True:
    print(ws.recv())
```

**Response:**

```
#Notice the "HeartBeat" message. This is sent every 30 seconds
{"data": {"subscriptionId": 61}, "response": {"message": "Success", "code": 200}, "messageType": "I"}
{"response": {"message": "HeartBeat", "code": 200}, "messageType": "H"}
```
