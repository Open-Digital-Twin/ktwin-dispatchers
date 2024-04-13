
docker-build:
	docker build -t ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1 --build-arg SERVICE_NAME=mqtt-dispatcher .
	docker build -t ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1 --build-arg SERVICE_NAME=cloud-event-dispatcher .

docker-push:
	docker push ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1
	docker push ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1