# rpapoc-auth

auth is a very simple microservice to handle user authentication

This is part of the Reverse Proxy Authentication proof of concept

### Default

user: tester
pass: password
super secure

### API

POST / {username: 'username', password: 'password'}

### Config

Run with environment variables:

* JWT_SECRET: some string
* LISTEN_PORT: tcp port to serve http
* DB_HOST: hostname of the database server (sql)

Compile for docker container:

CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w' .
