# Kurajj charity platform 

## Prerequisites

* Installed Docker and docker-compose in you machine
* Go version >= 1.18
* `configs` folder with all configs data
* gradle >= 7.4 

## About

In this project build tool is Gradle, to run a command you will have to do it in the next way - `./gradlew command`.
[IN FUTURE] After sometime http will be changed to https, so you will need to trust certs in your browser.

## Project structure

In docs folder you are able to see swagger documentation for API.

## Run the project locally

* `docker login` - put your credentials here
* `./gradlew gc-up`

Now api is available locally on localhost:8080/ URL
