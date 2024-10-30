# Spike take home assignment

To run the svc:
```bash
docker compose up -d
source .env
go run ./cmd/main.go
```

Some tests:
```bash
# Create widgets
curl -X POST -d '{"name": "widget1", "serial_number": "123", "ports": ["P", "R"]}' localhost:8080/widgets
curl -X POST -d '{"name": "widget1", "serial_number": "321", "ports": ["P", "R"]}' localhost:8080/widgets

# Create association
curl -X PUT -d '{"port_type": "P", "widget_serial_num": "123", "peer_widget_serial_num": "321"}' localhost:8080/widgets/associations
```


# Architecture, design notes and ideas below

- Data assumptions and back of the envelope calculations:
    - Widgets:  
      - Name should be a reasonably short strings - let's say up to 64 ASCII chars - 64bytes
      - Serial number could be a UUID or other string let's say up to - 16 bytes
        - we can assume that serial_num is unique and will most probably be used for querying and doing updates.
      - Ports could be a bitmask of 3 types (PRQ), 1 byte 
      - Total for single record: 81 bytes
      - Total for 10mil: 10*10^6 * 81 bytes = ~810MB of data to store all
        records. Would easily fit in memory if needed for example for caching
        purposes in redis.
    - Connections:
      - Up to 3 connections per widget. Would need to include 2 serial_num fields
        and 1 port type field.
      - Total for single record: ~33 bytes
      - Avg 2 conns per widget: 10*10^6 *2 *33 bytes = ~660MB
- Nothing is mentioned about reads
- Update heavy workload - thousands of updates per second. We can assume this
  will be mostly for the connection associations.
- 10 million widget entries:
    - Should be reasonable to store in a single postgres node with appropriate
      indexes for serial_num and id.
    - Cassandra would be optimal if we were having more append only write-heavy
      workloads, since this case is more a point update scenario, we would start
      running into issues with tombstone markers and frequent compactions and
      potentially degraded performance.
- Database schema:
    - Since not all widgets have all ports types, in order to prevent row level
      locking when doing updates it makes sense to separate the connection
      tables by port type.

- Thousands of updates per second. Let's assume it is up to 10K/s
    - We can assume that updates are for the port connections and not for the
      widget itself. Logically, serial_num and name should be immutable.
    - We cannot directly send individual updates to db as this would be overkill
      for a traditional RDBMS in terms of network latency and IOPS and we would
      need to start thinking about sharding the db immediately - too much
      hassle. We could get by with a simple db structure and a cache layer in
      front of it.
- Optimal POC solution:
    - Postgres as persistent source of truth for widgets and connection associations.
    - Redis cluser in fron of db in the same availability zone where the
      stateless API server instances are deployed. Redis would serve as an in
      front cache for writes and reads. Writes could be updated via the data
      reconciliation service or component within API.
  
# Service architecture
High level overview of system

## Data
- Postgres as main source of eventually consistent data for storing widgets and their connections
  - HA with 3 master nodes and a couple read replicas
    - RDS should be fine for less hassle
  - Should be deployed in same region where the stateless services are deployed
  - For hard consistency with direct writes to db - sharded HA cluster would be
    needed to achieve desired scale of updates. Shard key would be serial_num.
- Redis as a cache layer in front of Postgres for alleviating write pressure and acting as a read cache when needed
  - Single beefy machine would be enough, but for HA - redis cluster would be better.

###  Redis data model
- Widget cache (HASH) for fast widget lookups.
  - Hash key: `w:<serial_num>`
    - Fields:
      - `name`: string
      - `ports`: bitmask of available ports
- Widget connection associations cache (HASH):
  - Hash key: `c:<serial_num>`
    - Fields:
      - p_peer_sn: serial_num of connected widget to port P
      - r_peer_sn: serial_num of connected widget to port R
      - q_peer_sn: serial_num of connected widget to port Q
- Recently updated widgets list (LIST)
  - Key: `recently_updated_widgets`
    - Values: `[<serial_num>, ...]`
  - SET vs LIST rationale: set would ensure that serial_num values do not repeat
    (no multiple updates for single widget). However, using sscan is cumbersome when it comes to 
  - A set of recently updated widget serial_nums. Data reconciliation service
    SSCANS and retrieves list of serial nums along with their corresponding port
    connections and updates the Postgres db with the new data. For atomicity
    sscan and srem are ran in single eval.
  - SSCAN vs LIST reasoning: we might have multiple updates for same widget in a
    single update batch, therefore it makes sense to utilize set here.

# REST API
- Stateless API service which: 
  - writes connection association updates to redis cache
  - reads/writes connection associations to redis cache   
  - writes widget 
- Data reconciliation service:

## **POST /widgets - create a new widget**

Payload:
```json
{
    "name": "widget_name",
    "serial_number": "widget_serial_num",
    "ports": ["P", "R", "Q"]
}
```

Provided `ports` array should contain 1-3 available port types from (P, R, Q)
list.

## **DELETE /widgets/{serial_num} delete a widget**

## **PUT /widgets/associations - connect a widget to another widget**
Payload: 
```json
{
    "port_type": "P",
    "widget_serial_num": "<widget_serial_num>",
    "peer_widget_serial_num": "<widget_serial_num>"
}
```

# Deployment strategy

- Stateless services: K8S or ECS equivalent with autoscaling based on QPS for the API service
  and Data reconciliation servies. Simple rolling deployment updates should be
  sufficient.
- Postgres HA in same region as stateless services
  - Depending of resouce usage, we could also use pg bouncer for conn pooling if
    many reads start hitting the db.
- Redis (ValKey, KeyDB, etc) cluster in the same zone as the API service instances 
  - Configure with AOF persistence for data durability.

# Additional notes

- For the sake of time, no tests are provided
- Integration tests would be most appropriate for this services since there is
  not too much business logic to test.
    - Integration test plan could include spinning up the API service with pg
      and redis deps with testcontainers and running a series of API calls to
      test the functionality and correctness of the service. 
    - Load/Benchmark testing 

# TODO
[] Redis cache component
[] Db layer
[] Http api component
[] Data reconciliation component