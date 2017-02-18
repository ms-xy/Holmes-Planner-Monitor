all:
	protoc --go_out=generated-go --java_out=generated-java messages.proto
	go install ./generated-go
	cp -r generated-java/org/holmesprocessing/totem/monitoring/* ../java/org/holmesprocessing/totem/monitoring/
