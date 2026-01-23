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

## Validation Status

###  Verified
- Deployment reconciliation
- Automatic Service creation
- Automatic Ingress creation
- Lifecycle sync (delete Deployment → Service & Ingress deleted)
- Application reachable via Service

### macOS Limitation (Expected)
When running Minikube with the Docker driver on macOS,
Ingress traffic may not reach the ingress controller due to
Docker Desktop networking limitations.

### Workaround Used
```bash
kubectl port-forward svc/demo-nginx-svc 8080:80
curl http://localhost:8080
