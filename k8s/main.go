package main

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"log"
	"path/filepath"
)

func main() {
	// Kubernetes API 地址
	apiServer := "https://120.25.224.159:6443"
	// 证书和密钥文件路径
	certFile := filepath.Join("/Users/liyanfeng/go/src/code/platform/k8s/client.crt")
	keyFile := filepath.Join("/Users/liyanfeng/go/src/code/platform/k8s/client.key")
	config := &rest.Config{
		Host: apiServer,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // 如果不使用 CA 证书，则设置为 true
			CertFile: certFile,
			KeyFile:  keyFile,
		},
	}
	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	result, err := clientset.AppsV1().Deployments("ones").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get deployments: %v", err)
	}
	for i := 0; i < len(result.Items); i++ {
		log.Printf("Deployment %d: %s\n", i, result.Items[i].Name)
	}
	var replicas int32 = 1
	a := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "afeng-plugin",
			Namespace: "ones",
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "afeng-plugin-1",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "afeng-plugin-1",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "afeng-plugin-host",
							Image:           "localhost:5000/ones/plugin-host-node:v6.0.36",
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-host-node-volume",
									MountPath: "/data/plugin/upload/GvY5xyKR/7T91t3pb/Gvmbef2y/1.2.106",
									SubPath:   "upload/GvY5xyKR/7T91t3pb/Gvmbef2y/1.2.106",
								},
								{
									Name:      "plugin-host-node-volume",
									MountPath: "/data/plugin/runtime/2ef83637",
									SubPath:   "runtime/2ef83637",
								},
							},
							Command: []string{
								"/usr/local/plugin-host-pkg/nodejs/v1.0/bin/host",
							},
							Args: []string{
								"--conf_path=/usr/local/plugin-host-pkg/nodejs/v1.0/config/config.yaml",
								"--host_id=Host-nodejs1822881926256529409",
								"--host_timeout_sec=30",
								"--platform_address=tcp://ones-platform-api-service:9009",
								"--plugin_path=/data/plugin/upload/GvY5xyKR/7T91t3pb/Gvmbef2y/1.2.106",
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
						{
							Name:            "afeng-plugin-standalonesvc",
							Image:           "localhost:5000/ones/plugin-host-node:v6.0.36",
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-host-node-volume",
									MountPath: "/data/plugin/upload/GvY5xyKR/7T91t3pb/Gvmbef2y/1.2.106",
									SubPath:   "upload/GvY5xyKR/7T91t3pb/Gvmbef2y/1.2.106",
								},
								{
									Name:      "plugin-host-node-volume",
									MountPath: "/data/plugin/6dU9NKJN",
									SubPath:   "6dU9NKJN",
								},
							},
							WorkingDir: "/data/plugin/upload/GvY5xyKR/7T91t3pb/Gvmbef2y/1.2.106/workspace",
							Command: []string{
								"/bin/bash",
								"-c",
							},
							Args: []string{
								"sh start.sh start --port=10000 --args= --mysql='plugin_built_in_gvmbef2y:sTCRPhUFMuJW1Re6@tcp(mysql-cluster-mysql-master:3306)/plugin_built_in_gvmbef2y?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true' --volume=/data/plugin/6dU9NKJN --secret=5Z2cN2VpT38PLVOn2aDyU92E8sqFAiDKXusoqq9kRGhGwjmWUU3uMd8qiGNwFe84 && echo 'Running some commands...' && sleep infinity",
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 10000,
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt32(10000),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      3,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "plugin-host-node-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "platform-plugin-pvc",
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = clientset.AppsV1().Deployments("ones").Create(context.Background(), &a, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create deployment: %v", err)
	}
}
