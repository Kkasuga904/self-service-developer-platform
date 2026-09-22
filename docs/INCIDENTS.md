# Incident scenarios

No incident experiment has been executed in Phase 2. Phase 4 will create
evidence-driven records for a bad image/probe deployment, manual Git drift, a
policy violation, and Argo CD unavailability. Until then all four are **NOT
RUN**, and no recovery time or availability result is claimed.

The central hypothesis for an Argo CD failure is that existing application Pods
continue to serve while new reconciliation and deployment visibility stop. It
must be tested before being reported as observed behavior.
