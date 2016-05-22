# revel-apikit
A way to rapidly develop REST APIs using the Revel framework

#### Limitations
- `RESTController` instances cannot have Actions other than those provided by `RESTController`
- `conf/restcontroller-routes` cannot use catchall `:` Actions
- `conf/restcontroller-routes` cannot use wildcard `*` paths
