package controllers

import (
	"context"
	"testing"

	marin3rv1alpha1 "github.com/3scale/marin3r/apis/marin3r/v1alpha1"

	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func init() {
	s.AddKnownTypes(marin3rv1alpha1.GroupVersion,
		&marin3rv1alpha1.EnvoyConfigRevision{},
		&marin3rv1alpha1.EnvoyConfigRevisionList{},
		&marin3rv1alpha1.EnvoyConfig{},
	)
}

func TestReconcileSecret_Reconcile(t *testing.T) {
	t.Run("Sets ResourcesInSyncCondition to false in EnvoyConfigRevision resource when a referred secret changes", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "default",
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.key": []byte("xxxxx"),
				"tls.crt": []byte("xxxxx"),
			},
		}
		ecr := &marin3rv1alpha1.EnvoyConfigRevision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ecr",
				Namespace: "default",
			},
			Spec: marin3rv1alpha1.EnvoyConfigRevisionSpec{
				NodeID:  "node1",
				Version: "xxxx",
				EnvoyResources: &marin3rv1alpha1.EnvoyResources{
					Secrets: []marin3rv1alpha1.EnvoySecretResource{{
						Name: "secret",
						Ref: corev1.SecretReference{
							Name:      "secret",
							Namespace: "default",
						}}}},
			},
			Status: marin3rv1alpha1.EnvoyConfigRevisionStatus{
				Conditions: []status.Condition{{Type: marin3rv1alpha1.RevisionPublishedCondition, Status: corev1.ConditionTrue}},
			},
		}

		cl := fake.NewFakeClient(secret, ecr)
		r := &SecretReconciler{Client: cl, Scheme: s, Log: ctrl.Log.WithName("test")}

		_, gotErr := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "secret",
				Namespace: "default",
			},
		})

		if gotErr != nil {
			t.Errorf("TestReconcileSecret_Reconcile() returned error: '%v'", gotErr)
			return
		}

		r.Client.Get(context.TODO(), types.NamespacedName{Name: "ecr", Namespace: "default"}, ecr)
		if ecr.Status.Conditions.IsTrueFor(marin3rv1alpha1.ResourcesInSyncCondition) {
			t.Errorf("TestReconcileSecret_Reconcile() condition 'ResourcesInSyncCondition' was not set to false in EnvoyConfigRevision")
		}
	})
}
