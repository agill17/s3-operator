package controllers

import (
	"context"
	meta2 "k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FinalizerAction string

const (
	add    FinalizerAction = "add"
	remove FinalizerAction = "remove"
)

func SliceContainsString(slice []string, lookupString string) (bool, int) {

	if len(slice) == 0 {
		return false, -1
	}

	for idx, ele := range slice {
		if ele == lookupString {
			return true, idx
		}
	}

	return false, -1
}

func FinalizerOp(obj client.Object, client client.Client, action FinalizerAction, finalizer string) error {
	meta, err := meta2.Accessor(obj)
	if err != nil {
		return err
	}
	currentFinalizers := meta.GetFinalizers()
	exists, idx := SliceContainsString(currentFinalizers, finalizer)
	switch action {
	case add:
		if !exists {
			currentFinalizers = append(currentFinalizers, finalizer)
			meta.SetFinalizers(currentFinalizers)
			return client.Update(context.TODO(), obj)
		}
		break
	case remove:
		if exists {
			currentFinalizers = append(currentFinalizers[:idx], currentFinalizers[idx+1:]...)
			meta.SetFinalizers(currentFinalizers)
			return client.Update(context.TODO(), obj)
		}
	}
	return nil
}
