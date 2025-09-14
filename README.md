# Red Hat Identity Management Web Invitations 

 [![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

**NetID Activate** is a Go web app that provides controlled account invitation 
account activation via web portal. It acts as a companion to 
Red Hat Identity Management (Red Hat IdM) / FreeIPA.

It enables users to add accounts to a Red Hat IdM domain without needing 
[view ACI](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/8/html/configuring_and_managing_identity_management/access-control-in-idm_configuring-and-managing-idm#access-control-instructions-in-idm_access-control-in-idm), trusting the IdM integrated root certificate on browsers, or installing IdM 
with [external CA as the root CA](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/installing_identity_management/installing-an-idm-server-with-integrated-dns-with-an-external-ca-as-the-root-ca).

**NetID Activate** can be installed separately from Red Hat IdM and does not have 
run on a domain joined host. All calls to IdM are over the **[JSON-RPC](https://wikipedia.org/wiki/JSON-RPC) API**.

<!-- 850 900 -->

Here is what the web portal looks like:

<p align="center">
    <img src="docs/img/02_invite.png" width="350">
</p>

<p align="center">
    <img src="docs/img/01_landing.png" width="270">
    <img src="docs/img/03_claim_01.png" width="270">
    <img src="docs/img/03_claim_03.png" width="270">
</p>


## Configuration

### Session 
`SESSION_KEY`: Set the 32 byte cookie secret key. eg `openssl rand -base64 32`  

### Hostname and Port
`SERVER_PORT`: What port on localhost Go should listen to  
`SERVER_HOSTNAME`: Used in emails and OpenID Connect    
`OIDC_SERVER_PORT`: What port OpenID Connect redirect to and what users use  

### OpenID Connect
`OIDC_WELL_KNOWN`: URI of well-known endpoint    
`CLIENT_ID`: Client ID  
`CLIENT_SECRET`: Client secret   
`SCOPES`: OIDC Scopes  

### Data
`DB_PATH`: Relative path to db  

### Display Images
`LOGO_URL`: Fully qualified URI to image  
`FAVICON_URL`: Fully qualified URI to favicon   

### Display Text 
`SITE_NAME`: Shown in emails and header   
`TENANT_NAME`: Name of IdM directory, shown in emails  
`AFFILIATION`: JSON list, added to GECOS   
`LOGIN_REDIRECT`: Shown to users after successful account activation  

### Email 
`EMAIL_FROM`: Name and address  
`AWS_REGION`: AWS  
`AWS_ACCESS_KEY_ID`: SES Credentials  
`AWS_SECRET_ACCESS_KEY`: SES Credentials  

### Red Hat IdM 
`CACERT_PATH`: Absolute path to CA  
`IDM_HOST`: FQDN of IdM host  
`IDM_USERNAME`: IdM Username  
`IDM_PASSWORD`: IdM Password  
`IDM_ADD_GROUP`: Comma separated groups to add new users to (no spaces)  


## License  

NetID Activate is distributed under [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.txt).

