package controller

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	mysqlalpha1 "github.com/scorpinxia/mysql-operator/pkg/apis/mysql/v1alpha1"
	crclientset "github.com/scorpinxia/mysql-operator/pkg/clients/clientset/versioned"
	crinformer "github.com/scorpinxia/mysql-operator/pkg/clients/informers/externalversions/mysql/v1alpha1"

	"github.com/scorpinxia/mysql-operator/pkg/util"
)

type Controller struct {
	kubeclient kubernetes.Interface
	crClient   crclientset.Interface
	crSynced   cache.InformerSynced
}

func NewController(kubeClient kubernetes.Interface, crClient crclientset.Interface, crInformer crinformer.MySQLInformer) *Controller {
	controller := &Controller{
		kubeclient: kubeClient,
		crClient:   crClient,
		crSynced:   crInformer.Informer().HasSynced,
	}

	klog.InfoS("Set up event handlers.")
	crInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.add,
		UpdateFunc: controller.update,
		DeleteFunc: controller.delete,
	})

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	klog.InfoS("Run controller.")

	klog.InfoS("Wait for informer cache to sync.")
	if ok := cache.WaitForCacheSync(stopCh, c.crSynced); !ok {
		return errors.New("Failed to wait for caches to sync.")
	}

	klog.InfoS("Start worker.")
	<-stopCh
	klog.InfoS("Shut down.")

	return nil
}

func (c *Controller) add(obj interface{}) {
	klog.InfoS("Receive ADD Event.")
	c.applyObjectMySQL(obj)
}

func (c *Controller) update(old, new interface{}) {
	klog.InfoS("Receive UPDATE Event.")

	newMySQL := new.(*mysqlalpha1.MySQL)
	oldMySQL := old.(*mysqlalpha1.MySQL)
	if newMySQL.Spec.Version == oldMySQL.Spec.Version {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		return
	}
	c.applyObjectMySQL(new)
}

func (c *Controller) delete(obj interface{}) {
	klog.InfoS("Receive DELETE Event.")
	mysqlObj, ok := obj.(*mysqlalpha1.MySQL)
	if !ok {
		klog.Errorf("Failed parse to type object: %v", mysqlObj)
		return
	}

	util.DeleteStatefulSet(c.kubeclient, context.Background(), mysqlObj.Namespace, "mysql")
	util.DeleteSecret(c.kubeclient, context.Background(), mysqlObj.Namespace, "mysql-password")
	util.DeleteService(c.kubeclient, context.Background(), mysqlObj.Namespace, "mysql")
}

func (c *Controller) applyObjectMySQL(obj interface{}) {

	mysqlObj, ok := obj.(*mysqlalpha1.MySQL)
	if !ok {
		klog.Errorf("Failed to type assert object: %v", obj)
		return
	}
	klog.InfoS("obj", "namespace", mysqlObj.Namespace, "name", mysqlObj.Name)

	//创建对应命名空间
	err := util.CreateNamespaceOptional(c.kubeclient, context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: mysqlObj.Namespace},
	})
	if err != nil {
		klog.Errorf("Failed to create ns: %s", mysqlObj.Namespace)
	}

	ret := mysqlObj.DeepCopy()
	secretMysql := util.MakeSecretMysql(ret.Namespace)

	err = util.ApplySecret(c.kubeclient, context.Background(), secretMysql)
	if err != nil {
		klog.Errorf("Failed to apply secret: %v", secretMysql)
	}

	serviceMysql := util.MakeServiceMysql(ret.Namespace)
	err = util.ApplyService(c.kubeclient, context.Background(), serviceMysql)
	if err != nil {
		klog.Errorf("Failed to apply service: %v", serviceMysql)
	}

	statefulMysql := util.MakeStatefulSetMysql(ret.Namespace, ret.Spec.Version)
	util.ApplyStatefulSet(c.kubeclient, context.Background(), statefulMysql)
	if err != nil {
		klog.Errorf("Failed to apply statefulSet: %v", statefulMysql)
	}

	ret.Status.Message = "SUCCESS"
	_, err = c.crClient.ProductV1alpha1().MySQLs(ret.Namespace).UpdateStatus(context.TODO(), ret, metav1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to update status", "namespace", ret.Namespace, "name", ret.Name)
		return
	}
	klog.InfoS("Update Status.", "namespace", ret.Namespace, "name", ret.Name)
}
