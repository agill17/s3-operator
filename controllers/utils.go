package controllers

import (
	"context"
	"fmt"
	meta2 "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FinalizerAction string

const (
	Add    FinalizerAction = "add"
	Remove FinalizerAction = "remove"
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

func FinalizerOp(obj runtime.Object, client client.Client, action FinalizerAction, finalizer string) error {
	meta, err := meta2.Accessor(obj)
	if err != nil {
		return err
	}
	currentFinalizers := meta.GetFinalizers()
	exists, idx := SliceContainsString(currentFinalizers, finalizer)
	switch action {
	case Add:
		if !exists {
			currentFinalizers = append(currentFinalizers, finalizer)
			return client.Update(context.TODO(), obj)
		}
		break
	case Remove:
		if exists {
			currentFinalizers = append(currentFinalizers[:idx], currentFinalizers[idx+1:]...)
			meta.SetFinalizers(currentFinalizers)
			return client.Update(context.TODO(), obj)
		}
	}
	return nil
}

func FmtString(msg string) string {
	return fmt.Sprintf(msg)
}