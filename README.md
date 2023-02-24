# Addrbook

Copyright 2020-2023 Demian Harvill

## Overview

Addrbook is a microservice for describing and maintaining address books of people and businesses.
It is written in Go, and uses [gRPC](https://grpc.io) to define and implement it's application programming interface (API).
The server requires a JSON Web Token (JWT) generated by the [MService](https://github.com/gaterace/mservice) microservice
for authorization.

## Usage

Example client usage using the Go command line client (note that any thin client in any language supported by 
gRPC can be used instead):

**addrclient create_party --ptype person --fname Frodo --mname B  --lname Baggins -e frodo@baggins.org**

Creates a person record for Frodo Baggins.  Returns the integer party id.

**addrclient create_party --ptype business --company FrodoCorp -e frodocorp@baggins.org**

Creates a business record for FrodoCorp.  Returns the integer party id.

**addrclient get_parties**

Gets a list of all parties bound to the mservice account (based on the context JWT).

**addrclient get_party --id 7**

Gets the record for the party identified by party id 7 within the mservice account.

**addrclient get_party_wrapper --id 7**

Gets the record for the party identified by party id 7 within the mservice account, as well as any 
child address or phone records.

**addrclient create_address --id 7 --atype home --address1 '123 Main St' --city Anytown --state NV --postal_code 12345**

Creates a child address record of type home for the party identified by party 7.

**addrclient get_address --id 7 --atype home**

Get the home address record for the party identified by party id 7.

**addrclient create_phone --id 7 --phtype cell --phone 543-555-1212**

Creates a child phone record of type cell for the party identified by party 7.

**addrclient get_phone --id 7 --phtype cell**

Get the cell phone record for the party identified by party id 7.

**Other commands** for operations (eg. get, update, delete) can be discovered with 

**addrclient**

with no parameters. 

 
## Certificates

### JWT Certificates
The generated JWT uses RSA asymmetric encryption for the public and private keys. These should have been generated
when installing the MService microservice; in particular, the mproject server needs access to the jwt_public.pem public key.

### SSL / TLS Certificates

In a production environment, the connection between the client and the MService server should be encrypted. This is
accomplished with the configuration setting:

    tls: true

If using either a public certificate for the server (ie, from LetsEncrypt) or a self-signed certificate,  the server need to know the public certificate as
well as the private key. 

The server configuration is:

    cert_file: <location of public or self-signed CA certificate

    key_file: <location of private key>

The client configuration needs to know the location of the CA cert_file if using self-signed certificates.

## Database

There are MySql scripts in the **sql/** directory that create the addrbook database (addrbook.sql) as well as all
the required tables (tb_*.sql).  These need to be run on the MySql server to create the database and associated tables.

## Data Model

The persistent data is managed by a MySQL / MariaDB database associated with this microservice.

No data is shared across MService accounts.

The root object is a **party**, which is associated with a single MService account.

Each party can have zero or more child **address** objects differentiated by address type.
  
Each party can have zero or more child **phone** objects differentiated by phone type.


## Server

To build the server:

**cd cmd/addrserver**
  
**go build**

The addrserver executable can then be run.  It expects a YAML configuration file in the same directory named **conf.yaml** .  The location of the configuration file can be changed with an environment variable,**ADDR_CONF** . Configuration can also be
specified by command line flags or by environment variables (with ADDR_ prefix).

```
Usage:
  addrserver [flags]

Flags:
      --cert_file string      Path to certificate file.
      --conf string           Path to inventory config file. (default "conf.yaml")
      --db_pwd string         Database user password.
      --db_transport string   Database transport string.
      --db_user string        Database user name.
  -h, --help                  help for addrserver
      --jwt_pub_file string   Path to JWT public certificate.
      --key_file string       Path to certificate key file.
      --log_file string       Path to log file.
      --port int              Port for RPC connections (default 50057)
      --tls                   Use tls for connection.
```

A commented sample configuration file is at **cmd/addrserver/conf.sample** . The locations of the various certificates and 
keys need to be provided, as well as the database user and password and the MySql connection string.

## Go Client

A command line client written in Go is available:

**cd cmd/addrclient**

**go install** 
    
It also expects a YAML configuration file in the user's home directory, **~/.addrbook.config**. A commented sample for this
file is at **cmd/addrclient/conf.sample**

Running the executable file with no parameters will write usage information to stdout.  In particular, all subcommands expect
the user to have logged in with Mservice acctclient to establish the JWT. The JWT is also used to determine which
account is being used for the command.

Note that the use of the Go addrclient is merely a convenience, and not a requirement. Since we are using gRPC, the thin client
can be written in any supported language.  It can be part of a web or mobile application for example.


## Claims and Roles ##

The addrbook microservice relies on the **addrsvc** claim, and the following claim values:

**addradmin**: administrative access

**addrrw**: read-write access to addrbook objects 

**addrro**: read-only access to addrbook objects 


Note that within an account in Mservice, a role must be created to map these claims to a logged-in user.

















