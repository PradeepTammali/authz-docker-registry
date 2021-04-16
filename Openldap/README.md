# Openldap YAML file installation usage 

1. Docker run command
	
	Docker:
	```
	docker run --name openldap -p 389:389 -p 636:636  --env LDAP_ORGANISATION="My Example Organization" --env LDAP_DOMAIN="example.com" --env LDAP_ADMIN_PASSWORD="admin"  -d osixia/openldap:latest
	```
	Kubernetes:
	```
	kubectl apply -f openldap.yaml
	```
2. openldap.yaml file contains all the resources required for the openldap deployment.
3. Openldap admin username is 'admin' and passowrd is 'admin'.
4. Port 389 is exposed on ingress port 
5. To query the data inside the ldap use the following command.
   ```
   ldapsearch -x -H ldap://localhost:389 -b dc=example,dc=com -D  "cn=admin,dc=example,dc=com" -w admin
   ```
6. Add users and units by adding the schema 'ldap-schema.ldif' by logging inside the pod with following command.
   ```
   ldapadd -c -H ldap://localhost:389 -x -D "cn=admin,dc=example,dc=com" -w admin -f ldap-schema.ldif
   ```
7. Search users in specific OU or in whole sctructure with filter 
   ```
   ldapsearch -x -H ldap://localhost:389 -b ou=musicians,dc=example,dc=com -D "cn=admin,dc=example,dc=com" -w admin "(&(uid=bach)(objectClass=organizationalPerson))"
   ```
   OR
   ```
   ldapsearch -x -H ldap://localhost:389 -b dc=example,dc=com -D "cn=admin,dc=example,dc=com"   -w admin "(&(uid=bach)(objectClass=organizationalPerson))"
   ```
8. To change the password of a User
   ```
   ldappasswd -H ldap://localhost:389 -x -D "cn=admin,dc=example,dc=com" -w admin -s password "uid=kant,ou=philosophs,dc=example,dc=com"
   ```


Reference:
https://github.com/kwk/docker-registry-setup
   
