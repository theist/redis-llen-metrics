# redis-llen-metrics

Just testing stuff. It connects to a redis server databases, extract all keys considered lists and create a metric with their current sizes

```
Usage of ./redis-llen-metrics:
  -config
    	Dump config
  -db int
    	Redis db
  -host string
    	Redis host (default "localhost")
  -loop-time int
    	Loop time in seconds
  -port int
    	Redis port (default 6379)
  -statsd
    	Activate StatsD backend
  -statsd-host string
    	StatsD host (default "127.0.0.1")
  -statsd-port int
    	StatsD port (default 8125)
  -statsd-prefix string
    	StatsD prefix (default "redis.llen")
  -statsd-suffix string
    	StatsD suffix
  -text
    	Activate text backend
  -text-filename string
    	Filename to publish stats to (default "stdout")
  -text-prefix string
    	Prefix to prepend to each stat
  -text-separator string
    	Separator to use between fields (default "\t")
  -text-suffix string
    	Suffix to append to each stat
```
