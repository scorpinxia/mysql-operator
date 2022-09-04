package util

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

func ApplyStatefulSet(k kubernetes.Interface, ctx context.Context, ss *appsv1.StatefulSet) error {

	existed, err := k.AppsV1().StatefulSets(ss.Namespace).Get(ctx, ss.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = k.AppsV1().StatefulSets(ss.Namespace).Create(ctx, ss, metav1.CreateOptions{})
		return err
	}

	p := StrategicMergeFrom(existed)
	data, err := Merge.Data(ss)

	if err != nil {
		return err
	}

	_, err = k.AppsV1().StatefulSets(ss.Namespace).Patch(ctx, ss.Name, p.Type(), data, metav1.PatchOptions{})

	return err
}

func ApplySecret(k kubernetes.Interface, ctx context.Context, secret *corev1.Secret) error {
	existed, err := k.CoreV1().Secrets(secret.Namespace).Get(ctx, secret.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = k.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		return err
	}

	p := StrategicMergeFrom(existed)
	data, err := Merge.Data(secret)
	if err != nil {
		return err
	}

	_, err = k.CoreV1().Secrets(secret.Namespace).Patch(ctx, secret.Name, p.Type(), data, metav1.PatchOptions{})

	return err
}

func ApplyService(k kubernetes.Interface, ctx context.Context, service *corev1.Service) error {
	existed, err := k.CoreV1().Services(service.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = k.CoreV1().Services(service.Namespace).Create(ctx, service, metav1.CreateOptions{})
		return err
	}

	p := StrategicMergeFrom(existed)
	data, err := Merge.Data(service)
	if err != nil {
		return err
	}

	_, err = k.CoreV1().Services(service.Namespace).Patch(ctx, service.Name, p.Type(), data, metav1.PatchOptions{})

	return err
}

func DeleteStatefulSet(k kubernetes.Interface, ctx context.Context, namespace string, name string) error {
	err := k.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}

func DeleteService(k kubernetes.Interface, ctx context.Context, namespace string, name string) error {
	err := k.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}

func DeleteSecret(k kubernetes.Interface, ctx context.Context, namespace string, name string) error {
	err := k.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}

func CreateNamespaceOptional(k kubernetes.Interface, ctx context.Context, ns *corev1.Namespace) error {
	_, err := k.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func MakeStatefulSetMysql(namespace string, version string) *appsv1.StatefulSet {
	//API 组包法
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mysql",
				},
			},
			ServiceName: "mysql",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: pointer.Int64(10),
					Containers: []corev1.Container{
						{
							Name:  "mysql",
							Image: "mysql:" + version,
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3306,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mysql-store",
									MountPath: "/var/lib/mysql",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "MYSQL_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "mysql-password",
											},
											Key: "MYSQL_ROOT_PASSWORD",
										},
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysql-store",
						Namespace: namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}

	return ss
}

func MakeStatefulSetMysql2(namespace string, version string) *appsv1.StatefulSet {
	b, err := ioutil.ReadFile("/Users/bytedance/Downloads/workshopLab/lab4/mysql-operator/pkg/config/simple-ss.yaml")
	mysqlSS := &appsv1.StatefulSet{}
	mysqlJson, _ := yaml.ToJSON(b)
	if err = json.Unmarshal(mysqlJson, mysqlSS); err != nil {
		klog.InfoS("error Unmarshal SS")
		return nil
	}
	mysqlSS.Namespace = namespace
	//TODO 修改container
	return mysqlSS
}

func MakeSecretMysql(namespace string) *corev1.Secret {
	b, err := ioutil.ReadFile("/Users/bytedance/Downloads/workshopLab/lab4/mysql-operator/pkg/config/secret.yaml")
	mysqlSecret := &corev1.Secret{}
	mysqlJson, _ := yaml.ToJSON(b)
	if err = json.Unmarshal(mysqlJson, mysqlSecret); err != nil {
		klog.InfoS("error Unmarshal secret")
		return nil
	}
	mysqlSecret.Namespace = namespace
	return mysqlSecret
}

func MakeServiceMysql(namespace string) *corev1.Service {
	b, err := ioutil.ReadFile("/Users/bytedance/Downloads/workshopLab/lab4/mysql-operator/pkg/config/service.yaml")
	mysqlService := &corev1.Service{}
	mysqlJson, _ := yaml.ToJSON(b)
	if err = json.Unmarshal(mysqlJson, mysqlService); err != nil {
		klog.InfoS("error Unmarshal service")
		return nil
	}

	mysqlService.Namespace = namespace
	return mysqlService
}
