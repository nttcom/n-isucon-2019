#!/bin/bash

gcloud compute instances list --filter="labels.type=app" --format="table(name)" | tail -n +2 | xargs gcloud compute instances reset --zone=asia-northeast1-c
