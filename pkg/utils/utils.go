package utils

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateCr(runtimeObj runtime.Object, client client.Client) error {
	return client.Update(context.TODO(), runtimeObj)
}

func UpdateCrStatus(runtimeObj runtime.Object, client2 client.Client) error {
	return client2.Status().Update(context.TODO(), runtimeObj)
}

func AddFinalizer(whichFinalizer string, client client.Client, runtimeObj runtime.Object) error {
	meta, err := meta.Accessor(runtimeObj)
	if err != nil {
		return err
	}
	currentFinalizers := meta.GetFinalizers()
	finalizerExist, _ := sliceContainsString(currentFinalizers, whichFinalizer)
	if !finalizerExist {
		currentFinalizers := append(currentFinalizers, whichFinalizer)
		meta.SetFinalizers(currentFinalizers)
		return UpdateCr(runtimeObj, client)
	}
	return nil
}

func RemoveFinalizer(whichFinalizer string, runtimeOj runtime.Object, client client.Client) error{
	meta, err := meta.Accessor(runtimeOj)
	if err != nil {
		return err
	}
	currentFinalizers := meta.GetFinalizers()
	finalizerExist, idx := sliceContainsString(currentFinalizers, whichFinalizer)
	if finalizerExist {
		currentFinalizers := append(currentFinalizers[:idx], currentFinalizers[idx+1:]...)
		meta.SetFinalizers(currentFinalizers)
		return UpdateCr(runtimeOj, client)
	}
	return nil
}

func sliceContainsString(whichSlice []string, whichString string) (bool, int) {
	for i, e := range whichSlice {
		if e == whichString {
			return true, i
		}
	}
	return false, -1
}
