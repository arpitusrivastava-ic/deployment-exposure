package main

import (
	"context"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type ExposerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ExposerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Deployment not found, skipping", "name", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Only act on annotated deployments
	if deployment.Annotations == nil || deployment.Annotations["expose"] != "true" {
		return ctrl.Result{}, nil
	}

	// ---------- Service ----------
	svcName := deployment.Name + "-svc"
	var svc corev1.Service

	err := r.Get(ctx, client.ObjectKey{Name: svcName, Namespace: deployment.Namespace}, &svc)
	if errors.IsNotFound(err) {
		svc = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: deployment.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(&deployment, appsv1.SchemeGroupVersion.WithKind("Deployment")),
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: deployment.Spec.Template.Labels,
				Ports: []corev1.ServicePort{{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				}},
			},
		}

		if err := r.Create(ctx, &svc); err != nil {
			return ctrl.Result{}, err
		}
	}

	// ---------- Ingress ----------
	if deployment.Annotations["expose/type"] == "ingress" {
		ingName := deployment.Name + "-ing"
		var ing netv1.Ingress

		err = r.Get(ctx, client.ObjectKey{Name: ingName, Namespace: deployment.Namespace}, &ing)
		if errors.IsNotFound(err) {
			pathType := netv1.PathTypePrefix

			ing = netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ingName,
					Namespace: deployment.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(&deployment, appsv1.SchemeGroupVersion.WithKind("Deployment")),
					},
				},
				Spec: netv1.IngressSpec{
					Rules: []netv1.IngressRule{{
						Host: deployment.Name + ".local",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{{
									Path:     "/",
									PathType: &pathType,
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: svcName,
											Port: netv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								}},
							},
						},
					}},
				},
			}

			if err := r.Create(ctx, &ing); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = netv1.AddToScheme(scheme)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		os.Exit(1)
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(&ExposerReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}); err != nil {
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
