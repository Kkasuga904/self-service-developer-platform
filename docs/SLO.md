# SLO

## Proposed service objective

For the sample payment API, the intended 30-day objectives are:

- availability SLI: successful HTTP responses divided by valid HTTP requests;
- availability SLO: 99.9%;
- latency SLI: proportion of valid requests completed in at most 500 ms;
- latency SLO: 99% within 500 ms.

These are design targets, not measured commitments. Phase 2 exposes a minimal
Prometheus endpoint but does not install a collector, dashboard, recording rule
or alert. Therefore SLO collection and compliance are **NOT RUN**.

The platform will own collection plumbing, standard labels and a starter
dashboard. The application team owns meaningful request classification,
instrumentation correctness, dependency telemetry and service-specific target
approval.
