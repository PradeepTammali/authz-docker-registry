## Running keycloak on kubernetes with docker enabled mode

1. keycloak.yaml file contains all the resources required for the keycloak deployment.
2. Keycloak admin username is 'admin' and password is 'password'.
3. Docker feature is deault enabled in this deployment.
4. Port 8080 is exposed on 30005 port of ingress on http.

#### References
Please read this blog for more info.
https://medium.com/@wilson.wilson/manage-docker-registry-auth-with-keycloak-e0b4356cf7d0

## Running keycloak on https mode.
1. Install k8s-keycloak.yaml as following, it exposes http on 30004 and https on 30005 respectively.
   kubectl apply -f oidc-keycloak.yaml
2. Generate self signed certs and get them certified by third party. You can do the in the following.
   http://www.cacert.org/
   Generate self signed certs
   https://gist.github.com/fntlnz/cf14feb5a46b2eda428e000157447309#create-a-certificate-done-for-each-server
3. Exports the certs into pkcs12 keystore as following.
   openssl pkcs12 -export -name domain.com -in server.crt -inkey server.key -out serverkeystore.p12
4. Import created pkcs12 keystore into a jks keystore.
   keytool -importkeystore -destkeystore keycloak.jks -srckeystore serverkeystore.p12 -srcstoretype pkcs12 -alias domain.com
5. Copy the keycloak.jks file to configuration folder and run the following in jbosscli to change the configuration of keycloak.
   /core-service=management/security-realm=UndertowRealm:add()
   reload
   /core-service=management/security-realm=UndertowRealm/server-identity=ssl:add(keystore-path=keycloak.jks, keystore-relative-to=jboss.server.config.dir, keystore-password=password)
   reload
   /subsystem=undertow/server=default-server/https-listener=https:write-attribute(name=security-realm, value=UndertowRealm)
   reload

#### References:
https://github.com/keycloak/keycloak-documentation/blob/master/server_installation/topics/network/https.adoc
https://developer.jboss.org/thread/278360


## Securing application with Keycloak Gatekeeper
1. Configure the keycloak_gatekeeper_nginx.yaml yaml according to your application, change the keycloak details and run.
   kubectl apply -f keycloak_gatekeeper_nginx.yaml
2. The above file includes authentication and authorization for the nginx service.


#### References:
https://www.keycloak.org/docs/latest/securing_apps/#configuration-options
https://medium.com/@carlosedp/adding-authentication-to-your-kubernetes-front-end-applications-with-keycloak-6571097be090
