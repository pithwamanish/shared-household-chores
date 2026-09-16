import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { DocumentLoadInstrumentation } from '@opentelemetry/instrumentation-document-load';

// 1. Mandatory Golden Triangle Resource Attributes
const serviceName = import.meta.env.VITE_OTEL_SERVICE_NAME || 'web-frontend';
const environment = import.meta.env.VITE_APP_ENV || import.meta.env.MODE || 'development';
const serviceVersion = import.meta.env.VITE_APP_VERSION || 'dev-latest';

// Dynamic exporter target: use env var or calculate collector URL based on current browser host
const defaultCollectorURL =
  typeof window !== 'undefined'
    ? `${window.location.protocol}//${window.location.hostname}:4318/v1/traces`
    : 'http://localhost:4318/v1/traces';

const exporterUrl = import.meta.env.VITE_OTEL_EXPORTER_URL || defaultCollectorURL;

const provider = new WebTracerProvider({
  resource: new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: environment,
    [SemanticResourceAttributes.SERVICE_VERSION]: serviceVersion,
  }),
});

// 2. Export traces to OpenTelemetry Collector over HTTP OTLP
provider.addSpanProcessor(
  new BatchSpanProcessor(
    new OTLPTraceExporter({
      url: exporterUrl,
    }),
    {
      maxExportBatchSize: 32,
      scheduledDelayMillis: 1000,
    }
  )
);

provider.register();

// 3. Register auto-instrumentations with W3C traceparent CORS propagation
registerInstrumentations({
  instrumentations: [
    new DocumentLoadInstrumentation(),
    new FetchInstrumentation({
      propagateTraceHeaderCorsUrls: [
        new RegExp(window.location.origin),
        /localhost:\d+/,
        /127\.0\.0\.1:\d+/,
        /\/api\/.*/,
      ],
      clearTimingResources: true,
    }),
  ],
});

export const tracer = provider.getTracer('choresync-web');
