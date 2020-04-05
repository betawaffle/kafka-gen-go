KAFKA_VERSION ?= 2.4.1

.PHONY: all

all:
	rm -f *{Header,Request,Response}.json
	git clone https://github.com/apache/kafka.git -b $(KAFKA_VERSION) --depth 1
	cp kafka/clients/src/main/resources/common/message/*.json ./
	rm -rf kafka
