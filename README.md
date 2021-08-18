# datastore-experiment
This is a project for testing out read times in Google App Engine Datastore

## Setup
Make sure you have a Google Cloud Project setup
See:
* https://cloud.google.com/appengine/docs/standard/go/quickstart
* https://khanacademy.atlassian.net/wiki/spaces/TESTPREP/pages/1766162433/Creating+a+Test+Google+Cloud+App

Edit the app.yaml file to include your project ID for `GCLOUD_DATASET_ID`:
```
runtime: go115

env_variables:
  GCLOUD_DATASET_ID: YOUR_PROJECT_ID_HERE
```

## Check Your Code Before Deploy
```
go test -count=0
```

## Deploy
Again make sure to type your project ID in place of `YOUR_PROJECT_NAME_HERE`:
```
gcloud app deploy --project=YOUR_PROJECT_ID_HERE
```

## Run/View Results
```
gcloud app browse --project=YOUR_PROJECT_ID_HERE
```

## Check Entities
(be sure to select your project from the top dropdown)
https://console.cloud.google.com/datastore/entities
