docker build -t dev.local/ktwin/mqtt-dispatcher:0.1 --build-arg SERVICE_NAME=mqtt-dispatcher .
docker build -t dev.local/ktwin/cloud-event-dispatcher:0.1 --build-arg SERVICE_NAME=cloud-event-dispatcher .
kind load docker-image dev.local/ktwin/mqtt-dispatcher:0.1
kind load docker-image dev.local/ktwin/cloud-event-dispatcher:0.1