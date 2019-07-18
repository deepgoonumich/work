# ftp-engine

`ftp-engine` consumes pipeline content, transforms to proper format and sends it to configured FTP Destination. It supports consuming a content queue from `Kafka` and uses `groups` to allow for multiple workers. This means that running several instances will not consumer/send the same content multiple times if the instances use the same `KAFKA_GROUP_ID`. *`ftp-engine` is intended to run a single instance per destination.* Interfaces are used throughout the project to allow for changing the receiver/worker types or swapping FTP for another output in the future.

The main process initializes the FTP Sender, Processor(fitlers,converts to output format), and other dependencies and loads the worker. The worker process continually pulls from Kafka, calls Convert, then sends content out using the Sender before acknowledging the message.

There are two processes for `ftp-engine`, worker and updater. The *worker* process processes content from the pipeline and outputs via FTP. The *updater* process handles periodic refresh from `refDB` and inserts the ticker data into Redis for caching.

## Run

### Deployment

Managed by Helm on Kubernetes. See [https://gitlab.benzinga.io/benzinga/helm-deployments](https://gitlab.benzinga.io/benzinga/helm-deployments). Values for deployments are in Keybase `helm-secrets`.

### Configuration

*Unless noted as optional, all parameters are required.* Supported compression `gzip`,`lz4`, and `zstd`.

#### Env

 - `DEBUG`: `true`|`false`
 - `ENVIRONMENT`: `production`|`staging`|`development`|`testing`
 - `LISTEN_HOST`: `0.0.0.0` *you probably shouldn't change this*
 - `LISTEN_PORT`: `9000` *you probably shouldn't change this either*

 - `PROCESSOR`: `ravenpack`,`default`.
 - `PROCESSOR_EVENTS`: Based on `content-models`:`EventType` which are, as of writing, `Created`,`Updated`, and `Removed`.

 - `FTP_HOST`: `127.0.0.1:21`
 - `FTP_PATH`: `/home/ftpuser`
 - `FTP_USERNAME`: `ftpuser` *(optional)*
 - `FTP_PASSWORD`: `ftppass123` *(optional)*
 - `FTP_CONNECT_TIMEOUT`: `10s`
 - `FTP_KEEPALIVE_INTERVAL`: `10s`,`30m`,`1h` *(optional)*
 - `FTP_SEND_RETRIES`: `1` *(optional)* default `0`, set `0` to disable.

 - `KAFKA_BROKERS`: `kafka1:19092,kafka2:29092,kafka3:39092` // should be single entry or comma-seperated list
 - `KAFKA_TOPIC`: `ftp-engine`
 - `KAFKA_GROUP_ID`: `client-ftp-1` * see note above about how Kafka handles consumer groups.*
 - `KAFKA_USERNAME`: `bz-kafka-user` *(optional)*
 - `KAFKA_PASSWORD`: `password` *(optional)* supplying both user & pass attempts auth using scram-sha256
 - `KAFKA_TLS_CA`: `/path/to/ca.pem` *(optional)* supplying CA without client cert or key attempts plain TLS
 - `KAFKA_TLS_CERT`: `/path/to/client.cert` *(optional)*
 - `KAFKA_TLS_KEY`: `/path/to/client.key` *(optional)* supplying both key & cert attemts use of mTLS auth

 - `REFDB_ENDPOINT`: `http://data-api/refdb.json`
 - `REFDB_UPDATE_INTERVAL`: `10s`,`30m`,`1h`

#### Local

Must set `MY_IP` environment variable, this should be your *LAN IP*, configured this way to allow you to connect to Kafka from the local machine, but Docker `localhost` is a bit complicated. Use `export MY_IP=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1')` on Mac/Linux.

##### Testing

 - `make test`
 - `docker-compose down`

### Logging/Monitoring

In Kubernetes, logs are written to std.Error which is passed to Logentries.

#### Notes

##### FTP Docker
Creating FTP User/Password Docker, attached to container ex.`docker exec -it <container_id> /bin/bash`, then `pure-pw useradd benzinga -f /etc/pure-ftpd/passwd/pureftpd.passwd -m -u ftpuser -d /home/ftpusers/benzinga` in shell. Info [https://github.com/stilliard/docker-pure-ftpd](https://github.com/stilliard/docker-pure-ftpd).

#### Future Improvements

  [x] Support Kafka TLS Auth

