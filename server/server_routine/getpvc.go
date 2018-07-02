package server_routine

import (
	"fmt"
	"errors"
	ops "github.com/portworx/sched-ops/k8s"
	log "github.com/sirupsen/logrus"
)


func GetPvcPath() (map[string]string, error) {
	k8sops := ops.Instance()

	// get pod
	var podName = "nginx-px"
	pod, err := k8sops.GetPodByName(podName, "default")
	if err != nil {
		log.WithFields(log.Fields{podName: "Pod Not Found"}).Info(
			"Can not found Pod")
		return nil, errors.New("Can not found Pod")

	}

	//get pod's volumes
	volumes := pod.Spec.Volumes

	//get the list of volumes which are portworx volumes
	volumeNames := make(map[string]string)
	pvcMap := make(map[string]string)


	for _, v := range volumes {

		if v.PersistentVolumeClaim != nil {
			pvcName := v.PersistentVolumeClaim.ClaimName

			fmt.Printf("Found pvc: %s\n", pvcName)
			pvc, err:= k8sops.GetPersistentVolumeClaim(pvcName,"default")
			if err != nil {
				log.WithFields(log.Fields{pvcName: "PVC Not Found"}).Info(
					"Can not found PVC")
				return nil, errors.New("Can not found PVC")
			}

			fmt.Printf("Found pvc: %s\n", pvcName)
			provisioner, err:= k8sops.GetStorageProvisionerForPVC(pvc)
			if err != nil {
				log.WithFields(log.Fields{pvcName: "Provisioner Not Found"}).Info(
					"Can not found Provisioner")
				return nil, errors.New("Can not found Provisioner")
			}
			if provisioner == "kubernetes.io/portworx-volume" {
				pvcMap[v.Name] = pvcName
				volumeNames[v.Name] = v.Name
			}

		}
	}

	if len(pvcMap) == 0 {
		return nil, errors.New("Found no Portworx volumes")
	}

	// get the volume's path
	containers := pod.Spec.Containers
	for _, c := range containers {
		volMounts := c.VolumeMounts
		for _, vol := range volMounts {
			fmt.Printf("the volume mount is %s\n", vol.Name)
			if volumeNames[vol.Name] == vol.Name {
				pvc := pvcMap[vol.Name]
				pvcMap[pvc] = vol.MountPath
				delete(pvcMap, vol.Name)
				fmt.Printf("the voluem mount path is %v\n", vol.MountPath)
			}
		}
	}
	return pvcMap, nil
}