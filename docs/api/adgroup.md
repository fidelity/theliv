## POST theliv-api/v1/adgroup/{cluster}/{namespace}

### Description
This operation is used to assign the adgroups the right to detect issues in the specified cluster/namespace.

### Header
* **ACCESSKEY** - for authentication

### Path Parameter
* **cluster**
* **namespace**

### Request Body
```
{"adgroups": ["adgroup1","adgroup2"]}
```
### Sample request
```
curl --location --request POST 'http://theliv-endpoint/theliv-api/v1/adgroup/cluster/namespace' \
--header 'ACCESSKEY: xxxx' \
--header 'Content-Type: application/json' \
--data-raw '{"adgroups": ["adgroup1","adgroup2"]}'
```
### Successful response
```
Status Code: 200
Response Body: "Cluster/NS added to adgroups"
```
### Authorization

Theliv admin are maintaining an allow-list of cluster/namespaces, for the specified ACCESSKEY.
Any update to the allow-list, please contact with Theliv admin team.

___

## DELETE theliv-api/v1/adgroup/{cluster}/{namespace}

### Description
This operation is used to remove the adgroups the right to detect issues in the specified cluster/namespace.

### Header
* **ACCESSKEY** - for authentication

### Path Parameter
* **cluster**
* **namespace**

### Request Body
```
{"adgroups": ["adgroup1","adgroup2"]}
```
### Sample request
```
curl --location --request DELETE 'http://theliv-endpoint/theliv-api/v1/adgroup/cluster/namespace' \
--header 'ACCESSKEY: xxxx' \
--header 'Content-Type: application/json' \
--data-raw '{"adgroups": ["adgroup1","adgroup2"]}'
```
### Successful response
```
Status Code: 200
Response Body: "Cluster/NS removed from adgroups"
```
### Authorization

Theliv admin are maintaining an allow-list of cluster/namespaces, for the specified ACCESSKEY.
Any update to the allow-list, please contact with Theliv admin teams.