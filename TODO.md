# TODO

## Features to Implement

- [ ] Change or set Prometheus URL in settings. After setting it, reload the app to display metrics.

- [ ] Change password from settings. Show a warning that the user will be logged out on the change screen. After making the change, reload the app and prompt for login again.

- [ ] Add LDAP authentication.

- [ ] Once LDAP is working, implement a database-free ACL method by reading a ConfigMap containing a structured YAML file. The app should read the YAML and understand which group can access each namespace. Each namespace allowed in the group should be able to choose whether it can be viewed, edited, or administered. We will define the scopes of each role later.
