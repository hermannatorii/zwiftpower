clean:
	rm zwiftpower

service: container
	gcloud run deploy zwiftpower \
--image=gcr.io/coherent-parity-304720/zp:latest \
--platform=managed \
--region=us-central1 \
--project=coherent-parity-304720

container: zwiftpower
	docker build -t gcr.io/coherent-parity-304720/zp .
	docker push gcr.io/coherent-parity-304720/zp 

zwiftpower: *.go zp/*.go
	GOOS=linux go build .

local: *.go zp/*.go
	go build .

