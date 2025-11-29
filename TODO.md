# TODO

## Features to Implement


- [ ] Add LDAP authentication.

- [ ] Once LDAP is working, implement a database-free ACL method by reading a ConfigMap containing a structured YAML file. The app should read the YAML and understand which group can access each namespace. Each namespace allowed in the group should be able to choose whether it can be viewed, edited, or administered. We will define the scopes of each role later.
