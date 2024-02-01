package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func main() {
	//kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	var restConfig *rest.Config
	var err error

	restConfig, err = clientcmd.BuildConfigFromFlags("", "")

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	/*gitRepo := "git://github.com/shabrul2451/apnader-service-backend.git"
	dockerfile := "Dockerfile"
	imageName := "shabrul2451/kaniko-test:v1"*/

	if err := createSecret(clientset); err != nil {
		fmt.Printf("Error creating secret: %v\n", err)
		os.Exit(1)
	}

	//if err := buildAndPushImage(clientset, gitRepo, dockerfile, imageName); err != nil {
	//	fmt.Printf("Error building and pushing image: %v\n", err)
	//	os.Exit(1)
	//}

	fmt.Println("Image build and push successfully!")
}

func buildAndPushImage(clientset *kubernetes.Clientset, gitRepo, dockerfilePath, imageName string) error {
	pod := buildKanikoPod(gitRepo, dockerfilePath, imageName)

	// Create Pod
	_, err := clientset.CoreV1().Pods("default").Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %v", err)
	}

	return nil
}

func createSecret(clientset *kubernetes.Clientset) error {
	secret := buildDockerSecret()
	_, err := clientset.CoreV1().Secrets("default").Create(context.TODO(), secret, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("failed to create pod: %v", err)
	}

	return nil
}

func buildDockerSecret() *v1.Secret {
	secretData := map[string]string{}
	secretData[".dockerconfigjson"] = `{"auths":{"https://index.docker.io/v1/":{"username":"shabrul2451","password":"bh0974316","email":"shabrul2451@gmail.com","auth":"c2hhYnJ1bDI0NTE6YmgwOTc0MzE2"}}}`
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "dockerCred",
		},
		StringData: secretData,
		Type:       "kubernetes.io/dockerconfigjson",
	}

	return secret
}

func buildKanikoPod(gitRepo, dockerfilePath, imageName string) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "kaniko-pod-",
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{{
				Name:  "kaniko",
				Image: "gcr.io/kaniko-project/executor:v1.6.0",
				Args: []string{
					"--dockerfile=" + dockerfilePath,
					"--destination=" + imageName,
					"--context=" + gitRepo,
				},
				VolumeMounts: []v1.VolumeMount{{
					Name:      "kaniko-vol",
					MountPath: "/workspace",
				}},
			}},
			Volumes: []v1.Volume{{
				Name: "kaniko-vol",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			}},
		},
	}

	return pod
}
