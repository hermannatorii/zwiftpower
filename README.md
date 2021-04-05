# Zwift Power results automation

Grab latest results from our team members from ZwiftPower

Runs in Google Cloud Run as an http service. Trigger a new run with:

```bash
curl -H \                       
"Authorization: Bearer $(gcloud auth print-identity-token)" \
https://<service URL>/trigger
```

Environment variables on the Google Cloud Run service:

* SPREADSHEET: Google sheets ID
* LIMIT: for testing, limit the number of riders we get data for

SPREADSHEET doesn't work yet! But if you don't set it, you can get the results written to a results.csv file in the Google Cloud storage bucket. 