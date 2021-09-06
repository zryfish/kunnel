package app

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DeploymentTemplate = &v1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app": "kunnel",
		},
	},
	Spec: v1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "kunnel",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "kunnel",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "kunnel",
						Image: "jeffwithlove/kunnel:v0.1",
					},
				},
			},
		},
	},
}

func NewDeployment(namespace, service, localhost, server string, localport int, host string, headers []string) *v1.Deployment {
	deployment := DeploymentTemplate.DeepCopy()
	deployment.Name = fmt.Sprintf("kunnel-%s", service)
	deployment.Namespace = namespace

	command := []string{"client"}
	command = append(command, "--server", server, "--local", fmt.Sprintf("%s:%d", localhost, localport))
	if len(host) != 0 {
		command = append(command, "--host", host)
	}

	for _, header := range headers {
		command = append(command, "--header", header)
	}

	deployment.Spec.Template.Spec.Containers[0].Command = command

	return deployment
}
