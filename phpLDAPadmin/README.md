
# Create certs for the phpldap admin console

Follow the link to create self signed certs
https://dzone.com/articles/creating-self-signed-certificate

create a configmap with certs in the namespace.

```kubectl -n <namespace> create configmap console-certs --from-file=server.crt=server.crt --from-file=server.key=server.key```


# phpLDAPadmin YAML file installation usage 

1. Docker run command
   Docker:
   
   ```docker run --name phpldapadmin-service -p 80:80 -p 8443:443 --link openldap:ldap-host --env PHPLDAPADMIN_LDAP_HOSTS=ldap-host --detach osixia/phpldapadmin:latest```
   Kubernets:
   
   ```kubectl -n <namespace> apply -f phpLDAPadmin.yaml```
2. phpLDAPadmin.yaml file contains all the resources required for the phpLDAPadmin deployment.
3. phpLDAPadmin admin username is 'admin' and passowrd is 'admin'.
4. Port 80 or 443 is exposed on ingress port 30080

login through the console using username as `cn=admin,dc=example,dc=com` and password `admin`

