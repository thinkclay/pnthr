[![Build Status](https://travis-ci.org/pnthr/pnthr-go.svg?branch=master)](https://travis-ci.org/pnthr/pnthr-go)

# pnthr API with Golang

This simple app takes transport data, decrypts and encrypts that data with a custom secret to be sent back to the sender.

## About the application
Pnthr is a service that was created out of necessity. We don’t claim to be the most enterprise, or the most secure data encryption service, just one of the easiest drop in solutions for most applications to take their sensitive data to the next level of secure.

### The service works like this:

You need to ask a user in your application for a social security number so that annually you can file the proper tax forms. You don’t want to keep a plain text social security number in your database for fear that if you’re ever hacked and penetrated that data is easily stolen. Your default reaction would instead be to probably encrypt the data with a key and keep the key on a different server (like the database server) so that your database server would ALSO have to be hacked in order to decrypt the sensitive data. The problem with this solution is that your web server probably has some sort of access to your database server already creating a security vulnerability. By utilizing a third party api you could mitigate some of this risk by forcing your would-be-hacker to not only hacking into your servers but also the “cloud” or third party servers hosted by pnthr. This is the core of what pnthr is and does. It’s simple, it’s easy, it’s scalar and that makes for the best security.

### Run the server

To run the server locally:
`MONGO_DB=pnthr go run server.go`

You'll need to create a database called `pnthr` with a collection called `instances` with this document for tests:

```
{
  "_id": ObjectId("53a1c59f6239370002000000"),
  "name": "Test for Ruby Gem Spec",
  "description": "password is 'password'",
  "password": "5f4dcc3b5aa765d61d8327deb882cf99",
  "secret": "aa88906ffcf6c59aaf5908d3900f21a6",
  "user_id": ObjectId("535df0646636640002000000"),
  "updated_at": new Date(1403110815507),
  "created_at": new Date(1403110815507)
}
```

## Encryption and Transport

### Transport Send
1. Encrypt - Client encrypts payload with App Secret
2. Encode - Encrypted payload is base64 encoded
3. Send - Payload is sent to pnthr over SSL

### Pnthr Receive
1. Decode - Pnthr decodes base64
2. Decrypt - Payload is decrypted with App Secret
3. Encrypt - Payload is encrypted with App Password (only known by Pnthr)
4. Encrypt - A second transport encryption is added (again with the app secret)
5. Encode - The payload is then base64 encoded
6. Send

### Transport Receive
1. Receive - Clients receives payload
2. Decode (optional) - Payload can be stored as base64 or decoded
