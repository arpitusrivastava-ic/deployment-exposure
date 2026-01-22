# Deployment Exposure Controller (Kubernetes)

A Kubernetes controller that automatically exposes Deployments outside the cluster using Service and Ingress resources.

This project demonstrates:
- Kubernetes controller-runtime fundamentals
- Informers, caches, and RBAC
- Lifecycle synchronization using OwnerReferences
- A fully containerized, local demo using Minikube

---

## Problem Statement

Manually creating Services and Ingresses for every Deployment is repetitive and error-prone.

The goal of this controller is to:
- Automatically expose Deployments when explicitly requested
- Keep exposure resources in sync with the Deployment lifecycle
- Use Kubernetes-native patterns without introducing CRDs

---

## Design Overview

### Trigger Mechanism

Deployments opt-in using annotations:

```yaml
metadata:
  annotations:
    expose: "true"
    expose/type: "ingress"   # optional, defaults to service-only
