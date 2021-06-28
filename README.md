# reactive-micro
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![made-with-Go](https://img.shields.io/github/go-mod/go-version/AyushSenapati/reactive-micro/main?filename=authnsvc%2Fgo.mod&label=Go)](http://golang.org)
[![NATS streaming](https://img.shields.io/badge/NATS-Jet%20Stream-yellowgreen)](https://nats.io/)

This demonstrates how event driven micro services can be built. Event driven systems establishes boundary bewteen components which ensure loose coupling and isolation. With increasing systems managing consistent events become difficult. So like OpenAPI specification, we can maintain our contracts in [AsyncAPI](https://www.asyncapi.com/) specification. Non-blocking communication using persistent event bus allows services not to worry about other active services. They can just fire and forget and other services will process when they become active.

## services
* `authnsvc`: responsible for accounts management and authentication
* `authzsvc`: it stores all the authorization policies in its persistent storage. It provides APIs, so that based on events services can upsert or delete policies.
* `ordersvc`: manages orders
* `inventorysvc`: manages inventory
* `paymentsvc`: deals with payments

`authzsvc` implements ACL based authorization which provide granular control over the resources than RBAC systems. All possible policies for the resources are stored in this service. It follows who (subject) can perform what (action) on which resource (object) mechanism.  
format `sub:action:resource_type:resource_id`  
ex: 10:get:orders:15 means 10 can get/read order having ID 15.  
Internally it stores policies in different data structure optimised for querying.

Optimisations:  
On successful authentication, authnsvc fires event-account-authenticated. authzsvc uses this event to load all the policies of that account to its cache. Each service is wired with a local cached authorization(authz) library which holds policies related to the specific service. When an authenticated request hits a service, its authz library checks its local cache, on cache miss it queries authzsvc for the policies required by this specific service and caches those fetched policies for some time.  
On policy update, authzsvc fires event-policy-updated and the local authz libraries of all the services update their cache if required.  

Benefits of this authorization architecture is, every time a request comes in, services do not need to query the database and join multiple tables which might even scattered across different services to determine if the request is authorized. instead using the pre generated policies authz middlewares can decide whether to allow/deny the request with out even sending it to the service layer.

## Events
Followings are the events supported by these microservices. For more on these events check [events.json](events.json) file.
|Name|Description|
|------|-----------|
|`event-account-created`|fired when an account is created successfully|
|`event-account-deleted`|fired when an account is deleted. subscribers can use this information to clean up their resources associated with this account|
|`event-account-authenticated`|fired on successful authentication of an account. Can be used to improve performance of the system by preparing cache even before the actual authenticated request comes in|
|`event-upsert-policy`|fired to create/update new/existing authorization policy. A service can use this event to create/update an policy when a resource is created/updated|
|`event-policy-updated`|fired when an authorization policy changes for a subject. This can be used by all the services to update their local authz cache|
|`event-remove-policy`|can be fired to remove an authorization policy. A service can use this event to remove policies associated with the resources on resource deletion|
|`event-order-created`|ordersvc fires this event when an order is created. The svc itself does not check the validity of the product details.|
|`event-order-canceled`|ordersvc fires this event when an order is canceled may be due to payment failure or user cancels the order. services can consume this event to revert their order specific changes|
|`event-order-approved`|ordersvc fires this event when an order is placed successfully and ready for shipment|
|`event-product-reserved`|inventorysvc checks the validity of the event-order-created and tries to reserved requested product. on success it fires this event|
|`event-err-reserving-product`|if inventory service fails to reserve requested product for the user, this event is fired|
|`event-payment`|upon receiving event-product-reserved payment service tries to deduct the payble from the user account. this event is fired to indicate payment success/failure|
|`event-suspicious-activity`|can be fired by any of the services to indicate unusual activity for further investigation|

## License:
[MIT Licence](LICENSE)
