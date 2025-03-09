### Sidecar

Simple HTTP sidecar to proxy traffic from 'listen' to 'forward' and generate OTel traces in the process.
```
Usage of ./sidecar:
  -forward string
    	forward address (default "localhost:8000")
  -listen string
    	listen address (default ":8080")
```

TraceIDs are deterministic based on client IP address.

OpenTelemetry environment variables you might want to set:
```
	OTEL_SERVICE_NAME
	OTEL_EXPORTER_OTLP_ENDPOINT
	OTEL_RESOURCE_ATTRIBUTES
```


.. And that is pretty much it.
