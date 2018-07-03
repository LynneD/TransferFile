package server_routine

import (
	"fmt"
	"errors"
	ops "github.com/portworx/sched-ops/k8s"
	log "github.com/sirupsen/logrus"
)


func GetPvcPath(deploymentName string) (map[string]string, error) {

	volumeNames := make(map[string]string)
	pvcMap := make(map[string]string)

	k8sops := ops.Instance()
	//fmt.Printf("the deployment is %s\n", deploymentName)
	// get deployment
	deployment, err := k8sops.GetDeployment(deploymentName, "default")

	if err != nil {
		log.WithFields(log.Fields{deploymentName: "Deployment doesn't found"}).Info(
			"Can not found Deployment")
		return nil, errors.New("Can not found Deployment")
	}

	// get pod
	pods, err := k8sops.GetDeploymentPods(deployment)

	if err != nil {
		log.WithFields(log.Fields{deploymentName: "No pod"}).Info(
			"Can't find the deploymnet's pods")
		return nil, errors.New("Can't find the deploymnet's pods")
	}



	// get pod

	volumes := pods[0].Spec.Volumes

	//get the list of volumes which are portworx volumes


	for _, v := range volumes {
		//fmt.Println("==============volumes ====================")

		//volumename := v.Name
		//fmt.Printf("teh volume is :%v\n", volumename)
		//fmt.Println("%v\n", v)
		if v.VolumeSource.PersistentVolumeClaim != nil {
			pvcName := v.VolumeSource.PersistentVolumeClaim.ClaimName

			//fmt.Printf("Found pvc: %s\n", pvcName)
			pvc, err:= k8sops.GetPersistentVolumeClaim(pvcName,"default")
			if err != nil {
				log.WithFields(log.Fields{pvcName: "PVC Not Found"}).Info(
					"Can not found PVC")
				return nil, errors.New("Can not found PVC")
			}

			//fmt.Printf("Found pvc: %s\n", pvcName)
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
	containers := pods[0].Spec.Containers
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